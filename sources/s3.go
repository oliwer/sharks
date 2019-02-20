package sources

import (
	"bytes"
	"io"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/AirVantage/sharks"
)

// To simplify things, we limit the number of keys we can get from S3.
var maxKeysPerBucket int64 = 1000

// ScanS3Bucket regularly reads `s3bucket` to populate the cache.
func ScanS3Bucket(cache *sharks.KeyCache, scanFrequency time.Duration, s3bucket, s3region string) {
	sess := session.Must(session.NewSession())
	sess.Config.Region = &s3region
	svc := s3.New(sess)

	// The bucket name may also include a path prefix for the objects.
	splitted := strings.SplitN(s3bucket, "/", 2)
	bucket := splitted[0]
	var prefix string
	if len(splitted) == 2 {
		prefix = splitted[1]
	}

	s3Get := func(key *string) ([]byte, error) {
		params := &s3.GetObjectInput{
			Bucket: &bucket,
			Key:    key,
		}
		s3resp, err := svc.GetObject(params)
		if err != nil {
			return nil, err
		}
		defer s3resp.Body.Close()
		var buf bytes.Buffer
		if _, err = io.Copy(&buf, s3resp.Body); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	query := &s3.ListObjectsV2Input{
		Bucket:  &bucket,
		MaxKeys: &maxKeysPerBucket,
		Prefix:  &prefix,
	}

	for {
		new := 0
		objectList, err := svc.ListObjectsV2(query)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				if aerr.Code() == s3.ErrCodeNoSuchBucket {
					log.Fatalln(aerr.Error())
				} else {
					log.Println(err)
				}
			} else {
				log.Println(err)
			}
		}

		for _, obj := range objectList.Contents {
			keybytes, err := s3Get(obj.Key)
			if err != nil {
				log.Println(err)
				continue
			}

			if cache.Upsert(keybytes, *obj.Key) {
				new++
			}
		}

		if new > 0 {
			log.Printf("found %d new keys\n", new)
		}

		time.Sleep(scanFrequency)
	}
}

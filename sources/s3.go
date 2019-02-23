package sources

import (
	"bytes"
	"io"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/AirVantage/sharks"
)

// Same limit as in aws sdk.
var defaultMaxKeysPerBucket int64 = 1000

type S3Bucket struct {
	Cache   *sharks.KeyCache
	Bucket  string
	Region  string
	MaxKeys int64

	session  *session.Session
	s3svc    *s3.S3
	prefix   string
	queryAll *s3.ListObjectsV2Input
}

func (bk *S3Bucket) Init() {
	if bk.MaxKeys == 0 {
		bk.MaxKeys = defaultMaxKeysPerBucket
	}

	bk.session = session.Must(session.NewSession())
	bk.session.Config.Region = &bk.Region
	bk.s3svc = s3.New(bk.session)

	// The bucket name may also include a path prefix for the objects.
	parts := strings.SplitN(bk.Bucket, "/", 2)
	bk.Bucket = parts[0]
	if len(parts) == 2 {
		bk.prefix = parts[1]
	}

	bk.queryAll = &s3.ListObjectsV2Input{
		Bucket:  &bk.Bucket,
		MaxKeys: &bk.MaxKeys,
		Prefix:  &bk.prefix,
	}

	// Test if the bucket is readable.
	_, err := bk.s3svc.GetBucketPolicy(&s3.GetBucketPolicyInput{Bucket: &bk.Bucket})
	if err != nil {
		log.Fatalln(err)
	}
}

func (bk *S3Bucket) getObject(key *string) ([]byte, error) {
	s3resp, err := bk.s3svc.GetObject(&s3.GetObjectInput{
		Bucket: &bk.Bucket, Key: key})
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, s3resp.Body)
	s3resp.Body.Close()
	return buf.Bytes(), err
}

func (bk *S3Bucket) Scan() int {
	new := 0

	objectList, err := bk.s3svc.ListObjectsV2(bk.queryAll)
	if err != nil {
		log.Println(err)
		return 0
	}

	if int64(len(objectList.Contents)) == bk.MaxKeys {
		log.Println("warning: the limit of objects per bucket has been reached:", bk.MaxKeys)
	}

	for _, obj := range objectList.Contents {
		keybytes, err := bk.getObject(obj.Key)
		if err != nil {
			log.Println(err)
			continue
		}

		if bk.Cache.Upsert(keybytes, *obj.Key) {
			new++
		}
	}

	return new
}

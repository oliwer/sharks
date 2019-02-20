SharKS - SSH Key Server
=======================

Sharks is a public key server for OpenSSH allowing you to centralize the authentication
process, in a safe and scalable way. No need to copy SSH keys on your servers anymore!

# How does it works?

1. Put all your public keys in a single directory.
2. Start SharKS. It will scan the directory and calculate the fingerprint of each key.
3. When you connect to your OpenSSH server, it will send an HTTPS request to SharKS with
   the fingerprint of the key used to authenticate.
4. If SharKS knows about this fingerprint, it allows the user login.

We rely on the `AuthorizedKeysCommand` feature introduced in OpenSSH 6.2.

# OpenSSH Setup

To configure OpenSSH to query SharKS when authenticating a user, you
must add the following lines to `/etc/ssh/sshd_config`:

    AuthorizedKeysCommand /usr/bin/curl -sG https://sharks.company.io/lookup --data-urlencode "fingerprint=%f"
    AuthorizedKeysCommandUser nobody


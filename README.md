# go-aws-s3get
A stupid simple S3 downloader CLI tool with supporting AWS Access Key, implemented by Golang.

[![Test](https://github.com/livesense-inc/go-aws-s3get/actions/workflows/test.yml/badge.svg)](https://github.com/livesense-inc/go-aws-s3get/actions/workflows/test.yml)


## Concept
Simple. No install. No dependencies.

## How to use

### Install

Download a binary and put it to PATH.

For example, on Linux.

```bash
sudo curl -L -o /usr/local/bin/s3get $(curl --silent "https://api.github.com/repos/livesense-inc/go-aws-s3get/releases/latest" | jq --arg PLATFORM_ARCH "$(echo `uname -s`-`uname -m` | tr A-Z a-z)" -r '.assets[] | select(.name | endswith($PLATFORM_ARCH)) | .browser_download_url')
sudo chmod 755 /usr/local/bin/s3get
```

A full list of binaries are [here](./releases/latest).


### Run

Run simply. It's like wget.

```bash
s3get -i AWS_ACCESS_KEY_ID -s AWS_SECRET_ACCESS_KEY -r AWS_REGION_NAME s3://bucket-name/path/to/file output-file
```

You can use environment variables to set AWS Access Key and AWS Region.

```bash
export AWS_ACCESS_KEY_ID=xxxx
export AWS_SECRET_ACCESS_KEY=xxxx
export AWS_REGION=ap-north-east1

s3get s3://bucket-name/path/to/file output-file
```

You can use profile name.

```bash
s3get -p yo -r ap-northeast-1 s3://bucket-name/path/to/file output-file
```

You can output to stdout.

```bash
s3get -p yo -r ap-northeast-1 s3://bucket-name/path/to/file - | md5
```


# Hack and Develop

First, fork this repo, and get your clone locally.

1. Install [go](http://golang.org)
2. Install `make`
3. Install [golangci-lint](https://golangci-lint.run/usage/install/#local-installation)

Write code and remove unused modules.

```
make tidy
```

To test, run

```
make lint
make test
```

To build, run

```
make build
```


## AUTHORS

* etsxxx

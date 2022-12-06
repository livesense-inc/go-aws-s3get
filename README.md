# go-aws-s3get
A stupid simple S3 downloader CLI tool with supporting AWS Access Key, implemented by Golang.

[![Test](https://github.com/livesense-inc/go-aws-s3get/actions/workflows/test.yml/badge.svg)](https://github.com/livesense-inc/go-aws-s3get/actions/workflows/test.yml)


## Concept
Simple. No setup. No dependencies.

## How to use

### Install

Download a binary and put it to PATH.

For example, on Linux.

```bash
sudo curl -L -o /usr/local/bin/s3get $(curl --silent "https://api.github.com/repos/livesense-inc/go-aws-s3get/releases/latest" | jq --arg PLATFORM_ARCH "$(echo `uname -s`-`uname -m` | tr A-Z a-z)" -r '.assets[] | select(.name | endswith($PLATFORM_ARCH)) | .browser_download_url')
sudo chmod 755 /usr/local/bin/s3get
```

A full list of binaries are [here](https://github.com/livesense-inc/go-aws-s3get/releases/latest).


### Run

If you already setup [AWS configuration and credential](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html), run simply. It's like wget.

```bash
s3get s3://bucket-name/path/to/file output-file
```

You can use a profile name.

```bash
s3get -p PROFILE_NAME s3://bucket-name/path/to/file output-file
```

It can be run without an AWS profile, just with command line options.

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

Direct output to stdout is possible.

```bash
s3get s3://bucket-name/path/to/file - | md5
```


## Hack and Develop

### Build

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

### Integration test with MinIO

[MinIO](https://min.io/), an S3-compatible object storage, allows integration testing with Amazon S3 in local.

1. Install MinIO (see official document)
2. Run MinIO. Example is following.

```bash
MINIO_ROOT_USER=testid MINIO_ROOT_PASSWORD=testsecret minio server /tmp/minio
```

3. Upload test file with aws-cli

```bash
export AWS_ACCESS_KEY_ID=testid
export AWS_SECRET_ACCESS_KEY=testsecret
export AWS_DEFAULT_REGION=test

### create bucket
aws --endpoint-url http://127.0.0.1:9000 s3 mb s3://test

### upload test file
echo miniotest > /tmp/test.txt
aws --endpoint-url=http://127.0.0.1:9000 s3 cp /tmp/test.txt s3://test/test.txt
```

4. Run `s3get` with `--endpoint-url` option.

```bash
./bin/s3get --endpoint-url=http://127.0.0.1:9000 -i testid -s testsecret -r test s3://test/test.txt -
```

## AUTHORS

* [etsxxx](https://github.com/etsxxx)

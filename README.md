# go-aws-s3get
A stupid simple S3 downloader CLI tool with supporting AWS Access Key, implemented by Golang.

## Concept
Simple. No install. No dependencies.

## How to use

### Install

```bash
curl -o /usr/local/bin/s3get TBA
```

### Run

```bash
s3get -i AWS_ACCESS_KEY_ID -s AWS_SECRET_ACCESS_KEY -r AWS_REGION_NAME s3://bucket-name/path/to/file output-file
```

You can use environment variables to set AWS Access Key.

```bash
export AWS_ACCESS_KEY_ID=xxxx
export AWS_SECRET_ACCESS_KEY=xxxx
export AWS_DEFAULT_REGION=ap-north-east1

s3get s3://bucket-name/path/to/file output-file
```


## AUTHORS

* etsxxx

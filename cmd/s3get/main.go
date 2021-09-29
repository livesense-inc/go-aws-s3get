package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/urfave/cli/v2"
)

var version, gitcommit string

const incompleteFileSuffix = ".incomplete"

func argumentError() error {
	exe, _ := os.Executable()
	return fmt.Errorf("argument error. see '%s -h'", path.Base(exe))
}

func getMD5(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed open file : %s", err)
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed calc md5 : %s", err)
	}

	hashInBytes := h.Sum(nil)[:16]
	return hex.EncodeToString(hashInBytes), nil
}

func splitS3Path(path string) (bucketName string, key string, err error) {
	if !strings.HasPrefix(path, "s3://") {
		return "", "", fmt.Errorf("s3 path is invalid format")
	}

	sub := strings.SplitN(path[len("s3://"):], "/", 2)
	bucketName = sub[0]
	key = "/" + sub[1]
	return
}

func download(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 || 2 < ctx.Args().Len() {
		return argumentError()
	}

	// parse 1st arg
	src := ctx.Args().Get(0)
	bucketName, key, err := splitS3Path(src)
	if err != nil {
		return argumentError()
	}
	if bucketName == "" || key == "/" {
		return argumentError()
	}

	// parse 2nd arg
	dest := path.Base(src)
	if ctx.Args().Len() == 2 {
		arg := ctx.Args().Get(1)
		if s, err := os.Stat(arg); err == nil && s.IsDir() {
			// if directory path is given, output to directory
			dest = strings.TrimRight(ctx.Args().Get(1), "/") + "/" + path.Base(src)
		} else {
			// output to arg
			dest = arg
		}
	}

	awsAccessKeyID := ctx.String("id")
	awsSecretAccessKey := ctx.String("secret")
	awsRegion := ctx.String("region")

	fmt.Printf("Download file from region=%s, bucket=%s, key=%s\n", awsRegion, bucketName, key)

	// check outfile
	hash := ""
	if s, err := os.Stat(dest); err == nil && !s.IsDir() {
		if hash, err = getMD5(dest); err != nil {
			return fmt.Errorf("failed to open file %q : %v", dest, err)
		}
	}

	// create tmpfile
	tmpfile := dest + incompleteFileSuffix
	f, err := os.Create(tmpfile)
	if err != nil {
		return fmt.Errorf("failed to create file %s : %v", tmpfile, err)
	}
	defer func() {
		f.Close()
		if _, err := f.Stat(); err != nil {
			os.Remove(f.Name())
		}
	}()

	creds := credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, "")
	sess := session.Must(
		session.NewSession(
			&aws.Config{
				Credentials: creds,
				Region:      aws.String(awsRegion),
			},
		),
	)

	downloader := s3manager.NewDownloader(sess)
	size, err := downloader.Download(
		f, &s3.GetObjectInput{
			Bucket:      aws.String(bucketName),
			Key:         aws.String(key),
			IfNoneMatch: aws.String(hash),
		},
	)
	if err != nil {
		if reqErr, ok := err.(awserr.RequestFailure); ok {
			if reqErr.StatusCode() == 304 {
				fmt.Printf("file is not modified: %s\n", dest)
				return nil
			}
			return fmt.Errorf("failed to download : %v", reqErr.Message())
		}
		return fmt.Errorf("failed to download : %v", err)
	}

	// rename tmpfile to dest
	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close %s : %v", f.Name(), err)
	}
	if err := os.Rename(f.Name(), dest); err != nil {
		return fmt.Errorf("failed to rename %s to %s : %v", f.Name(), dest, err)
	}

	fmt.Printf("successfuly downloaded to %s (%d bytes)\n", dest, size)
	return nil
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "id",
				Aliases:     []string{"i"},
				Usage:       "AWS_ACCESS_KEY_ID, recommend to use env",
				EnvVars:     []string{"AWS_ACCESS_KEY_ID"},
				DefaultText: "secret",
			},
			&cli.StringFlag{
				Name:        "secret",
				Aliases:     []string{"s"},
				Usage:       "AWS_SECRET_ACCESS_KEY, recommend to use env",
				EnvVars:     []string{"AWS_SECRET_ACCESS_KEY"},
				DefaultText: "secret",
			},
			&cli.StringFlag{
				Name:    "region",
				Aliases: []string{"r"},
				Usage:   "AWS region name",
				EnvVars: []string{"AWS_DEFAULT_REGION"},
			},
		},
		Action: download,
	}

	app.Name = "s3get"
	app.Usage = "A stupid simple downloader for AWS S3"
	app.Version = fmt.Sprintf("%s (rev:%s)", version, gitcommit)
	app.ArgsUsage = "s3://bucket-name/path/to/file output-file"

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

func splitS3Path(str string) (bucketName string, key string, err error) {
	if !strings.HasPrefix(str, "s3://") {
		return "", "", fmt.Errorf("s3 path is invalid format")
	}

	sub := strings.SplitN(str[len("s3://"):], "/", 2)
	if len(sub) != 2 {
		return "", "", fmt.Errorf("s3 path is invalid format")
	}
	bucketName = sub[0]
	key = sub[1]
	if bucketName == "" {
		return "", "", fmt.Errorf("s3 path is invalid format")
	}
	return
}

func getOutputPath(src string, dest string) (outputPath string, err error) {
	if src == "" {
		return "", fmt.Errorf("src is invalid")
	}
	outputPath = path.Base(src)
	if dest != "" {
		if dest == "-" {
			// stdout
			outputPath = ""
		} else if s, err := os.Stat(dest); err == nil && s.IsDir() {
			// if directory path is given, output to directory
			outputPath = strings.TrimRight(dest, "/") + "/" + path.Base(src)
		} else {
			// output to dest
			outputPath = dest
		}
	}
	return outputPath, nil
}

func writeToFile(res *s3.GetObjectOutput, dest string) (bytes int, err error) {
	// create tmpfile
	tmpfile := dest + incompleteFileSuffix
	f, err := os.Create(tmpfile)
	if err != nil {
		return 0, fmt.Errorf("failed to create file %s : %v", tmpfile, err)
	}
	defer func() {
		f.Close()
		if _, err := f.Stat(); err != nil {
			os.Remove(f.Name())
		}
	}()

	// write to file
	bytes = 0
	buf := make([]byte, 4096)
	for {
		n, err := res.Body.Read(buf)
		if err != nil && err != io.EOF {
			return bytes, err
		}
		if n == 0 {
			break
		}
		if _, err := f.Write(buf[:n]); err != nil {
			return bytes, err
		}
		bytes += n
	}

	// rename tmpfile to dest
	if err := f.Close(); err != nil {
		return bytes, fmt.Errorf("failed to close %s : %v", f.Name(), err)
	}
	if err := os.Rename(f.Name(), dest); err != nil {
		return bytes, fmt.Errorf("failed to rename %s to %s : %v", f.Name(), dest, err)
	}

	return bytes, nil
}

func writeToStdout(res *s3.GetObjectOutput) (bytes int, err error) {
	bytes = 0
	buf := make([]byte, 4096)
	for {
		n, err := res.Body.Read(buf)
		if err != nil && err != io.EOF {
			return bytes, err
		}
		if n == 0 {
			break
		}
		if _, err := os.Stdout.Write(buf[:n]); err != nil {
			return bytes, err
		}
		bytes += n
	}
	return bytes, nil
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
	dest, err := getOutputPath(src, ctx.Args().Get(1))
	if err != nil {
		return argumentError()
	}
	outputToStdout := (dest == "")

	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithSharedConfigProfile(ctx.String("profile")),
	)
	if err != nil {
		return err
	}

	// read AWS Access Keys and overwrite credential
	awsAccessKeyID := ctx.String("id")
	awsSecretAccessKey := ctx.String("secret")
	awsSessionToken := os.Getenv("AWS_SESSION_TOKEN")
	if awsAccessKeyID != "" && awsSecretAccessKey != "" {
		cfg.Credentials = credentials.NewStaticCredentialsProvider(awsAccessKeyID, awsSecretAccessKey, awsSessionToken)
	}

	// if region is specified, overwrite
	awsRegion := ctx.String("region")
	if awsRegion != "" {
		cfg.Region = *aws.String(awsRegion)
	}

	// check outfile
	hash := ""
	if !outputToStdout {
		if s, err := os.Stat(dest); err == nil && !s.IsDir() {
			if hash, err = getMD5(dest); err != nil {
				return fmt.Errorf("failed to open file %q : %v", dest, err)
			}
		}
	}

	if !outputToStdout {
		fmt.Printf("Download file from region=%s, bucket=%s, key=%s\n", awsRegion, bucketName, key)
	}

	// Create s3 client and download object
	s3cli := s3.NewFromConfig(cfg)
	res, err := s3cli.GetObject(
		context.TODO(),
		&s3.GetObjectInput{
			Bucket:      aws.String(bucketName),
			Key:         aws.String(key),
			IfNoneMatch: aws.String(hash),
		},
	)
	if err != nil {
		var respErr *awshttp.ResponseError
		if errors.As(err, &respErr) && respErr.HTTPStatusCode() == 304 {
			fmt.Printf("file is not modified: %s\n", dest)
			return nil
		}
		return err
	}

	if outputToStdout {
		bytes, err := writeToStdout(res)
		if err != nil {
			return nil
		}
		os.Stderr.WriteString(fmt.Sprintf("successfuly output to stdout (%d bytes)\n", bytes))
	} else {
		bytes, err := writeToFile(res, dest)
		if err != nil {
			return nil
		}
		fmt.Printf("successfuly downloaded to %s (%d bytes)\n", dest, bytes)
	}
	return nil
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "region",
				Aliases:  []string{"r"},
				Usage:    "AWS region name",
				EnvVars:  []string{"AWS_DEFAULT_REGION", "AWS_REGION"},
				Required: true,
			},
			&cli.StringFlag{
				Name:    "profile",
				Aliases: []string{"p"},
				Usage:   "AWS_PROFILE, profile name to use",
				EnvVars: []string{"AWS_PROFILE"},
				Value:   "default",
			},
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
		},
		HideHelpCommand: true,
		Action:          download,
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

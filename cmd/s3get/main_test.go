package main

import (
	"testing"
)

func Test_splitS3Path(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name           string
		args           args
		wantBucketName string
		wantKey        string
		wantErr        bool
	}{
		{
			name:           "str is valid s3 path format, return valid response",
			args:           args{str: "s3://yo/hey"},
			wantBucketName: "yo",
			wantKey:        "/hey",
			wantErr:        false,
		},
		{
			name:           "str is valid s3 path '/', return valid response",
			args:           args{str: "s3://yo/"},
			wantBucketName: "yo",
			wantKey:        "/",
			wantErr:        false,
		},
		{
			name:           "str is valid s3 path '/hey/hey', return valid response",
			args:           args{str: "s3://yo/hey/hey"},
			wantBucketName: "yo",
			wantKey:        "/hey/hey",
			wantErr:        false,
		},
		{
			name:           "str is blank, return error",
			args:           args{str: ""},
			wantBucketName: "",
			wantKey:        "",
			wantErr:        true,
		},
		{
			name:           "str is typo, return error",
			args:           args{str: "s3:/yo/hey"},
			wantBucketName: "",
			wantKey:        "",
			wantErr:        true,
		},
		{
			name:           "str is no bucket name, return error",
			args:           args{str: "s3:///hey"},
			wantBucketName: "",
			wantKey:        "",
			wantErr:        true,
		},
		{
			name:           "str is no object key, return error",
			args:           args{str: "s3://yo"},
			wantBucketName: "",
			wantKey:        "",
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBucketName, gotKey, err := splitS3Path(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitS3Path() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotBucketName != tt.wantBucketName {
				t.Errorf("splitS3Path() gotBucketName = %v, want %v", gotBucketName, tt.wantBucketName)
			}
			if gotKey != tt.wantKey {
				t.Errorf("splitS3Path() gotKey = %v, want %v", gotKey, tt.wantKey)
			}
		})
	}
}

func Test_getOutputPath(t *testing.T) {
	type args struct {
		src  string
		dest string
	}
	tests := []struct {
		name           string
		args           args
		wantOutputPath string
		wantErr        bool
	}{
		{
			name:           "src is invalid, return error",
			args:           args{src: "", dest: ""},
			wantOutputPath: "",
			wantErr:        true,
		},
		{
			name:           "src is valid and dest is blank, return base name of src",
			args:           args{src: "s3://yo/hey", dest: ""},
			wantOutputPath: "hey",
			wantErr:        false,
		},
		{
			name:           "src is valid and dest is blank file path, return file path",
			args:           args{src: "s3://yo/hey", dest: "/tmp/outputpath"},
			wantOutputPath: "/tmp/outputpath",
			wantErr:        false,
		},
		{
			name:           "src is valid and dest is directory, return directory + base name of src",
			args:           args{src: "s3://yo/hey", dest: "/tmp"},
			wantOutputPath: "/tmp/hey",
			wantErr:        false,
		},
		{
			name:           "src is valid and dest is directory with /, return directory + base name of src",
			args:           args{src: "s3://yo/hey", dest: "/tmp/"},
			wantOutputPath: "/tmp/hey",
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOutputPath, err := getOutputPath(tt.args.src, tt.args.dest)
			if (err != nil) != tt.wantErr {
				t.Errorf("getOutputPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOutputPath != tt.wantOutputPath {
				t.Errorf("getOutputPath() = %v, want %v", gotOutputPath, tt.wantOutputPath)
			}
		})
	}
}

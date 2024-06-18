package main

import (
	"os"
	"testing"
)

func Test_fileProcess(t *testing.T) {
	type args struct {
		replaceURLPrefix string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test",
			args: args{
				replaceURLPrefix: "https://test/",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Chdir(os.Getenv("WORKSPACE_DIR"))
			if err := fileProcess("sg", tt.args.replaceURLPrefix, "output/sg"); (err != nil) != tt.wantErr {
				t.Errorf("fileProcess() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

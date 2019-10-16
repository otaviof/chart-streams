package chartstreams

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestPackage_Add(t *testing.T) {
	type fields struct {
		content bytes.Buffer
	}
	type args struct {
		path string
		info os.FileInfo
		r    io.Reader
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Package{
				content: tt.fields.content,
			}
			if err := p.Add(tt.args.path, tt.args.info, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("Package.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

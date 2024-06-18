#! /bin/bash

set -ex

mkdir -p bin 
go build -o ./bin/build_and_upload_tool ./upload_cdn_tool
chmod +x ./bin/build_and_upload_tool

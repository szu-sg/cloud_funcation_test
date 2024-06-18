#!/bin/sh
set -ex

./bin/build_and_upload_tool $@

# scm 编译需要以下路径全部存在，否则 cdn 上传会失败，找不到对应路径
mkdir -p output/cn && echo " " > output/cn/.placeholder
mkdir -p output/va && echo " " > output/va/.placeholder
mkdir -p output/sg && echo " " > output/sg/.placeholder


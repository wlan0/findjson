#!/bin/bash

set -e

src_dir=$(dirname $0)

#build from parent dir
cd $src_dir

CGO_ENABLED=0 go build -o findjson github.com/wlan0/findjson/cmd 

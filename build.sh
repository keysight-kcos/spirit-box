#!/bin/bash

set -ex

/usr/local/go/bin/go build -o spirit-box main.go
mv spirit-box /usr/bin/spirit-box

CONFIG_DIR=/usr/share/spirit-box
if [ ! -d "$CONFIG_DIR" ]; then
	mkdir "$CONFIG_DIR"
fi

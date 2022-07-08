#!/bin/bash

set -ex

/usr/local/go/bin/go build -o spirit-box main.go
mv spirit-box /usr/bin/spirit-box

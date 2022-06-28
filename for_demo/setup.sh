#!/bin/bash

set -ex

echo "$(pwd)"/example_scripts/script2.sh > scripts

cp ./scripts /usr/share/spirit-box/
cp ./whitelist /usr/share/spirit-box/whitelist
cp $(echo "$(pwd)"/example_scripts/script1.sh) /usr/share/spirit-box/

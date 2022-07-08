#!/bin/bash

set -ex
#debug=echo

DEBPATH=./packages/spirit-box_1.0.0_arm64
CONFPATH=/lib/systemd/system
allPaths=(/spirit-box@.service /disableSystemdLogging.service /getty@.service.d/wait-for-spirit-box.conf /serial-getty@.service.d/wait-for-spirit-box.conf)
idirs=(/getty@.service.d /serial-getty@.service.d)  # intermediate directories

$debug cp /usr/bin/spirit-box "$DEBPATH"/usr/bin/
$debug rm -rf ${DEBPATH}${CONFPATH}/*

for d in ${idirs[@]}; do # make intermediate directories if they don't exist
	[ ! -d "${DEBPATH}${CONFPATH}${d}" ] && mkdir "${DEBPATH}${CONFPATH}${d}" && echo Created "${DEBPATH}${CONFPATH}${d}"
done

for t in ${allPaths[@]}; do
	$debug cp "${CONFPATH}${t}" "${DEBPATH}${CONFPATH}${t}"
done

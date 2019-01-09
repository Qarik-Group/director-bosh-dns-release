#!/bin/bash -x
export GOOS=linux
export GOARCH=amd64

go build -o bosh-dns-local-fsnotify main.go

ext_path=/var/vcap/jobs/bosh-dns-local-fsnotify/bin

set -x
bucc ssh "sudo mkdir -p ${ext_path}"
tar -c bosh-dns-local-fsnotify | bucc ssh "sudo tar -v -x --no-same-owner -C ${ext_path}"

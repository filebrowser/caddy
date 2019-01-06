#!/bin/sh

set -e

go get -v ./...
cd /go/src/github.com/mholt/caddy/caddy
go get -v ./...
go get -v github.com/caddyserver/builds
sed -i.bak 's#// This is where other plugins get plugged in (imported)#_ "github.com/filebrowser/caddy"#' caddymain/run.go
go run build.go
ls -la

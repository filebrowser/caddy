#!/bin/sh

set -e

cd ..
rm -rf filebrowser
cd caddy
go get -v ./...
cd ../../mholt/caddy/caddy
go get -v ./...
go get -v github.com/caddyserver/builds
sed -i.bak 's#// This is where other plugins get plugged in (imported)#_ "github.com/filebrowser/caddy/filemanager"\n_ "github.com/filebrowser/caddy/hugo"\n_ "github.com/filebrowser/caddy/jekyll"#' caddymain/run.go
go run build.go

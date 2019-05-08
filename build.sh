#!/bin/sh

set -euo pipefail

git clone https://github.com/mholt/caddy
cd caddy/caddy
sed -i.bak 's#// This is where other plugins get plugged in (imported)#_ "github.com/filebrowser/caddy"#' caddymain/run.go
go get
go run build.go
ls -la

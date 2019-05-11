#!/bin/sh

set -euo pipefail

commit=$(git rev-parse --short HEAD)
mkdir -p build/caddy
cd build/caddy

cat >main.go <<EOL
package main

import (
	"github.com/mholt/caddy/caddy/caddymain"

	// plug in plugins here, for example:
	_ "github.com/filebrowser/caddy"
)

func main() {
	// optional: disable telemetry
	// caddymain.EnableTelemetry = false
	caddymain.Run()
}
EOL

cat >go.mod <<EOL
module caddy

EOL

go get "github.com/filebrowser/caddy@$commit"
go get
go build main.go

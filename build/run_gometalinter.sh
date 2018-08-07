#!/bin/sh

set -e

cd $(dirname $0)/..

docker run --rm -itv $(pwd):/go/src/github.com/filebrowser/caddy filebrowser/dev sh -c "\
  cd .. && rm -rf filebrowser && cd caddy && \
  CGO_ENABLED=0 gometalinter --exclude='rice-box.go' --deadline=300s"

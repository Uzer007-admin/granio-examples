#!/bin/sh
set -eu
cd "$(dirname "$0")"
module_dir=$(go list -m -f '{{.Dir}}' github.com/Uzer007-admin/granio/v2)
granio bundle -o metadata_bundled.go -package data -name resourceAuthors "$module_dir/AUTHORS"

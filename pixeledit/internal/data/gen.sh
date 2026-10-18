#!/bin/sh
set -eu
cd "$(dirname "$0")"
granio bundle -package data -name pencil -o bundled.go pencil.svg
granio bundle -package data -append -name dropper -o bundled.go dropper.svg

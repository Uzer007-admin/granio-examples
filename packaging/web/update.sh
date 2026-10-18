#!/bin/sh
set -eu
script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_dir=$(CDPATH= cd -- "$script_dir/../.." && pwd)
ref=${1:-HEAD}
commit=$(git -C "$repo_dir" rev-parse --verify --end-of-options "${ref}^{commit}")
work_dir=$(mktemp -d "${TMPDIR:-/tmp}/granio-examples-web.XXXXXX")
trap 'rm -rf -- "$work_dir"' EXIT HUP INT TERM
mkdir "$work_dir/source" "$work_dir/bin"
git -C "$repo_dir" archive --output "$work_dir/source.tar" "$commit"
tar -xf "$work_dir/source.tar" -C "$work_dir/source"
bin_dir="$work_dir/bin"
if command -v cygpath >/dev/null 2>&1; then
    bin_dir=$(cygpath -w "$bin_dir")
fi
GOWORK=off GOBIN="$bin_dir" go install github.com/Uzer007-admin/granio-tools/cmd/granio@v1.7.4-0.20261018072944-90863175404b
(cd "$work_dir/source/demo" && GOWORK=off "$work_dir/bin/granio" package --release --target wasm --name "Granio Demo")
mkdir -p "$script_dir/site"
cp -R "$work_dir/source/demo/wasm/." "$script_dir/site/"
printf '{"source":"https://github.com/Uzer007-admin/granio-examples","commit":"%s","path":"demo","tool_version":"v1.7.4-0.20261018072944-90863175404b"}\n' "$commit" > "$script_dir/site/SOURCE.json"
printf 'Готово: %s/site\n' "$script_dir"

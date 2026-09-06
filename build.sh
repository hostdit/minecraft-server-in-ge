#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")"

if [ "${1:-host}" = "all" ]; then
	mkdir -p dist
	for t in "darwin arm64" "darwin amd64" "windows amd64" "linux amd64"; do
		set -- $t
		out="dist/gemc-$1-$2"
		[ "$1" = "windows" ] && out="$out.exe"
		GOOS=$1 GOARCH=$2 CGO_ENABLED=0 go build -trimpath -o "$out" .
		echo "built $out"
	done
	exit 0
fi

out="gemc"
case "$(go env GOOS)" in windows) out="gemc.exe" ;; esac
CGO_ENABLED=0 go build -trimpath -o "$out" .
echo "built ./$out"
echo "run it with: ./$out"

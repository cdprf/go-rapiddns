#!/bin/bash
set -e
mkdir -p build
PLATFORMS=(
  "windows amd64"
  "windows 386"
  "linux amd64"
  "linux 386"
  "linux arm64"
  "linux arm"
  "darwin amd64"
  "darwin arm64"
)
for plat in "${PLATFORMS[@]}"; do
  set -- $plat
  GOOS=$1
  GOARCH=$2
  output_name=build/rapiddnsquery-$GOOS-$GOARCH
  if [ $GOOS = "windows" ]; then output_name+='.exe'; fi
  echo "Building $output_name ..."
  env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name main.go
  echo "OK: $output_name"
done

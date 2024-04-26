#! /bin/bash

set -uexo pipefail

echo "For local development use only."
echo "This approach is a bit brittle so be sure to understand the expected setup."

remote_url=$(git config --get remote.origin.url)
bare_remote=${remote_url/.git/}
bare_remote=${bare_remote/git@github.com:/}
short_sha=$(git rev-parse --short HEAD)
tag="sha-${short_sha}"
full_container_tag="ghcr.io/${bare_remote}:${tag}"

go build -o build/ubuntu-reportd ./cmd/ubuntu-reportd
docker build -f docker/ubuntu-reportd/Dockerfile -t "$full_container_tag" .
docker push "$full_container_tag"

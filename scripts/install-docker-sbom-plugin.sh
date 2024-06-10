#!/bin/bash

set -euxo pipefail

version=${DOCKER_SBOM_PLUGIN_VERSION:-latest}
os=${OS:-linux}
arch=${ARCH:-amd64}
target=$TMPDIR/sbom-cli-plugin.tgz
plugin_dir="$HOME/.docker/cli-plugins/"

download_url_base=https://github.com/docker/sbom-cli-plugin/releases/download
url="${download_url_base}/v${version}/sbom-cli-plugin_${version}_${os}_${arch}.tar.gz"

curl -L "${url}" -o "${target}"
pushd "${TMPDIR}"
tar zxvf "${target}" docker-sbom
chmod +x docker-sbom
popd
cp "${TMPDIR}/docker-sbom" "${plugin_dir}"

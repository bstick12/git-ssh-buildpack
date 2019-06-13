#!/usr/bin/env bash
set -euo pipefail

TARGET_OS=${1:-linux}

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

echo "Target OS is $TARGET_OS"
echo -n "Creating buildpack directory..."
bp_dir="${PWD##*/}_built"
rm -rf $bp_dir
mkdir $bp_dir
echo "done"

echo -n "Copying buildpack.toml..."
cp buildpack.toml $bp_dir/buildpack.toml
echo "done"

for b in $(ls cmd); do
    echo -n "Building $b..."
    GOOS=$TARGET_OS go build -o $bp_dir/bin/$b ./cmd/$b
    echo "done"
done

fullPath=$(realpath "$bp_dir")
echo "Buildpack packaged into: $fullPath"

buildpack_name="$(basename `pwd`)"

pushd $bp_dir
    tar czvf "../$buildpack_name.tgz" *
popd

#!/bin/bash

set -e

# generate compiles the *.pb.go stubs from the *.proto files.
function generate() {
  echo "Generating root gRPC server protos"

  PROTOS="tapmonrpc.proto"

  # For each of the sub-servers, we then generate their protos, but a restricted
  # set as they don't yet require REST proxies, or swagger docs.
  for file in $PROTOS; do
    DIRECTORY=$(dirname "${file}")
    echo "Generating protos from ${file}, into ${DIRECTORY}"

    # Generate the protos.
    protoc -I/usr/local/include -I. \
      --go_out . --go_opt paths=source_relative \
      --go-grpc_out . --go-grpc_opt paths=source_relative \
      "${file}"

  done

}

# format formats the *.proto files with the clang-format utility.
function format() {
  find . -name "*.proto" -print0 | xargs -0 clang-format --style=file -i
}

# Compile and format the tapmonrpc package.
pushd tapmonrpc
format
generate
popd

if [[ "$COMPILE_MOBILE" == "1" ]]; then
  pushd mobile
  ./gen_bindings.sh $FALAFEL_VERSION
  popd
fi

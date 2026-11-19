#!/bin/bash
# Regenerates internal/gomesh/gomeshproto and internal/gomesh/nanopb from the
# Meshtastic protobuf definitions (https://github.com/meshtastic/protobufs).
#
# Requires protoc and protoc-gen-go on the PATH, for example:
#   nix-shell -p protobuf protoc-gen-go --run ./genproto.sh
#
# PROTOBUFS_TAG selects the upstream tag to clone (default below).
# PROTOBUFS_DIR points at an existing checkout instead of cloning.
set -euo pipefail

PROTOBUFS_TAG="${PROTOBUFS_TAG:-v2.8.3}"
MODULE="github.com/lmatte7/meshtastic-go"
PROTO_PKG="$MODULE/internal/gomesh/gomeshproto"
NANOPB_PKG="$MODULE/internal/gomesh/nanopb"

cd "$(dirname "$0")"

SRC="${PROTOBUFS_DIR:-}"
if [ -z "$SRC" ]; then
	TMP="$(mktemp -d)"
	trap 'rm -rf "$TMP"' EXIT
	git clone -q --depth 1 --branch "$PROTOBUFS_TAG" https://github.com/meshtastic/protobufs "$TMP/protobufs"
	SRC="$TMP/protobufs"
fi

# The upstream go_package option points at a package that does not exist, so
# map every file onto our own packages (import path;package name).
OPTS=(--go_opt="module=$MODULE" --go_opt="Mnanopb.proto=$NANOPB_PKG;nanopb")
for f in "$SRC"/meshtastic/*.proto; do
	OPTS+=(--go_opt="Mmeshtastic/$(basename "$f")=$PROTO_PKG;gomeshproto")
done

rm -rf internal/gomesh/gomeshproto internal/gomesh/nanopb
protoc -I "$SRC" --go_out=. "${OPTS[@]}" "$SRC/nanopb.proto" "$SRC"/meshtastic/*.proto

echo "generated internal/gomesh/gomeshproto from meshtastic/protobufs $PROTOBUFS_TAG"

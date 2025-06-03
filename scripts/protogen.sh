#!/usr/bin/env bash

SRC_DIR=./protobuf
DST_DIR=./
protoc -I=$SRC_DIR --go_out=$DST_DIR $SRC_DIR/options/options.proto

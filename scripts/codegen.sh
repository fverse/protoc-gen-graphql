#!/usr/bin/env bash

protoc --plugin=protoc-gen-graphql=./protoc-gen-graphql \
--graphql_out=target=client,keep_case,keep_prefix,combine_output=true:./out hello.proto

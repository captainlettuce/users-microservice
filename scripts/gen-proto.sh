#!/bin/sh

find generated/ -type f -name '*.pb.go' -delete;
for f in proto/*
do
  protoc \
    --go_out=generated \
    --go-grpc_out=generated \
    -I=./proto \
    $f
done

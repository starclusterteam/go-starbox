#!/bin/bash

protoc --go_out=../generated \
			 --go_opt=paths=source_relative \
       --go-grpc_out=../generated \
			 --go-grpc_opt=paths=source_relative \
			test.proto

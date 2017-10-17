#!/bin/bash

CLIENT_COUNT=5
SRC_DIR="go"

# protoc --go_out=$SRC_DIR message.proto

go build -o ${SRC_DIR}/main ${SRC_DIR}/main.go

for ((i=0;i<$CLIENT_COUNT;i+=1))
do
	ttab ${SRC_DIR}/main -n ${CLIENT_COUNT} -pid ${i}
done


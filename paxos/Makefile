all: *.go
	go build *.go
run: *.go
	go run *.go
message.pb.go: message.proto
	protoc --go_out=. message.proto

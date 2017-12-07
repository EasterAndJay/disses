all: *.go
	go build *.go
run: *.go
	go run *.go
euca: *.go
	env GOOS=linux GOARCH=amd64 go build -o paxos-client *.go 
message.pb.go: message.proto
	protoc --go_out=. message.proto

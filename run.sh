#! /bin/sh

go build -o runner cmd/main.go cmd/body.go
go build -o golang_client example/go/client.go
./runner ./golang_client
./runner ruby ./example/ruby/client.rb

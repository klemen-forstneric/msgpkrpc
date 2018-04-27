# msgpkrpc: A MessagePack RPC implementation for Go

![MIT](https://img.shields.io/badge/license-MIT-blue.svg)

The Go implementation of [MessagePack-RPC](https://github.com/msgpack-rpc/msgpack-rpc/blob/master/spec.md): An RPC library that uses the MessagePack format for data serialization.

Installation
============

To install this package, you need to install Go and setup your Go workspace on your computer. The simplest way to install the library is to run:

```
$ go get -u github.com/ReveelTechnologies/msgpkrpc
```

Prerequisites
=============

- Go 1.10
- [vmihailenco/msgpack](github.com/vmihailenco/msgpack)

Examples
========

Client
------
```
type EchoRequest struct {
	_msgpack struct{} `msgpack:",asArray"`

	Id      int
	Message string
}

type EchoResponse struct {
	_msgpack struct{} `msgpack:",asArray"`

	Id      int
	Message string
}

address := "127.0.0.1"
port := 9000

client := rpc.NewClient(address, port)

decoder, err := client.Call(
	"Echo",
	&EchoRequest{Id: 1, Message: "Sample message"})
	
if !decoder.IsValid() {
	log.Printf("No result\n")
	return
}

var response EchoResponse
err = decoder.Decode(&response)

if err != nil {
	log.Fatal(err)
}

log.Printf("Got response: %v\n", response)

```

Server
---------------
```
type EchoRequest struct {
	_msgpack struct{} `msgpack:",asArray"`

	Id      int
	Message string
}

type EchoResponse struct {
	_msgpack struct{} `msgpack:",asArray"`

	Id      int
	Message string
}

type EchoHandler struct {
}

func Handle(request EchoRequest) (*EchoResponse, error) {
	return &EchoResponse{
		Id:      request.Id, 
		Message: request.Message}, nil
}

port := 9000

server := rpc.NewServer(port)
server.Bind("Echo", Handle)

server.Run()
```

Server (with Binders interface)
---------------
```
type EchoRequest struct {
	_msgpack struct{} `msgpack:",asArray"`

	Id      int
	Message string
}

type EchoResponse struct {
	_msgpack struct{} `msgpack:",asArray"`

	Id      int
	Message string
}

type EchoHandler struct {
}

func (e *EchoHandler) Handle(request EchoRequest) (*EchoResponse, error) {
	return &EchoResponse{
		Id:      request.Id, 
		Message: request.Message}, nil
}

func (e *EchoHandler) Bind(s rpc.Server) {
	s.Bind("Echo", func(request EchoRequest) (*EchoResponse, error) {
		return e.Handle(request)
	})
}

port := 9000
server := rpc.NewServerWithBinders(port, []rpc.Binder{&EchoHandler{}})

server.Run()
```

Author
======
Klemen Forstnerič, Reveel Technologies Inc.
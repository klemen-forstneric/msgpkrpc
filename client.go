/**
 * Copyright 2018 Reveel Technologies, Inc. and contributors
 */

package rpc

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/vmihailenco/msgpack"
	"net"
)

type Client interface {
	Call(methodName string, parameters ...interface{}) (Decoder, error)
}

type ClientImpl struct {
	address string
	port    int
}

func NewClient(address string, port int) Client {
	return &ClientImpl{address: address, port: port}
}

type ClientFactory interface {
	Create(address string, port int) Client
}

type ClientFactoryImpl struct {
}

func NewClientFactory() ClientFactory {
	return &ClientFactoryImpl{}
}

func (c *ClientImpl) Call(methodName string, parameters ...interface{}) (Decoder, error) {
	fullAddress := fmt.Sprintf("%s:%d", c.address, c.port)
	conn, err := net.Dial(RpcConnectionType, fullAddress)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	var buffer bytes.Buffer
	encoder := msgpack.NewEncoder(&buffer)

	err = encoder.Encode(&Request{
		Type:       RequestMessageType,
		MessageId:  1,
		MethodName: methodName,
		Parameters: parameters})

	if err != nil {
		return nil, err
	}

	_, err = conn.Write(buffer.Bytes())

	if err != nil {
		return nil, err
	}

	responseReader := bufio.NewReader(conn)
	decoder := msgpack.NewDecoder(responseReader)

	var response Response
	err = decoder.Decode(&response)

	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("%s", response.Error.Description)
	}

	return NewDecoder(response.Result)
}

func (c *ClientFactoryImpl) Create(address string, port int) Client {
	return NewClient(address, port)
}

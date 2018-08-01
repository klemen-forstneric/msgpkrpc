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
	Notify(methodName string, parameters ...interface{}) error
}

type clientImpl struct {
	address string
	port    int
}

func NewClient(address string, port int) Client {
	return &clientImpl{address: address, port: port}
}

type ClientFactory interface {
	Create(address string, port int) Client
}

type clientFactoryImpl struct {
}

func NewClientFactory() ClientFactory {
	return &clientFactoryImpl{}
}

func (c *clientImpl) Call(methodName string, parameters ...interface{}) (Decoder, error) {
	conn, err := c.getConnection()

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	var buffer bytes.Buffer
	encoder := msgpack.NewEncoder(&buffer)

	err = encoder.Encode(&Request{
		Type:       requestMessageType,
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

func (c *clientImpl) Notify(methodName string, parameters ...interface{}) error {
	conn, err := c.getConnection()

	if err != nil {
		return err
	}

	defer conn.Close()

	var buffer bytes.Buffer
	encoder := msgpack.NewEncoder(&buffer)

	err = encoder.Encode(&Request{
		Type:       notificationMessageType,
		MessageId:  1,
		MethodName: methodName,
		Parameters: parameters})

	if err != nil {
		return err
	}

	_, err = conn.Write(buffer.Bytes())

	return err
}

func (c *clientImpl) getConnection() (net.Conn, error) {
	fullAddress := fmt.Sprintf("%s:%d", c.address, c.port)
	return net.Dial(rpcConnectionType, fullAddress)
}

func (c *clientFactoryImpl) Create(address string, port int) Client {
	return NewClient(address, port)
}

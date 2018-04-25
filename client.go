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
	conn, err := c.GetConnection()

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	buffer, err := c.EncodeRequest(
		RequestMessageType,
		methodName,
		parameters)

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

func (c *ClientImpl) Notify(methodName string, parameters ...interface{}) error {
	conn, err := c.GetConnection()

	if err != nil {
		return err
	}

	defer conn.Close()

	buffer, err := c.EncodeRequest(
		NotificationMessageType,
		methodName,
		parameters)

	if err != nil {
		return err
	}

	_, err = conn.Write(buffer.Bytes())

	return err
}

func (c *ClientImpl) GetConnection() (net.Conn, error) {
	fullAddress := fmt.Sprintf("%s:%d", c.address, c.port)
	return net.Dial(RpcConnectionType, fullAddress)
}

func (c *ClientImpl) EncodeRequest(requestType int, methodName string, parameters ...interface{}) (bytes.Buffer, error) {
	var buffer bytes.Buffer
	encoder := msgpack.NewEncoder(&buffer)

	err := encoder.Encode(&Request{
		Type:       requestType,
		MessageId:  1,
		MethodName: methodName,
		Parameters: parameters})

	return buffer, err
}

func (c *ClientFactoryImpl) Create(address string, port int) Client {
	return NewClient(address, port)
}

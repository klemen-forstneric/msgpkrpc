/**
 * Copyright 2018 Reveel Technologies, Inc. and contributors
 */

package rpc

const (
	RpcConnectionType = "tcp"
)

const (
	RequestMessageType  = 0
	ResponseMessageType = 1
)

type Parameters []interface{}

type Request struct {
	_msgpack struct{} `msgpack:",asArray"`

	Type       int
	MessageId  int
	MethodName string
	Parameters Parameters
}

type Error struct {
	_msgpack struct{} `msgpack:",asArray"`

	Description string
	Code        string
}

type Response struct {
	_msgpack struct{} `msgpack:",asArray"`

	Type      int
	MessageId int
	Error     *Error
	Result    interface{}
}

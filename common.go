/**
 * Copyright 2018 Reveel Technologies, Inc. and contributors
 */

package rpc

const (
	rpcConnectionType = "tcp"
)

const (
	requestMessageType  = 0
	responseMessageType = 1
	notificationMessageType = 2
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

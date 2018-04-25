/**
 * Copyright 2018 Reveel Technologies, Inc. and contributors
 */

package rpc

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/vmihailenco/msgpack"
	"log"
	"net"
	"reflect"
)

type RespondFunction func(conn net.Conn, messageId int, rpcError error, rpcResult interface{})

type Function interface{}

type Handler struct {
	Function   Function
	Parameters []reflect.Type
}

type Server interface {
	Bind(name string, function Function)
	Run(port int) error
}

type FunctionBinder interface {
	Bind(s Server)
}

type ServerImpl struct {
	handlers map[string]Handler
}

func Respond(conn net.Conn, messageId int, rpcError error, rpcResult interface{}) {
	var buffer bytes.Buffer
	encoder := msgpack.NewEncoder(&buffer)

	var rpcErrorEncoded *Error = nil
	if rpcError != nil {
		rpcErrorEncoded = &Error{
			Description: rpcError.Error(),
			Code:        ""}
	}

	err := encoder.Encode(&Response{
		Type:      ResponseMessageType,
		MessageId: messageId,
		Error:     rpcErrorEncoded,
		Result:    rpcResult})

	if err != nil {
		log.Printf("Failed to encode RPC response (%v)\n", err)
		return
	}

	_, err = conn.Write(buffer.Bytes())

	if err != nil {
		log.Printf("Failed to send RPC response(%v)\n", err)
	}
}

func EmptyRespond(conn net.Conn, messageId int, rpcError error, rpcResult interface{}) {
}

func NewServer(functionBinders []FunctionBinder) Server {
	server := &ServerImpl{handlers: make(map[string]Handler)}

	for _, b := range functionBinders {
		b.Bind(server)
	}

	return server
}

func (s *ServerImpl) Bind(name string, function Function) {
	if _, exists := s.handlers[name]; exists {
		log.Printf("A function is already bound to name %s! Rebinding..\n", name)
	}

	functionType := reflect.TypeOf(function)

	numParameters := functionType.NumIn()
	parameters := make([]reflect.Type, numParameters)

	for i := 0; i < numParameters; i++ {
		parameters[i] = functionType.In(i)
	}

	s.handlers[name] = Handler{
		Function:   function,
		Parameters: parameters}
}

func (s *ServerImpl) Run(port int) error {
	ln, err := net.Listen(RpcConnectionType, fmt.Sprintf(":%d", port))

	if err != nil {
		return err
	}

	for {
		conn, err := ln.Accept()

		if err != nil {
			return err
		}

		go s.HandleConnection(conn)
	}
}

func (s *ServerImpl) HandleConnection(conn net.Conn) {
	defer conn.Close()

	request, err := s.ParseRequest(conn)

	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	var respond RespondFunction

	switch request.Type {
	case RequestMessageType:
		respond = Respond
	case NotificationMessageType:
		respond = EmptyRespond
	}

	ProcessRequest(conn, request, respond)
}

func (s *ServerImpl) ProcessRequest(conn net.Conn, request Request, respond RespondFunction) {
	handler, exists := s.handlers[request.MethodName]

	if !exists {
		err = fmt.Errorf("No handler exists for method %s", request.MethodName)
		log.Printf("%v\n", err)

		respond(conn, request.MessageId, err, nil)
		return
	}

	if len(handler.Parameters) != len(request.Parameters) {
		err = fmt.Errorf(
			"Parameter count for %s doesn't match. Should be %d, but is %d",
			request.MethodName,
			len(handler.Parameters),
			len(request.Parameters))
		log.Printf("%v\n", err)

		respond(conn, request.MessageId, err, nil)
		return
	}

	parameters, err := s.DecodeParameters(request.Parameters, handler.Parameters)

	if err != nil {
		log.Printf("%v\n", err)
		respond(conn, request.MessageId, err, nil)
		return
	}

	function := reflect.ValueOf(handler.Function)
	values := function.Call(parameters)

	returnValues, err := s.DecodeFunctionResult(values)

	if err != nil {
		log.Printf("%v\n", err)
		respond(conn, request.MessageId, err, nil)
		return
	}

	switch len(returnValues) {
	case 0:
		respond(conn, request.MessageId, nil, nil)
	case 1:
		respond(conn, request.MessageId, nil, returnValues[0])
	default:
		respond(conn, request.MessageId, nil, returnValues)
	}
}

func (s *ServerImpl) ParseRequest(conn net.Conn) (Request, error) {
	requestReader := bufio.NewReader(conn)
	decoder := msgpack.NewDecoder(requestReader)

	var request Request
	err := decoder.Decode(&request)

	if err != nil {
		return Request{}, fmt.Errorf("Failed to decode request (%v)", err)
	}

	return request, nil
}

func (s *ServerImpl) DecodeParameters(parameters Parameters, handlerParameters []reflect.Type) ([]reflect.Value, error) {
	decodedParameters := make([]reflect.Value, len(parameters))

	for i, parameter := range parameters {
		decoder, err := NewDecoder(parameter)

		if err != nil {
			return nil, fmt.Errorf("Failed to create the decoder for a parameter (%v)", err)
		}

		var reflectParameter, decodedParameter reflect.Value
		parameterType := handlerParameters[i]

		if parameterType.Kind() == reflect.Ptr {
			reflectParameter = reflect.New(parameterType.Elem()).Elem()
			decodedParameter = reflectParameter.Addr()
		} else {
			reflectParameter = reflect.New(parameterType).Elem()
			decodedParameter = reflectParameter
		}

		parameterPtr := reflectParameter.Addr().Interface()
		err = decoder.Decode(parameterPtr)

		if err != nil {
			return nil, fmt.Errorf("Failed to decode a parameter (%v)", err)
		}

		decodedParameters[i] = decodedParameter
	}

	return decodedParameters, nil
}

func (s *ServerImpl) DecodeFunctionResult(functionResult []reflect.Value) ([]interface{}, error) {
	numFunctionResults := len(functionResult)

	if numFunctionResults == 0 {
		return make([]interface{}, 0), nil
	}

	err, exists := s.ParseError(functionResult)

	if err != nil {
		return make([]interface{}, 0), err
	}

	if exists {
		numFunctionResults -= 1
	}

	returnValues := make([]interface{}, numFunctionResults)

	for i := 0; i < len(returnValues); i++ {
		var value reflect.Value

		if functionResult[i].Kind() == reflect.Ptr {
			value = functionResult[i].Elem()
		} else {
			value = functionResult[i]
		}

		var returnValue interface{} = nil
		if value.IsValid() {
			returnValue = value.Interface()
		}

		returnValues[i] = returnValue
	}

	return returnValues, nil
}

func (s *ServerImpl) ParseError(functionResult []reflect.Value) (error, bool) {
	numFunctionResults := len(functionResult)
	lastFunctionResult := functionResult[numFunctionResults-1]

	if lastFunctionResult.Kind() != reflect.Interface {
		return nil, false
	}

	errorInterface := reflect.TypeOf((*error)(nil)).Elem()

	if !lastFunctionResult.Type().Implements(errorInterface) {
		return nil, false
	}

	if !lastFunctionResult.Elem().IsValid() {
		return nil, true
	}

	return lastFunctionResult.Interface().(error), true
}

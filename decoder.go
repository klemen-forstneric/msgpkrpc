/**
 * Copyright 2018 Reveel Technologies, Inc. and contributors
 */

package rpc

import (
	"bytes"
	"github.com/vmihailenco/msgpack"
)

type Decoder interface {
	Decode(value interface{}) error
	IsValid() bool
}

type DecoderImpl struct {
	decoder *msgpack.Decoder
	valid   bool
}

func NewDecoder(r interface{}) (Decoder, error) {
	if r == nil {
		return &DecoderImpl{decoder: nil, valid: false}, nil
	}

	var buffer bytes.Buffer

	encoder := msgpack.NewEncoder(&buffer)
	err := encoder.Encode(r)

	if err != nil {
		return nil, err
	}

	decoder := msgpack.NewDecoder(&buffer)
	return &DecoderImpl{decoder: decoder, valid: true}, nil
}

func (r *DecoderImpl) Decode(value interface{}) error {
	return r.decoder.Decode(value)
}

func (r *DecoderImpl) IsValid() bool {
	return r.valid
}

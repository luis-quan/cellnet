// Generated by github.com/luis-quan/cellnet/protoc-gen-msg
// DO NOT EDIT!
// Source: pb.proto

package test

import (
	"github.com/luis-quan/cellnet"
	"github.com/luis-quan/cellnet/codec"
	_ "github.com/luis-quan/cellnet/codec/gogopb"
	"reflect"
)

func init() {

	// pb.proto
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("gogopb"),
		Type:  reflect.TypeOf((*ContentACK)(nil)).Elem(),
		ID:    60952,
	})
}

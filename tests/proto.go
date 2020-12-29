package tests

import (
	"fmt"
	"reflect"

	"github.com/luis-quan/cellnet"
	"github.com/luis-quan/cellnet/codec"
	_ "github.com/luis-quan/cellnet/codec/binary"
	"github.com/luis-quan/cellnet/util"
)

type TestEchoACK struct {
	Msg   string
	Value int32
}

func (self *TestEchoACK) String() string { return fmt.Sprintf("%+v", *self) }

func init() {
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("binary"),
		Type:  reflect.TypeOf((*TestEchoACK)(nil)).Elem(),
		ID:    int(util.StringHash("tests.TestEchoACK")),
	})
}

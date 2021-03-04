package binary

import (
	"github.com/luis-quan/cellnet"
	"github.com/luis-quan/cellnet/codec"
	"github.com/luis-quan/cellnet/serial/binaryserial"
)

type binaryCodec struct {
}

func (self *binaryCodec) Name() string {
	return "binary"
}

func (self *binaryCodec) MimeType() string {
	return "application/binary"
}

func (self *binaryCodec) Encode(msgObj interface{}, ctx cellnet.ContextSet) (data interface{}, err error) {

	return binaryserial.BinaryWrite(msgObj, 4)

}

func (self *binaryCodec) Decode(data interface{}, msgObj interface{}) error {

	return binaryserial.BinaryRead(data.([]byte), msgObj, 4)
}

func init() {

	codec.RegisterCodec(new(binaryCodec))
}

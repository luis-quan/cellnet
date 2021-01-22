package gorillawsheaders

import (
	"encoding/binary"

	"github.com/gorilla/websocket"
	"github.com/luis-quan/cellnet"
	"github.com/luis-quan/cellnet/codec"
	"github.com/luis-quan/cellnet/util"
)

type header struct {
	identity uint8
	encode   uint8
	length   uint16
	version  uint8
	reserve  uint8
	utype    uint16
}

func encodeHeader(data []byte, id uint16) {
	var offset uint8 = 0
	data[offset] = uint8(0x05) //identity
	offset += 1
	data[offset] = uint8(0) //encode
	offset += 1
	binary.LittleEndian.PutUint16(data[offset:offset+2], 0) //length
	offset += 2
	data[offset] = uint8(0x03) //version
	offset += 1
	data[offset] = uint8(0) //reserve
	offset += 1
	binary.LittleEndian.PutUint16(data[offset:offset+2], id) //utype
	offset += 2
}

func decodeHeader(data []byte) uint16 {
	id := binary.LittleEndian.Uint16(data[6:]) //utype
	return id
}

const (
	MsgIDSize = 8 // header
)

type WSMessageTransmitter struct {
}

func (WSMessageTransmitter) OnRecvMessage(ses cellnet.Session) (msg interface{}, id int, err error) {

	conn, ok := ses.Raw().(*websocket.Conn)

	// 转换错误，或者连接已经关闭时退出
	if !ok || conn == nil {
		return nil, 0, nil
	}

	var messageType int
	var raw []byte
	messageType, raw, err = conn.ReadMessage()

	if err != nil {
		return
	}

	if len(raw) < MsgIDSize {
		return nil, 0, util.ErrMinPacket
	}

	switch messageType {
	case websocket.BinaryMessage:
		msgID := decodeHeader(raw)
		msgData := raw[MsgIDSize:]
		msg, _, err = codec.DecodeMessage(int(msgID), msgData)
		id = int(msgID)
	}

	return
}

func (WSMessageTransmitter) OnSendMessage(ses cellnet.Session, msg interface{}) error {

	conn, ok := ses.Raw().(*websocket.Conn)

	// 转换错误，或者连接已经关闭时退出
	if !ok || conn == nil {
		return nil
	}

	var (
		msgData []byte
		msgID   int
	)

	switch m := msg.(type) {
	case *cellnet.RawPacket: // 发裸包
		msgData = m.MsgData
		msgID = m.MsgID
	default: // 发普通编码包
		var err error
		var meta *cellnet.MessageMeta

		// 将用户数据转换为字节数组和消息ID
		msgData, meta, err = codec.EncodeMessage(msg, nil)

		if err != nil {
			return err
		}

		msgID = meta.ID
	}

	pkt := make([]byte, MsgIDSize+len(msgData))
	encodeHeader(pkt, uint16(msgID))
	copy(pkt[MsgIDSize:], msgData)

	conn.WriteMessage(websocket.BinaryMessage, pkt)

	return nil
}

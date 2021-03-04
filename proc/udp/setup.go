package udp

import (
	"github.com/luis-quan/cellnet"
	"github.com/luis-quan/cellnet/msglog"
	"github.com/luis-quan/cellnet/peer/udp"
	"github.com/luis-quan/cellnet/proc"
)

type UDPMessageTransmitter struct {
}

func (UDPMessageTransmitter) OnRecvMessage(ses cellnet.Session) (msg interface{}, id int, err error) {

	data := ses.Raw().(udp.DataReader).ReadData()

	msg, id, err = RecvPacket(data)

	msglog.WriteRecvLogger(log, "udp", ses, msg)

	return
}

func (UDPMessageTransmitter) OnSendMessage(ses cellnet.Session, msg interface{}) error {

	writer := ses.(udp.DataWriter)

	msglog.WriteSendLogger(log, "udp", ses, msg)

	// ses不再被复用, 所以使用session自己的contextset做内存池, 避免串台
	return sendPacket(writer, ses.(cellnet.ContextSet), msg)
}

func init() {

	proc.RegisterProcessor("udp.ltv", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback, args ...interface{}) {

		bundle.SetTransmitter(new(UDPMessageTransmitter))
		bundle.SetCallback(userCallback)

	})
}

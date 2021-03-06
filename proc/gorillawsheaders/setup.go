package gorillawsheaders

import (
	"github.com/luis-quan/cellnet"
	"github.com/luis-quan/cellnet/proc"
)

func init() {

	proc.RegisterProcessor("gorillawsheaders.ltv", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback, args ...interface{}) {

		bundle.SetTransmitter(new(WSMessageTransmitter))
		bundle.SetHooker(new(MsgHooker))
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))

	})
}

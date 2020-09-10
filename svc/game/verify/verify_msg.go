package verify

import (
	"fmt"

	"github.com/davyxu/cellmesh/service"
	"github.com/davyxu/cellnet"

	"github.com/davyxu/cellmesh_demo/proto"
	"github.com/davyxu/cellmesh_demo/svc/agent/api"
)

func init() {

	proto.Handle_Game_VerifyREQ = agentapi.HandleBackendMessage(func(ev cellnet.Event, cid proto.ClientID) {

		msg := ev.Message().(*proto.VerifyREQ)

		fmt.Printf("verfiy: %+v \n", msg.GameToken)

		service.Reply(ev, &proto.VerifyACK{})
	})
}

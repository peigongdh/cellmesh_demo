package hubstatus

import (
	"time"

	"github.com/davyxu/cellmesh/service"
	"github.com/davyxu/cellnet/timer"

	"github.com/davyxu/cellmesh_demo/basefx/model"
	"github.com/davyxu/cellmesh_demo/proto"
	"github.com/davyxu/cellmesh_demo/svc/hub/api"
)

func StartSendStatus(channelName string, updateInterval time.Duration, statusGetter func() int) {

	timer.NewLoop(fxmodel.Queue, updateInterval, func(loop *timer.Loop) {

		var ack proto.SvcStatusACK
		ack.SvcID = service.GetLocalSvcID()
		ack.UserCount = int32(statusGetter())

		hubapi.Publish(channelName, &ack)

	}, nil).Notify().Start()
}

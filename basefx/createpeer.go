package basefx

import (
	"time"

	"github.com/davyxu/cellmesh/discovery"
	"github.com/davyxu/cellmesh/service"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	"github.com/davyxu/cellnet/proc"

	"github.com/davyxu/cellmesh_demo/basefx/model"
	"github.com/davyxu/cellmesh_demo/proto"
)

func CreateCommnicateAcceptor(param fxmodel.ServiceParameter) cellnet.Peer {

	if param.NetPeerType == "" {
		param.NetPeerType = "tcp.Acceptor"
	}

	var q cellnet.EventQueue
	if !param.NoQueue {
		q = fxmodel.Queue
	}

	p := peer.NewGenericPeer(param.NetPeerType, param.SvcName, param.ListenAddr, q)

	msgFunc := proto.GetMessageHandler(param.SvcName)

	// "tcp.svc"
	proc.BindProcessorHandler(p, param.NetProcName, func(ev cellnet.Event) {

		if msgFunc != nil {
			msgFunc(ev)
		}
	})

	if opt, ok := p.(cellnet.TCPSocketOption); ok {
		opt.SetSocketBuffer(2048, 2048, true)
	}

	fxmodel.AddLocalService(p)

	p.Start()

	service.Register(p)

	return p
}

func CreateCommnicateConnector(param fxmodel.ServiceParameter) {
	if param.NetPeerType == "" {
		param.NetPeerType = "tcp.Connector"
	}

	msgFunc := proto.GetMessageHandler(service.GetProcName())

	opt := service.DiscoveryOption{
		MaxCount: param.MaxConnCount,
	}

	opt.Rules = service.LinkRules

	var q cellnet.EventQueue
	if !param.NoQueue {
		q = fxmodel.Queue
	}

	mp := service.DiscoveryService(param.SvcName, opt, func(multiPeer service.MultiPeer, sd *discovery.ServiceDesc) {

		p := peer.NewGenericPeer(param.NetPeerType, param.SvcName, sd.Address(), q)

		proc.BindProcessorHandler(p, param.NetProcName, func(ev cellnet.Event) {

			if msgFunc != nil {
				msgFunc(ev)
			}
		})

		if opt, ok := p.(cellnet.TCPSocketOption); ok {
			opt.SetSocketBuffer(2048, 2048, true)
		}

		p.(cellnet.TCPConnector).SetReconnectDuration(time.Second * 3)

		//
		multiPeer.AddPeer(sd, p)

		p.Start()
	})

	mp.(service.MultiPeer).SetContext("multi", param)

	fxmodel.AddLocalService(mp)

}

func GetRemoteServiceWANAddress(svcName, svcid string) string {

	result := service.QueryService(svcName,
		service.Filter_MatchSvcID(svcid))

	if result == nil {
		return ""
	}

	desc := result.(*discovery.ServiceDesc)

	wanAddr := desc.GetMeta("WANAddress")

	if wanAddr != "" {
		return wanAddr
	}

	// 开发阶段， 返回内网IP
	return desc.Address()
}

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/davyxu/cellmesh/service"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/msglog"
	_ "github.com/davyxu/cellnet/peer/gorillaws"
	"github.com/davyxu/cellnet/timer"
	"github.com/davyxu/golog"

	"github.com/davyxu/cellmesh_demo/proto"
)

var log = golog.New("main")

type ClientParam struct {
	NetPeerType string
	NetProcName string
}

var (
	flagProtocolType = flag.String("commtype", "tcp", "communicate type, tcp or ws")
)

// 登录服务器是一个短连接服务器,获取到网关连接后就断开
func login(param *ClientParam) (agentAddr, gameSvcID string) {

	log.Debugln("Create login connection...")

	// 这里为了方便,使用服务发现来连接login, 真正的客户端不应该使用服务发现,而是使用固定的登录地址连接登录服务器
	loginReq := CreateConnection("login", param.NetPeerType, param.NetProcName)

	// TODO 短连接请求完毕关闭

	// 封装连接和接收以及超时的过程,真正的客户端可以添加这套机制
	RemoteCall(loginReq, &proto.LoginREQ{
		Version:  "1.0",
		Platform: "demo",
		UID:      "1234",
	}, func(ack *proto.LoginACK) {

		if ack.Result == proto.ResultCode_NoError {
			agentAddr = fmt.Sprintf("%s:%d", ack.Server.IP, ack.Server.Port)
			gameSvcID = ack.GameSvcID
		} else {
			panic(ack.Result.String())
		}

	})

	return
}

func getAgentSession(agentAddr string, param *ClientParam) (ret cellnet.Session) {

	log.Debugln("Prepare agent connection...")

	waitGameReady := make(chan struct{})
	go KeepConnection("agent", agentAddr, param.NetPeerType, param.NetProcName, func(ses cellnet.Session) {
		ret = ses
		waitGameReady <- struct{}{}
	}, func() {
		os.Exit(0)
	})

	<-waitGameReady

	log.Debugln("Agent connection ready")

	return
}

func ReadConsole(callback func(string)) {

	for {

		// 从标准输入读取字符串，以\n为分割
		text, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			break
		}

		// 去掉读入内容的空白符
		text = strings.TrimSpace(text)

		callback(text)

	}
}

func main() {

	msglog.SetMsgLogRule("proto.PingACK", msglog.MsgLogRule_BlackList)

	service.Init("client")
	service.ConnectDiscovery()

	var currParam *ClientParam
	switch *flagProtocolType {
	case "tcp":
		currParam = &ClientParam{NetPeerType: "tcp.SyncConnector", NetProcName: "tcp.demo"}
	case "ws":
		currParam = &ClientParam{NetPeerType: "gorillaws.SyncConnector", NetProcName: "ws.demo"}
	}

	agentAddr, gameSvcID := login(currParam)

	if agentAddr == "" {
		return
	}

	fmt.Println("agent:", agentAddr)

	agentReq := getAgentSession(agentAddr, currParam)

	RemoteCall(agentReq, &proto.VerifyREQ{
		GameToken: "verify",
		GameSvcID: gameSvcID,
	}, func(ack *proto.VerifyACK) {

		fmt.Println(ack)
	})

	//
	timer.NewLoop(nil, time.Second*5, func(loop *timer.Loop) {
		agentReq.Send(&proto.PingACK{})

	}, nil).Start()

	fmt.Println("Start chat now !")

	ReadConsole(func(s string) {

		RemoteCall(agentReq, &proto.ChatREQ{
			Content: s,
		}, func(ack *proto.ChatACK) {

			fmt.Println(ack.Content)
		})

	})

}

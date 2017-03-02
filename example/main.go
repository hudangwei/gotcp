package main

import (
    "github.com/hudangwei/gotcp"
    "os"
    "os/signal"
    "syscall"
    "log"
    "net"
    "strconv"
    "time"
    proto "github.com/hudangwei/gotcp/example/protocol"
)

const (
    DEFAULT_TCP_PORT uint16 = 8181
)

type MyEchoServer struct {
    EchoServer *gotcp.Server
}

func (this *MyEchoServer)watch()  {
    chSig := make(chan os.Signal)
    signal.Notify(chSig,syscall.SIGINT,syscall.SIGTERM)

    log.Println("收到系统信号:",<-chSig)

    this.EchoServer.Stop()
}

func main() {
    config := &gotcp.Config{
        PacketSendChanLimit: 20,
        PacketReceiveChanLimit: 20,
    }

    srv := &MyEchoServer{}

    /*
     * 监听端口
     */
    tcpAddr,err := net.ResolveTCPAddr("tcp4","127.0.0.1:"+strconv.Itoa(int(DEFAULT_TCP_PORT)))
    checkError(err)
    listener,err := net.ListenTCP("tcp",tcpAddr)
    checkError(err)

    srv.EchoServer = gotcp.NewServer(config,&proto.EchoServerCallback{},&proto.MyProtocol{})
    go srv.EchoServer.Start(listener,time.Second)

    log.Println("服务已监听:",listener.Addr())

    //监听系统关闭消息
    log.Println("正在监听系统信号，可按Command+C键停止该程序...")
    srv.watch()
    log.Println("程序退出。")
    os.Exit(0)
}

func checkError(err error) {
    if err != nil {
        log.Fatal(err)
    }
}
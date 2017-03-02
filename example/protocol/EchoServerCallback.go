package protocol

import (
	"log"
	"github.com/hudangwei/gotcp"
)

type EchoServerCallback struct {
}

func (this *EchoServerCallback) OnConnect(c *gotcp.Conn) bool {
	addr := c.GetRawConn().RemoteAddr()
	c.PutExtraData(addr)
	log.Println("连接成功:", addr)
	//	c.AsyncWritePacket(NewLolLauncherPacket(0, []byte("onconnect")), 0)
	return true
}

func (this *EchoServerCallback) OnMessage(c *gotcp.Conn, p gotcp.Packet) bool {
	packet := p.(*MyPacket)
	data := packet.GetData()
	commandType := packet.GetCommand()
	switch commandType {
	case 0x00:
		c.AsyncWritePacket(NewMyPacket(0x01, []byte{}), 0)
		log.Println("data:", string(data))
	default:
		log.Println("OnMessage:", commandType, " -- ", packet.pHead, " -- ", data)
	}

	return true
}

func (this *EchoServerCallback) OnClose(c *gotcp.Conn) {
	log.Println("已经退出:", c.GetExtraData())
}

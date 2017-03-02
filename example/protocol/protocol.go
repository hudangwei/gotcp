package protocol

import (
	"bytes"
	"encoding/binary"
	"github.com/hudangwei/gotcp"
	"net"
)

const (
	PACKET_HEADER_SIZE = 16
)

type Header struct {
	pHead0   uint32 //默认0x10
	pHead1   uint32 //默认0x01
	pCommand uint32 //默认0x00
	pLength  uint32 //默认0x00
}

func (this *Header) Bytes() [PACKET_HEADER_SIZE]byte {
	var p [PACKET_HEADER_SIZE]byte
	binary.LittleEndian.PutUint32(p[:4], this.pHead0)
	binary.LittleEndian.PutUint32(p[4:8], this.pHead1)
	binary.LittleEndian.PutUint32(p[8:12], this.pCommand)
	binary.LittleEndian.PutUint32(p[12:16], this.pLength)
	return p
}

func (head *Header) Read(buf []byte) {
	head.pHead0 = binary.LittleEndian.Uint32(buf[:4])
	head.pHead1 = binary.LittleEndian.Uint32(buf[4:8])
	head.pCommand = binary.LittleEndian.Uint32(buf[8:12])
	head.pLength = binary.LittleEndian.Uint32(buf[12:16])
}

// Packet
type MyPacket struct {
	pHead Header
	pData []byte
}

func (p *MyPacket) Serialize() []byte {
	// 拼装head部分
	length := len(p.pData)
	p.pHead.pLength = uint32(length)

	buff := make([]byte, PACKET_HEADER_SIZE+length)
	head := p.pHead.Bytes()
	copy(buff[:PACKET_HEADER_SIZE], head[:PACKET_HEADER_SIZE])
	copy(buff[PACKET_HEADER_SIZE:], p.pData)
	return buff
}

func (p *MyPacket) GetCommand() uint32 {
	return p.pHead.pCommand
}

func (p *MyPacket) GetHeader() Header {
	return p.pHead
}

func (p *MyPacket) GetData() []byte {
	return p.pData
}

func NewMyPacket(pCommand uint32, pData []byte) *MyPacket {
	head := Header{
		pHead0:   0x10,
		pHead1:   0x01,
		pCommand: pCommand,
		pLength:  uint32(len(pData)),
	}
	return &MyPacket{
		pHead: head,
		pData: pData,
	}
}

type MyProtocol struct {
	unfinished bool //是否为未完成的TCP 包，默认false
	header     Header
	body       []byte
}

func (this *MyProtocol) ReadPacket(conn *net.TCPConn) (gotcp.Packet, error) {
	fullBuf := bytes.NewBuffer([]byte{})
	for {
		data := make([]byte, 1024)

		//暂时不支持超过1024字节长度的单个TCP包
		readLength, err := conn.Read(data)

		if err != nil { //EOF, or worse
			return nil, err
		}

		if readLength == 0 {
			return nil, gotcp.ErrConnClosing
		} else {
			fullBuf.Write(data[:readLength])
		}

		//粘包处理，先判断header.pLenght 是否为0，若为0，则为新的包
		if this.unfinished == true {
			if this.header.pLength <= 0 {
				//内部错误
				return nil, gotcp.ErrReadBlocking
			}
			//是未完成的包
			body := fullBuf.Next(int(this.header.pLength))
			if len(body) != int(this.header.pLength) {
				return nil, gotcp.ErrReadBlocking
			}
			this.body = body
		} else {
			//新包
			//转化到header中
			head := fullBuf.Next(16)
			if len(head) < 16 {
				return nil, gotcp.ErrReadBlocking
			}
			this.header = Header{}
			this.header.Read(head)
			length := int(this.header.pLength)
			body := make([]byte, length)
			body = fullBuf.Next(length)
			if len(body) < length {
				this.unfinished = true
				return nil, nil
			}
			this.body = body
		}

		//设置为已经完成的包
		this.unfinished = false
		return NewMyPacket(this.header.pCommand, this.body), nil
	}
}

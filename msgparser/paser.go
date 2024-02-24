package msgparser

import (
	"bytes"
	"encoding/binary"
	"errors"
	"gobonbon/conf"
	"gobonbon/iface"
)

type MsgParser struct {
	littleEndian bool // 大小端
}

func NewMsgParser() *MsgParser {
	p := new(MsgParser)
	p.littleEndian = false
	return p
}

// 设置大小端
func (p *MsgParser) SetByteOrder(littleEndian bool) {
	p.littleEndian = littleEndian
}

// 获取包头长度方法
func (p *MsgParser) GetHeadLen() uint32 {
	//Id uint32(4字节) +  DataLen uint32(4字节)
	return 8
}

// 编码
func (p *MsgParser) Encode(msg iface.IMessage) ([]byte, error) {
	bufMsgLen := bytes.NewBuffer([]byte{})

	if p.littleEndian {
		// Write the message ID
		if err := binary.Write(bufMsgLen, binary.LittleEndian, msg.GetMsgId()); err != nil {
			return nil, err
		}
		// Write the data length
		if err := binary.Write(bufMsgLen, binary.LittleEndian, msg.GetDataLen()); err != nil {
			return nil, err
		}
		// Write the data
		if err := binary.Write(bufMsgLen, binary.LittleEndian, msg.GetData()); err != nil {
			return nil, err
		}
	} else {
		if err := binary.Write(bufMsgLen, binary.BigEndian, msg.GetMsgId()); err != nil {
			return nil, err
		}
		if err := binary.Write(bufMsgLen, binary.BigEndian, msg.GetDataLen()); err != nil {
			return nil, err
		}
		if err := binary.Write(bufMsgLen, binary.BigEndian, msg.GetData()); err != nil {
			return nil, err
		}
	}
	return bufMsgLen.Bytes(), nil
}

// 解码
func (p *MsgParser) Decode(binaryData []byte) (iface.IMessage, error) {
	bufMsgLen := bytes.NewReader(binaryData)

	// (只解压head的信息，得到dataLen和msgID)
	msg := &Message{}

	if p.littleEndian {
		// Write the message ID
		if err := binary.Read(bufMsgLen, binary.LittleEndian, &msg.Id); err != nil {
			return nil, err
		}
		// Write the data length
		if err := binary.Read(bufMsgLen, binary.LittleEndian, &msg.DataLen); err != nil {
			return nil, err
		}
	} else {
		if err := binary.Read(bufMsgLen, binary.BigEndian, &msg.Id); err != nil {
			return nil, err
		}
		if err := binary.Read(bufMsgLen, binary.BigEndian, &msg.DataLen); err != nil {
			return nil, err
		}

	}
	// (判断dataLen的长度是否超出我们允许的最大包长度)
	if conf.GlobalObject.MaxPacketSize > 0 && msg.GetDataLen() > conf.GlobalObject.MaxPacketSize {
		return nil, errors.New("too large msg data received")
	}
	// (这里只需要把head的数据拆包出来就可以了，然后再通过head的长度，再从conn读取一次数据)
	return msg, nil
}

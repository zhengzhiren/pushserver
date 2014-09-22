package packet

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
)

// packet type
const (
	PKT_Regist		= uint8(1)
	PKT_ACK			= uint8(2)
	PKT_Heartbeat	= uint8(3)
	PKT_Push		= uint8(4)
)

const PKT_HEADER_SIZE = 9

type PktHeader struct {
	Type uint8
	Len  uint32		// data length, not including header length
	Id uint32
}

type Pkt struct {
	Header PktHeader
	Data []byte
}

// Pkt to bytes
func (this *Pkt) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, this.Header); err != nil {
		return nil, err
	}
	var b []byte
	b = append(b, buf.Bytes()...)
	b = append(b, this.Data...)
	return b, nil
}

// bytes to PktHeader
func (this *PktHeader) Deserialize(b []byte) (error) {
	buf := bytes.NewReader(b)
	if err := binary.Read(buf, binary.BigEndian, this); err != nil {
		return err
	}
	return nil
}

func Pack (pktType uint8, pktId uint32, data interface {}) (*Pkt, error) {
	buf, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	pkt := new(Pkt)
	pkt.Header.Type = pktType
	pkt.Header.Id = pktId
	pkt.Header.Len = uint32(len(buf))
	pkt.Data = buf
	return pkt, nil
}

func Unpack(pkt *Pkt, data interface{}) error {
	err := json.Unmarshal(pkt.Data, data)
	if err != nil {
		return err
	}
	return nil
}

type PktDataRegist struct {
	DevId string
}

type PktDataMessage struct {
	Msg string
}

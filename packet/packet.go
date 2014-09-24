package packet

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
)

// packet type
const (
	PKT_Init        = iota
	PKT_Init_Resp   = iota
	PKT_Regist      = iota
	PKT_Regist_Resp = iota
	PKT_ACK         = iota
	PKT_Heartbeat   = iota
	PKT_Push        = iota
)

const PKT_HEADER_SIZE = 10

type PktHeader struct {
	Type uint8
	Ver  uint8
	Id   uint32
	Len  uint32 // data length, not including header length
}

type Pkt struct {
	Header PktHeader
	Data   []byte
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
func (this *PktHeader) Deserialize(b []byte) error {
	buf := bytes.NewReader(b)
	if err := binary.Read(buf, binary.BigEndian, this); err != nil {
		return err
	}
	return nil
}

func Pack(pktType uint8, pktId uint32, data interface{}) (*Pkt, error) {
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

type PktDataInit struct {
	DevId string
}

type PktDataInitResp struct {
}

type PktDataRegist struct {
	AppIds []string
}

type PktDataMessage struct {
	Msg   string
	AppId string
}

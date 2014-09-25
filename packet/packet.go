package packet

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
)

// packet type
const (
	PKT_Heartbeat     = iota
	PKT_Init          = iota
	PKT_Init_Resp     = iota
	PKT_Regist        = iota
	PKT_Regist_Resp   = iota
	PKT_Unregist      = iota
	PKT_Unregist_Resp = iota
	PKT_Push          = iota
	PKT_ACK           = iota
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
	pkt := new(Pkt)
	pkt.Header.Type = pktType
	pkt.Header.Id = pktId

	if data != nil {
		buf, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		pkt.Header.Len = uint32(len(buf))
		pkt.Data = buf
	} else {
		pkt.Header.Len = 0
		pkt.Data = nil
	}

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
	DevId string `json:"device_id"`
}

type PktDataInitResp struct {
	Result bool   `json:"result"`
}

type PktDataRegist struct {
	AppId  string `json:"app_id"`
	AppKey string `json:"app_key"`
	RegId  string `json:"reg_id"`
}

type PktDataRegResp struct {
	AppId  string `json:"app_id"`
	RegId  string `json:"reg_id"`
	Result bool   `json:"result"`
}

type PktDataUnregist struct {
	AppId  string `json:"app_id"`
	AppKey string `json:"app_key"`
	RegId  string `json:"reg_id"`
}

type PktDataUnregResp struct {
	AppId  string `json:"app_id"`
	RegId  string `json:"reg_id"`
	Result bool   `json:"result"`
}

type PktDataMessage struct {
	Msg   string
	AppId string
}

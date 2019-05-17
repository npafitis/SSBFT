package types

import (
	"SSBFT/logger"
	"bytes"
	"encoding/gob"
)

type ClientMessage struct {
	Req *Request
	Ack bool
}

func (cm *ClientMessage) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&cm.Req)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	err = decoder.Decode(&cm.Ack)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	return nil
}

func (cm *ClientMessage) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(cm.Req)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	err = encoder.Encode(cm.Ack)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	return w.Bytes(), nil
}

func (cm *ClientMessage) Equals(cmsg *ClientMessage) bool {
	return cm.Req.Equals(cmsg.Req) && cm.Ack == cmsg.Ack
}

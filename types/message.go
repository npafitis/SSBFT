package types

import (
	"SSBFT/logger"
	"SSBFT/variables"
	"bytes"
	"encoding/gob"
)

type Message struct {
	Payload []byte
	Type    string
	From    int
}

func NewMessage(payload []byte, Type string) Message {
	return Message{Payload: payload, Type: Type, From: variables.Id}
}

func (m *Message) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&m.Payload)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	err = decoder.Decode(&m.Type)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	err = decoder.Decode(&m.From)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	return nil
}

func (m Message) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(m.Payload)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	err = encoder.Encode(m.Type)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	err = encoder.Encode(m.From)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	return w.Bytes(), nil
}

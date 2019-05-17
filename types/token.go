package types

import (
	"SSBFT/logger"
	"bytes"
	"encoding/gob"
)

type Token struct {
	FDSet int
	PrimSusp bool
}

func (t *Token) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(t.FDSet)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	err = encoder.Encode(t.PrimSusp)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	return w.Bytes(), nil
}

func (t *Token) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&t.FDSet)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	err = decoder.Decode(&t.PrimSusp)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	return nil
}



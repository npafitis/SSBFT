package types

import (
	"SSBFT/logger"
	"bytes"
	"encoding/gob"
)

type VCM struct {
	VStatus    VStatus
	Prim       int
	NeedChange bool
	NeedChgSet []int
}

func (vcm *VCM) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&vcm.VStatus)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	err = decoder.Decode(&vcm.Prim)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	err = decoder.Decode(&vcm.NeedChange)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	err = decoder.Decode(&vcm.NeedChgSet)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	return nil
}

func (vcm *VCM) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(vcm.VStatus)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	err = encoder.Encode(vcm.Prim)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	err = encoder.Encode(vcm.NeedChange)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	err = encoder.Encode(vcm.NeedChgSet)
	return w.Bytes(), nil
}

type VStatus int

const (
	OK        VStatus = 1
	NoService VStatus = 2
	VChange   VStatus = 3
)

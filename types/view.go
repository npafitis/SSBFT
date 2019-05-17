package types

import (
	"SSBFT/logger"
	"bytes"
	"encoding/gob"
)

type Phase int

const (
	ZERO Phase = 0
	ONE  Phase = 1
)


type AutomatonInfo struct {
	View    int
	Phase   Phase
	VChange bool
	Witness bool
}

type VPair struct {
	Cur  int
	Next int
}

func (vp VPair) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(vp.Cur)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	err = encoder.Encode(vp.Next)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	return w.Bytes(), nil
}

func (vp *VPair) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&vp.Cur)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	err = decoder.Decode(&vp.Next)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	return nil
}

func (vp VPair) Equals(pair VPair) bool {
	return vp.Cur == pair.Cur && vp.Next == pair.Next
}

type Mode int

const (
	Remain Mode = 0
	Follow Mode = 1
)

type Type int

const (
	ACT  Type = 0
	PRED Type = 1
)

type ViewVChange struct {
	View       VPair
	ViewChange bool
}

//func (vv *ViewVChange) GobEncoder() ([]byte, error) {
//	w := new(bytes.Buffer)
//	encoder := gob.NewEncoder(w)
//	err := encoder.Encode(vv.View)
//	if err != nil {
//		logger.ErrLogger.Fatal(err)
//	}
//	err = encoder.Encode(vv.ViewChange)
//	if err != nil {
//		logger.ErrLogger.Fatal(err)
//	}
//	return w.Bytes(), nil
//}
//
//func (vv *ViewVChange) GobDecode(buf []byte) error {
//	r := bytes.NewBuffer(buf)
//	decoder := gob.NewDecoder(r)
//	err := decoder.Decode(vv.View)
//	if err != nil {
//		logger.ErrLogger.Fatal(err)
//	}
//	err = decoder.Decode(&vv.ViewChange)
//	if err != nil {
//		logger.ErrLogger.Fatal(err)
//	}
//	return nil
//}

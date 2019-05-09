package types

import (
	"bytes"
	"encoding/gob"
)

type CoordinationMessage struct {
	Phase      Phase
	Views      []VPair
	Witness    bool
	Pair       ViewVChange
	LastReport *LastReport
}

type LastReport struct {
	Phase   Phase
	Witness bool
	Pair    ViewVChange
}

func (cm *CoordinationMessage) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(cm.Phase)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(cm.Views)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(cm.Witness)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(cm.Pair)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(cm.LastReport)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (cm *CoordinationMessage) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&cm.Phase)
	if err != nil {
		return err
	}
	err = decoder.Decode(&cm.Views)
	if err != nil {
		return err
	}
	err = decoder.Decode(&cm.Witness)
	if err != nil {
		return err
	}
	err = decoder.Decode(&cm.Pair)
	if err != nil {
		return err
	}
	err = decoder.Decode(&cm.LastReport)
	if err != nil {
		return err
	}
	return nil
}

func (lr *LastReport) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(lr.Phase)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(lr.Witness)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(lr.Pair)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (lr *LastReport) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&lr.Phase)
	if err != nil {
		return err
	}
	err = decoder.Decode(&lr.Witness)
	if err != nil {
		return err
	}
	err = decoder.Decode(&lr.Pair)
	if err != nil {
		return err
	}
	return nil
}

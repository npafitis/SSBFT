package types

import (
	"bytes"
	"encoding/gob"
	"time"
)

// TODO: Operation is unfinished
type Operation struct {
	Op    OP
	Value rune
}

func (op Operation) Equals(operation Operation) bool {
	return op.Value == operation.Value && op.Op == operation.Op
}

func (op Operation) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&op.Op)
	if err != nil {
		return err
	}
	err = decoder.Decode(&op.Value)
	if err != nil {
		return err
	}
	return nil
}

func (op Operation) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(op.Op)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(op.Value)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

type OP int

const (
	ADD OP = 0
)

type Request struct {
	Client    int
	TimeStamp time.Time
	Operation Operation
}

func (r *Request) ThisIsRequest() {
	return
}

func (r *Request) GobDecode(buf []byte) error {
	read := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(read)
	err := decoder.Decode(&r.Client)
	if err != nil {
		return err
	}
	err = decoder.Decode(&r.TimeStamp)
	if err != nil {
		return err
	}
	err = decoder.Decode(&r.Operation)
	if err != nil {
		return err
	}
	return nil
}

func (r *Request) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(r.Client)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(r.TimeStamp)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

type Reply struct {
	View      int
	TimeStamp time.Time
	Client    int
	Id        int
	Result    RepState
}

func (rep *Reply) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&rep.View)
	if err != nil {
		return err
	}
	err = decoder.Decode(&rep.TimeStamp)
	if err != nil {
		return err
	}
	err = decoder.Decode(&rep.Client)
	if err != nil {
		return err
	}
	err = decoder.Decode(&rep.Id)
	if err != nil {
		return err
	}
	err = decoder.Decode(&rep.Result)
	if err != nil {
		return err
	}
	return nil
}

func (rep *Reply) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(rep.View)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(rep.TimeStamp)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(rep.Client)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(rep.Id)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(rep.Result)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

type RequestReply struct {
	Req    *Request // Request TODO []: Check between AcceptedRequest and Request
	Client int
	Rep    *Reply
}

func (rr *RequestReply) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&rr.Req)
	if err != nil {
		return err
	}
	err = decoder.Decode(&rr.Client)
	if err != nil {
		return err
	}
	err = decoder.Decode(&rr.Rep)
	if err != nil {
		return err
	}
	return nil
}

func (rr *RequestReply) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(rr.Req)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(rr.Client)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(rr.Rep)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

type AcceptedRequest struct {
	// Request
	Request *Request
	// View
	View int
	// Sequence Number
	Sq int
}

func (r *AcceptedRequest) ThisIsRequest() {
	return
}

func (r *AcceptedRequest) GobDecode(buf []byte) error {
	read := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(read)
	err := decoder.Decode(&r.Request)
	if err != nil {
		return err
	}
	err = decoder.Decode(r.View)
	if err != nil {
		return err
	}
	err = decoder.Decode(r.Sq)
	if err != nil {
		return err
	}
	return nil
}

func (r *AcceptedRequest) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(r.Request)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(r.View)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(r.Sq)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), err
}

func (r *Request) Equals(request *Request) bool {
	return r.Client == request.Client && r.TimeStamp.Equal(request.TimeStamp) &&
		r.Operation.Equals(request.Operation)
}

func (rep *Reply) Equals(request *Reply) bool {
	return rep.Client == request.Client && rep.TimeStamp.Equal(request.TimeStamp) &&
		rep.Id == request.Id && rep.Result == request.Result
}

func (r *AcceptedRequest) Equals(req *AcceptedRequest) bool {
	return r.Request.Equals(req.Request) && r.View == req.View &&
		r.Sq == req.Sq
}

func MergeRequestSets(r []*Request, q []*Request) []*Request {
	var set []*Request
	set = append([]*Request(nil), r...)
	for _, el := range r {
		flag := false
		for _, e := range q {
			if el.Equals(e) {
				flag = true
				break
			}
		}
		if !flag {
			set = append(set, el)
		}
	}
	return set
}

func (rr *RequestReply) Equals(req *RequestReply) bool {
	return rr.Req.Equals(req.Req) && (rr.Rep.Equals(rr.Rep) || rr.Rep == nil)
}

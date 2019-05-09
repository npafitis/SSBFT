package types

import (
	"SSBFT/helper"
	"bytes"
	"container/list"
	"encoding/gob"
	"reflect"
)

type PendReqs *list.List
type ReqQ *list.List

/**
Data structure used in rLog
*/
type LogTuple struct {
	// Request TODO [] : Check between AcceptedRequest and Request
	Req *AcceptedRequest
	// TODO: Type to be defined
	XSet []*ReplicaStructure
}

func (lt *LogTuple) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&lt.Req)
	if err != nil {
		return err
	}
	err = decoder.Decode(&lt.XSet)
	if err != nil {
		return err
	}
	return nil
}

func (lt *LogTuple) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(lt.Req)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(lt.XSet)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (lt *LogTuple) Equals(tp *LogTuple) bool {
	if len(lt.XSet) != len(tp.XSet) {
		return false
	}
	for i := range lt.XSet {
		if lt.XSet[i] != tp.XSet[i] {
			return false
		}
	}
	return lt.Req.Equals(tp.Req)
}

type Status int

const (
	PRE_PREP Status = 0
	PREP     Status = 1
	COMMIT   Status = 2
)

type RequestStatus struct {
	Req *AcceptedRequest
	St  Status
}

func (rs *RequestStatus) Equals(r *RequestStatus) bool {
	return rs.Req.Equals(r.Req) && rs.St == r.St
}

//TODO Check for capacities
type ReplicaStructure struct {
	RepState    RepState
	RLog        []*LogTuple
	PendReqs    *list.List //Queue Of Request
	ReqQ        *list.List //Queue Of RequestStatus
	LastReq     []*RequestReply
	ConFlag     bool
	ViewChanged bool
	Prim        int
}

func (rs *ReplicaStructure) Equals(repl *ReplicaStructure) bool {
	return rs.RepState.Equals(repl.RepState) &&
		helper.RLogEquals(rs.RLog, repl.RLog) &&
		helper.ListEquals(rs.PendReqs, repl.PendReqs) &&
		helper.ListEquals(rs.ReqQ, repl.ReqQ) &&
		helper.LastReqEquals(rs.LastReq, repl.LastReq) &&
		rs.ConFlag == repl.ConFlag &&
		rs.ViewChanged == repl.ViewChanged &&
		rs.Prim == repl.Prim

}

func (rs *ReplicaStructure) GobDecode(buf []byte) error {
	read := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(read)
	err := decoder.Decode(&rs.RepState)
	if err != nil {
		return err
	}
	err = decoder.Decode(&rs.RLog)
	if err != nil {
		return err
	}
	var pendReqs []*Request
	err = decoder.Decode(&pendReqs)
	if err != nil {
		return err
	}
	rs.PendReqs = list.New()
	for _, req := range pendReqs {
		rs.PendReqs.PushBack(req)
	}
	var reqQ []*RequestStatus
	err = decoder.Decode(&reqQ)
	rs.ReqQ = list.New()
	for _, req := range reqQ {
		rs.ReqQ.PushBack(req)
	}
	err = decoder.Decode(&rs.LastReq)
	if err != nil {
		return err
	}
	err = decoder.Decode(&rs.ConFlag)
	if err != nil {
		return err
	}
	err = decoder.Decode(&rs.ViewChanged)
	if err != nil {
		return err
	}
	err = decoder.Decode(&rs.Prim)
	if err != nil {
		return err
	}
	return nil
}

func (rs *ReplicaStructure) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(rs.RepState)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(rs.RLog)
	if err != nil {
		return nil, err
	}
	var pendReqs []*Request
	for e := rs.PendReqs.Front(); e != nil; e = e.Next() {
		pendReqs = append(pendReqs, e.Value.(*Request))
	}
	err = encoder.Encode(pendReqs)
	if err != nil {
		return nil, err
	}
	var reqQ []*RequestStatus
	for e := rs.ReqQ.Front(); e != nil; e = e.Next() {
		reqQ = append(reqQ, e.Value.(*RequestStatus))
	}
	err = encoder.Encode(reqQ)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(rs.LastReq)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(rs.ConFlag)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(rs.ViewChanged)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(rs.Prim)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (rs *ReplicaStructure) LastExec() *LogTuple {
	if len(rs.RLog) < 1 {
		return nil
	}
	max := rs.RLog[0]
	for _, element := range rs.RLog {
		if element.Req.Sq > max.Req.Sq {
			max = element
		}
	}
	return max
}

//TODO: Apply this to all sections necessary.
func ContainsRequest(x *Request, arr []*Request) bool {
	for _, req := range arr {
		if req.Equals(x) {
			return true
		}
	}
	return false
}

func ContainsAcceptedRequest(x *AcceptedRequest, arr []*AcceptedRequest) bool {
	for _, req := range arr {
		if req.Equals(x) {
			return true
		}
	}
	return false
}

func GetRequestsListFromLog(rLog []*LogTuple) *list.List {
	reqs := list.New()
	for _, log := range rLog {
		reqs.PushBack(log.Req)
	}
	return reqs
}

func GetRequestsFromLog(rLog []*LogTuple) []*AcceptedRequest {
	var reqs []*AcceptedRequest
	for _, log := range rLog {
		reqs = append(reqs, log.Req)
	}
	return reqs
}

/*************** OPERATORS **************/
/**
enqueue(x) adds an element (or set of elements) x to a queue.
If any element enqueued already exists, then only the most
recent copy of it is kept and it is carried back to the
queue.
*/
func (rs *ReplicaStructure) Enqueue(el ...interface{}) {
	if len(el) == 0 {
		return
	}
	elType := reflect.TypeOf(el).Elem().Name()
	switch elType {
	case "Request":
		for _, element := range el {
			element := element.(*Request)
			for e := rs.PendReqs.Front(); e != nil; e = e.Next() {
				if e.Value.(*Request).Equals(element) {
					rs.PendReqs.Remove(e)
				}
			}
			rs.PendReqs.PushBack(element)
		}

	case "RequestStatus":
		for _, element := range el {
			element := element.(*RequestStatus)
			for e := rs.ReqQ.Front(); e != nil; e = e.Next() {
				if e.Value.(*RequestStatus).Equals(element) {
					rs.ReqQ.Remove(e)
				}
			}
			rs.ReqQ.PushBack(element)
		}
		break
	case "RequestReply":
		for _, element := range el {
			element := element.(*RequestReply)
			for i, reqRep := range rs.LastReq {
				if element.Equals(reqRep) {
					rs.LastReq = append(rs.LastReq[:i], rs.LastReq[i+1:]...)
				}
			}
			rs.LastReq = append(rs.LastReq, element)
		}
	}
}

/**
remove(x) removes element x from a structure
*/
func (rs *ReplicaStructure) Remove(el ...interface{}) {
	if len(el) == 0 {
		return
	}
	elType := reflect.TypeOf(el[0]).Elem().Name()
	switch elType {
	case "Request":
		for _, element := range el {
			element := element.(*Request)
			for e := rs.PendReqs.Front(); e != nil; e = e.Next() {
				if e.Value.(*Request).Equals(element) {
					rs.PendReqs.Remove(e)
				}
			}
		}

	case "AcceptedRequest":
		for _, element := range el {
			element := element.(*AcceptedRequest)
			for e := rs.ReqQ.Front(); e != nil; e = e.Next() {
				if e.Value.(*RequestStatus).Req.Equals(element) {
					rs.ReqQ.Remove(e)
				}
			}
		}
	case "RequestReply":
		for _, element := range el {
			element := element.(*RequestReply)
			for i, reqRep := range rs.LastReq {
				if element.Equals(reqRep) {
					rs.LastReq = append(rs.LastReq[:i], rs.LastReq[i+1:]...)
				}
			}
		}
	}
}

/**
add(x) adds element x to a structure
*/
func (rs *ReplicaStructure) Add(el ...interface{}) {
	if len(el) == 0 {
		return
	}
	elType := reflect.TypeOf(el[0]).Elem().Name()
	switch elType {
	case "Request":
		for _, element := range el {
			element := element.(*Request)
			rs.PendReqs.PushBack(element)
		}
		break
	case "AcceptedRequest":
		for _, element := range el {
			element := element.(*AcceptedRequest)
			rs.ReqQ.PushBack(&RequestStatus{Req: element, St: PRE_PREP})
		}
		break
	case "RequestReply":
		for _, element := range el {
			element := element.(*RequestReply)
			rs.LastReq = append(rs.LastReq, element)
		}
		break
	case "LogTuple":
		for _, element := range el {
			element := element.(*LogTuple)
			rs.RLog = append(rs.RLog, element)
		}

	}
}

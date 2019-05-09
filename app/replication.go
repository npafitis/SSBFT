package app

import (
	"SSBFT/app/messenger"
	"SSBFT/helper"
	"SSBFT/types"
	"SSBFT/variables"
	"container/list"
	"errors"
	"reflect"
	"time"
)

/**
Constants
*/
//Number of processors
//var N int
var MAXINT int
//var F int
/*****************/
// Identifier(i) of processor p[i]
//var types.Id int

var (
	// System-Defined Constant
	Sigma int
)

// Default/Incorruptible Fallback state
//var DEF_STATE = &types.ReplicaStructure{RepState: nil, RLog: nil, PendReqs: list.New(), ReqQ: list.New(), LastReq: nil, ConFlag: false, ViewChanged: false}
var seqn int // [0, MAXINT]
var rep []*types.ReplicaStructure
var needFlush bool
var flush bool
//var types.Prim int

func DEF_STATE() *types.ReplicaStructure {
	return &types.ReplicaStructure{RepState: make([]rune, 0), RLog: make([]*types.LogTuple, 0), PendReqs: list.New(), ReqQ: list.New(), LastReq: make([]*types.RequestReply, 0), ConFlag: false, ViewChanged: false}
}

/*
Interface for BFT
*/

func ReplicationInit() {
	rep = make([]*types.ReplicaStructure, variables.N)
}
func GetPendReqs() (set []*types.Request, err error) {
	if !rep[variables.Id].ViewChanged {
		return types.MergeRequestSets(knownPendReqs(), unassignedReqs()), nil
	}
	return make([]*types.Request, 0), errors.New("View Change")

}

func RepRequestReset() bool {
	if needFlush {
		needFlush = false
		return true
	}
	return false
}

func GetFlush() bool {
	return flush
}

func ReplicaFlush() {
	flush = true
}

/*
Macros for BFT
*/

func viewChanged() bool {
	return rep[variables.Id].ViewChanged
}

/**
findConsState(S, x) returns a consolidated replica state rep[] based on a set
of processor states S with types non-empty repState prefix and consistency
among request queues reqQ and pendReqs(). It returns ⊥ if such a replica state set
does not exist (indicated as nil); It produces dummy requests in the case where
at least 3f + 1 processors appear to have committed a sufficient number of requests
but they have no evidence of a previous request exists or is assigned. This request
is blocking the execution of the requests that follow.

*/
func findConsState(S []types.RepState) struct {
	RepState types.RepState
	Rlog     []*types.LogTuple
} {
	return struct {
		RepState types.RepState
		Rlog     []*types.LogTuple
	}{RepState: nil, Rlog: nil}
}

/**
TODO checkNewVstate()
checkNewVstate() checks the state proposed by a newly installed primary after
view change. This involves checking whether the proposed pre-prepare messages of committed
processors are verified by another 3f+1 processors and the new state has a correct
prefix as per findConsState().
*/
func checkNewVState(id int) bool {
	return false
}

/**
TODO renewReqs()
renewReqs() is executed by a new primary, in order to issue a consistent set
of pending requests messages for reqQ and pendReqs, where these are now allocated for execution
to the new view.
*/
func renewReqs() {
}

// TODO apply(x)
func apply(x *types.Request) *types.Reply {
	reply := new(types.Reply)
	switch x.Operation.Op {
	case types.ADD:
		rep[variables.Id].RepState = append(rep[variables.Id].RepState, x.Operation.Value)
		reply.TimeStamp = time.Now()
		reply.Id = variables.Id
		reply.Result = rep[variables.Id].RepState
		reply.Client = x.Client
		reply.View = variables.View
		break
	}
	return reply
}

/**************** MACROS ****************/
/**
flushLocal() resets all local values of rep[].
*/
func flushLocal() {
	seqn = 0
	for i := 0; i < variables.N; i++ {
		rep[i] = DEF_STATE()
	}
}

/**
messenger(status t, int j) returns all the requests the p[j]
reported to p[i] that have a specific status or set of
statuses
*/
func msg(status []types.Status, j int) []*types.AcceptedRequest {
	var messages []*types.AcceptedRequest

	// Add to messages items that match status
	for item := rep[j].ReqQ.Front(); item != nil; item = item.Next() {
		element := item.Value.(*types.RequestStatus)
		for _, st := range status {
			if element.St == st {
				messages = append(messages, element.Req)
				break
			}
		}
	}
	return messages
}

/**
lastExec() returns the last request sequence number
executed by p[i].
*/
func lastExec() int {
	return rep[variables.Id].LastExec().Req.Sq
}

/**
Helper function. Not part of pseudocode
*/
func isExecutedByMajority(tuple *types.LogTuple) bool {
	count := 0
	for _, rs := range rep {
		for _, lt := range rs.RLog {
			if lt.Equals(tuple) { //TODO: Make sure comparison works as intended(probably not).
				count++
			}
		}
	}
	if count >= 3*variables.F+1 {
		return true
	}
	return false

}

/**
Returns the last request that p[i] sees locally to have been executed by
at least 3f + 1 processors.
TODO: Check if making any unnecessary loops.
*/
func LastCommonExec() *types.LogTuple {
	var max *types.LogTuple
	maxSq := 0
	for _, rs := range rep {
		if len(rs.RLog) < 1 {
			continue
		}
		for _, e := range rs.RLog {
			if e.Req.Sq > maxSq && isExecutedByMajority(e) {
				maxSq = e.Req.Sq
				max = e
			}
		}
	}
	return max
}

/**
Returns True if 4f + 1 processors report to have their conflict flag conFlag = True
*/
func conflict() bool {
	count := 0
	for _, rs := range rep {
		if rs.ConFlag {
			count++
		}
	}
	if count >= 4*variables.F+1 {
		return true
	}
	return false
}

/**
Returns a set of repStates that satisfy the types prefix relation <-/\/-> for
at least d processors. If no such exists then it returns nil.
*/
func comPrefStates(d int) []types.RepState {
	var states []types.RepState
	for i := range rep {
		states = make([]types.RepState, 0)
		for j := range rep {
			if rep[i].RepState.PrefixRelation(rep[j].RepState) {
				states = append(states, rep[j].RepState)
			}
		}
		if len(states) >= d {
			return states
		}
	}
	return make([]types.RepState, 0)
}

/**
getDsState() Returns a prefix suggested by 2f + 1 <= x < 3f + 1 processors,
with the requirement that there exist another y processors that have rep[]=⊥
such that x + y >= 4f + 1. If such a prefix doesn't exist it returns ⊥.
*/
func getDsState() types.RepState {
	var dState types.RepState
	var x, y []int
	for i := range rep {
		x = make([]int, 0)
		y = make([]int, 0)
		dState = make(types.RepState, 0)
		dState = append(dState, rep[i].RepState...)
		for j := range rep {
			if dState.PrefixRelation(rep[j].RepState) {
				x = append(x, j)
				dState = append(dState, dState.GetCommonPrefix(rep[j].RepState)...)
			} else {
				if rep[j].RepState == nil {
					y = append(y, j)
				}
			}
		}
		if 2*variables.F+1 <= len(x) && len(x) < 3*variables.F+1 &&
			len(x)+len(y) >= 4*variables.F+1 {
			return dState
		}

	}
	return make(types.RepState, 0)
}

/**
double() returns True if the reqQ of p[i] contains two copies of a request and they
have different view or sequence number.
*/
func double() bool {
	for e := rep[variables.Id].ReqQ.Front(); e != nil; e = e.Next() {
		reqI := e.Value.(*types.RequestStatus)
		for el := rep[variables.Id].ReqQ.Front(); el != nil; el = el.Next() {
			reqJ := el.Value.(*types.RequestStatus)
			if reqI == reqJ {
				continue
			}
			if reqI.Req.Equals(reqJ.Req) &&
				(!(reqI.Req.View == reqJ.Req.View) ||
					reqI.Req.Sq != reqJ.Req.Sq) {
				return true
			}
		}
	}
	return false
}

/**
staleReqSeqn() returns True if the sequence number has reached the maximal counter
value MAXINT
*/
func staleReqSeqn() bool {
	if (lastExec() + Sigma*variables.K) > MAXINT {
		return true
	}
	return false
}

/**
unsupReq()  returns True if a request exists in reqQ
less than 2f+1 times.
*/
func unsupReq() bool {
	if rep[variables.Id].ReqQ.Len() == 0 {
		return false
	}
	for e := rep[variables.Id].ReqQ.Front(); e != nil; e = e.Next() {
		item := e.Value.(*types.AcceptedRequest)
		count := 0
		for _, rs := range rep {
			reqs := rs.ReqQ
			for el := reqs.Front(); el != nil; el = el.Next() {
				req := el.Value.(*types.AcceptedRequest)
				if item.Equals(req) {
					count++
				}
			}
		}
	}
	return false
}

/**
staleRep() returns True if any of double(), unsupReq()
 or staleRep are True.
*/
func staleRep() bool {
	if staleReqSeqn() {
		return true
	}
	if unsupReq() {
		return true
	}
	if double() {
		return true
	}
	for _, log := range rep[variables.Id].RLog {
		if len(log.XSet) <= 3*variables.F+1 {
			return true
		}
	}
	return false
}

/**
knownPendReqs() returns a set of requests that appear
in the rep[i.pendReqs and also appear in the message
queues of at least another 3f + 1 processors
*/
func knownPendReqs() []*types.Request {
	var returnReqs []*types.Request
	pendReqs := rep[variables.Id].PendReqs
	for e := pendReqs.Front(); e != nil; e = e.Next() {
		req := e.Value.(*types.Request)
		count := 0
		for _, rs := range rep {
			//TODO: This condition might be unnecessary/wrong
			if rs == rep[variables.Id] {
				continue
			}
			var items []*types.Request
			for el := rs.PendReqs.Front(); el != nil; el = el.Next() {
				items = append(items, el.Value.(*types.RequestStatus).Req.Request)
			}
			for el := rs.ReqQ.Front(); el != nil; el = el.Next() {
				//TODO might need to check for duplicates
				items = append(items, el.Value.(*types.RequestStatus).Req.Request)
			}
			for _, item := range items {
				if req.Equals(item) {
					count++
				}
			}
		}
		if count >= 3*variables.F+1 {
			returnReqs = append(returnReqs, req)
		}
	}
	return returnReqs
}

/**
knownReqs(status t) returns a set of requests that
appear in the rep[i]reqQ and of at least another 3f + 1
processors and have a status in t.
*/
func knownReqs(set ...types.Status) []*types.RequestStatus {
	var retReqs []*types.RequestStatus
	reqs := rep[variables.Id].ReqQ
	for e := reqs.Front(); e != nil; e = e.Next() {
		req := e.Value.(*types.RequestStatus)
		statusFlag := false
		for _, t := range set {
			if req.St == t {
				statusFlag = true
			}
		}
		if statusFlag {
			continue
		}
		count := 0
		for _, rs := range rep {
			//if rs == rep[types.Id] {
			//	continue
			//}
			items := rs.ReqQ
			for el := items.Front(); el != nil; el = el.Next() {
				item := el.Value.(*types.RequestStatus)
				if req.Req.Equals(item.Req) {
					count++
					break
				}
			}
		}
		if count >= 3*variables.F+1 {
			retReqs = append(retReqs, req)
		}
	}
	return retReqs
}

/**
delayed() TODO: Check if LastCommonExec().Req.Sq is correct
*/
func delayed() bool {
	return lastExec() < (LastCommonExec().Req.Sq + 3*variables.K*Sigma)
}

/**
existsPPrepMsg(x, processor)
*/
func existsPPrepMsg(request *types.Request, processor int) bool {
	var statuses []types.Status
	statuses = append(statuses, types.PRE_PREP)
	messages := msg(statuses, processor)
	for _, msg := range messages {
		if msg.Request.Equals(request) {
			return true
		}
	}
	return false
}

/**
unassignedReqs() returns the set of pending requests
for which p[i] has neither seen a PRE-PREP message from
the types.Primary, nor has it seen 3f + 1 processors that have
a PREP message for the same client request.
*/
func unassignedReqs() []*types.Request {
	var returnReqs []*types.Request
	pendReqs := rep[variables.Id].PendReqs
	for e := pendReqs.Front(); e != nil; e = e.Next() {
		req := e.Value.(*types.Request)
		exists := false
		var st []types.Status
		st = append(st, types.PREP)
		st = append(st, types.COMMIT)
		set := knownReqs(st...)
		for _, rq := range set {
			if rq.Req.Request.Equals(req) {
				exists = true
				break
			}
		}
		if !existsPPrepMsg(req, variables.Prim) && !exists {
			returnReqs = append(returnReqs, req)
		}
	}
	return returnReqs
}

/**
acceptReqPPrep(x, types.Prim) returns True if there is a pre-prepare message
from the types.Primary types.Prim for a request x and the request
content is the same for 3f + 1 processors with the same
sequence number and view types.Identifier.
*/
func acceptReqPPrep(x *types.Request, prim int) bool {
	isKnownPendReq := types.ContainsRequest(x, knownPendReqs())
	if !isKnownPendReq {
		return false
	}
	reqs := rep[variables.Id].ReqQ
	for e := reqs.Front(); e != nil; e = e.Next() {
		req := e.Value.(*types.RequestStatus).Req
		isEqualReq := req.Request.Equals(x)
		if !isEqualReq {
			continue
		}
		existsPPrep := existsPPrepMsg(req.Request, prim)
		if !existsPPrep {
			continue
		}
		isPrim := req.View == prim
		if !isPrim {
			continue
		}
		lastExec := lastExec()
		isValidSq := req.Sq >= lastExec && req.Sq < lastExec+Sigma*variables.K
		if !isValidSq {
			continue
		}
		logSet := types.GetRequestsListFromLog(rep[variables.Id].RLog)
		logSet.PushBackList(reqs)
		for el := logSet.Front(); el != nil; el = el.Next() {
			var typeTest *types.RequestStatus
			if reflect.TypeOf(el.Value) == reflect.TypeOf(typeTest) {
				rq := el.Value.(*types.RequestStatus)
				if rq.Req.Sq == req.Sq && req.Request.Equals(rq.Req.Request) {
					continue
				}
			} else {
				rq := el.Value.(*types.AcceptedRequest)
				if rq.Sq == req.Sq && req.Request.Equals(rq.Request) {
					continue
				}
			}
		}
		return true
	}
	return false
}

/**
committedSet(x)
*/
func committedSet(x *types.AcceptedRequest) []*types.ReplicaStructure {
	var committedSet []*types.ReplicaStructure
	for i := range rep {
		if types.ContainsAcceptedRequest(x, msg([]types.Status{types.COMMIT}, i)) ||
			types.ContainsAcceptedRequest(x, types.GetRequestsFromLog(rep[i].RLog)) {
			committedSet = append(committedSet, rep[i])
		}
	}
	return committedSet
}

func ByzantineReplication() {
	go handleClientMessages()
	go handleReplicaMessages()
	for {
		/*
			Checking for View Change
		*/
		if viewChanged() && AllowService() {
			!viewChanged() = rep[variables.Id] != nil &&
				GetView(variables.Id) != variables.Prim
		}
		variables.Prim = GetView(variables.Id)
		if viewChanged() && variables.Prim == variables.Id {
			var processorSet []*types.ReplicaStructure
			for i, rs := range rep {
				if rs.ViewChanged &&
					variables.Prim == GetView(i) {
					processorSet = append(processorSet, rs)
				}
			}
			if len(processorSet) > 4*variables.F+1 {
				x := findConsState(comPrefStates(3*variables.F + 1))
				rep[variables.Id].RepState, rep[variables.Id].RLog = x.RepState, x.Rlog
				renewReqs()
				rep[variables.Id].ViewChanged = false
			}
		} else if viewChanged() &&
			rep[variables.Prim].ViewChanged == rep[variables.Id].ViewChanged &&
			rep[variables.Prim].Prim == rep[variables.Id].Prim &&
			checkNewVState(variables.Prim) &&
			countCommonPrimary() >= 4*variables.F+1 {

			rep[variables.Id] = rep[variables.Prim]
			rep[variables.Id].ViewChanged = false
		}
		/*
			Checking for state consistency
		*/
		x := findConsState(comPrefStates(3*variables.F + 1))
		y := getDsState()
		//TODO This is DEFINITELY NOT CORRECT
		if len(y) != 0 && len(x.RepState) == 0 {
			x.RepState = y
		}
		rep[variables.Id].ConFlag = len(x.RepState) == 0
		if !rep[variables.Id].ConFlag &&
			(!rep[variables.Id].RepState.PrefixRelation(x.RepState) ||
				rep[variables.Id].RepState.Equals(DEF_STATE().RepState) ||
				delayed()) {
			rep[variables.Id].RepState = x.RepState
		}
		if staleRep() || conflict() {
			flushLocal()
		}
		rep[variables.Id] = nil
		needFlush = true
		if flush {
			flushLocal()
		}
		for _, el := range knownPendReqs() {
			rep[variables.Id].Enqueue(el)
		}
		//Serviceable View and no conflicts
		if AllowService() && !needFlush {
			if NoViewChange() && viewChanged() {
				if variables.Prim == variables.Id {
					for req := rep[variables.Id].PendReqs.Front(); req != nil; req = req.Next() {
						if seqn < lastExec()+Sigma*variables.K {
							// 3-phase commit Replication
							reqSt := new(types.RequestStatus)
							reqSt.Req = new(types.AcceptedRequest)
							reqSt.Req.Request = req.Value.(*types.Request)
							reqSt.Req.View = variables.Prim
							seqn++
							reqSt.Req.Sq = seqn
							reqSt.St = types.PRE_PREP

							reqSt = new(types.RequestStatus)
							reqSt.Req = new(types.AcceptedRequest)
							reqSt.Req.Request = req.Value.(*types.Request)
							reqSt.Req.View = variables.Prim
							seqn++
							reqSt.Req.Sq = seqn
							reqSt.St = types.PREP
						}
					}
				} else {
					reqs := helper.ExcludeRequests(knownPendReqs(), unassignedReqs())
					reqs = helper.FilterRequests(reqs, func(request *types.Request) bool {
						flag := false
					outer:
						for _, replica := range rep {
							for reqSt := replica.ReqQ.Front(); reqSt != nil; reqSt = reqSt.Next() {
								if reqSt.Value.(*types.AcceptedRequest).Request.Equals(request) {
									flag = true
									break outer
								}
							}
						}
						return flag
					})
					for _, x := range reqs {
						if acceptReqPPrep(x, variables.Prim) {
							rep[variables.Id].Add(
								&types.RequestStatus{
									Req: &types.AcceptedRequest{Request: x},
									St:  types.PREP}) // TODO: Check how to add a simple request
							rep[variables.Id].Add(
								&types.RequestStatus{
									Req: &types.AcceptedRequest{Request: x},
									St:  types.PRE_PREP})
						}
					}
					for _, x := range knownReqs(types.PREP) {
						x.St = types.COMMIT
						rep[variables.Id].Remove(x.Req.Request)
					}
					//	if 3f + 1 commited then commit.
					// TODO: Union or Section
					reqQ := helper.ToSliceRequestStatus(rep[variables.Id].ReqQ)
					reqQ = helper.FilterRequestStatus(reqQ, func(rs *types.RequestStatus) bool {
						for _, x := range knownReqs(types.PREP, types.COMMIT) {
							if rs.Equals(x) {
								return false
							}
						}
						return true
					})
					for _, rs := range reqQ {
						x := committedSet(rs.Req)
						if len(x) > 3*variables.F+1 && rs.Req.Sq == lastExec()+1 { // TODO Check this part with Chryssis
							rr := new(types.RequestReply)
							rr.Req = rs.Req.Request
							rr.Client = rs.Req.Request.Client
							rr.Rep = apply(rs.Req.Request)
							rep[variables.Id].Enqueue(rr)
							ltuple := new(types.LogTuple)
							ltuple.Req = rs.Req
							ltuple.XSet = x
							rep[variables.Id].Add(ltuple)
							rep[variables.Id].Remove(rs.Req.Request)
							rep[variables.Id].Remove(rs.Req) // TODO: Make sure this works properly
						}
					}
				}
			}
		}
		//	Todo: Send Info
		for i := 0; i < variables.N; i++ {
			if i == variables.Id {
				continue
			}
			messenger.SendReplica(rep[variables.Id], i)
		}
		for cl := 0; cl < variables.K; cl++ {
			for _, lastReq := range rep[variables.Id].LastReq {
				if lastReq.Client == cl {
					messenger.ReplyClient(lastReq.Rep)
				}
			}
		}
	}
}

/*
TODO: Check wherever
if i == variables.Id {
	continue
}
if necessary
*/

func countCommonPrimary() int {
	count := 0
	for i := range rep {
		if GetView(i) == variables.Prim {
			count++
		}
	}
	return count
}

func handleReplicaMessages() {
	for {
		message := <-messenger.ReplicaChan
		j := message.From
		replica := message.Rep
		if AllowService() {
			if NoViewChange() {
				rep[j] = replica
			} else {
				rep[j].RepState = replica.RepState
			}
		}
	}
}

func handleClientMessages() {
	for {
		cm := <-messenger.RequestChan
		if cm.Ack {
			handleAck(cm)
		} else {
			handleRequest(cm.Req)
		}
	}
}

func handleRequest(req *types.Request) {
	if NoViewChange() && AllowService() {
		rep[variables.Id].Enqueue(req)
	}
}

func handleAck(cm *types.ClientMessage) {
	rep[variables.Id].Remove(&types.RequestReply{Req: cm.Req, Rep: nil})
}

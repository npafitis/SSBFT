package app

import (
	"SSBFT/app/messenger"
	"SSBFT/config"
	"SSBFT/logger"
	"SSBFT/types"
	"SSBFT/variables"
	"container/list"
	"errors"
	"github.com/jinzhu/copier"
	"log"
	"math"
	"reflect"
	"time"
)

/**
Constants
*/
var MAXINT uint64

/*****************/
// Identifier(i) of processor p[i]
//var types.Id int

var (
	// System-Defined Constant
	Sigma int
)

// Default/Incorruptible Fallback state

var seqn int // [0, MAXINT]
var rep []*types.ReplicaStructure
var needFlush bool
var flush bool

//var types.Prim int

func DEF_STATE() *types.ReplicaStructure {
	return &types.ReplicaStructure{RepState: make([]rune, 0), RLog: make([]*types.LogTuple, 0), PendReqs: list.New(), ReqQ: list.New(), LastReq: make([]*types.RequestReply, 0), ConFlag: false, ViewChanged: false, Prim: 0}
}

/*
Interface for BFT
*/

func InitializeReplication() {
	rep = make([]*types.ReplicaStructure, variables.N)
	for i := range rep {
		rep[i] = new(types.ReplicaStructure)
		replica := DEF_STATE()
		if variables.F == 0 {
			log.Fatal("F is 0")
		}
		if variables.Id != 0 && variables.Id%(variables.N/variables.F) == 0 {
			if config.TestCase == config.STALE_VIEWS {
				replica.Prim = 1
			}
			if config.TestCase == config.STALE_STATES && i == variables.Id {
				replica.RepState = []rune{'c', 'f'}
			}
		}
		rep[i] = replica
	}
	seqn = 0
	Sigma = 100
	MAXINT = uint64(math.Pow(2, 64))
	logger.OutLogger.Println("MaxInt =", MAXINT)
	needFlush = false
	flush = false
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
	var consState = struct {
		RepState types.RepState
		Rlog     []*types.LogTuple
	}{RepState: nil, Rlog: nil}
	if len(S) == 0 {
		consState.RepState = make([]rune, 0)
	} else {
		state := S[0]
		for i := range S {
			state = state.GetCommonPrefix(S[i])
		}
		consState.RepState = state
	}
	logs := types.RLog(rep[0].RLog)
	for i := range rep {
		logs = logs.CommonPrefix(rep[i].RLog)
	}
	consState.Rlog = logs

	return consState
}

/**
checkNewVstate() checks the state proposed by a newly installed primary after
view change. This involves checking whether the proposed pre-prepare messages of committed
processors are verified by another 3f+1 processors and the new state has a correct
prefix as per findConsState().
*/
func checkNewVState(id int) bool {
	consState := findConsState(comPrefStates(3*variables.F + 1))
	if !consState.RepState.PrefixRelation(rep[id].RepState) {
		return false
	}
	return true
}

/**
renewReqs() is executed by a new primary, in order to issue a consistent set
of pending requests messages for reqQ and pendReqs, where these are now allocated for execution
to the new view.
*/
func renewReqs() {
	pendReqs := list.New()
	reqQ := list.New()
	for e := rep[variables.Id].PendReqs.Front(); e != nil; e = e.Next() {
		req := e.Value.(*types.Request)
		count := 0
		for i := range rep {
			if types.PendReqsContains(rep[i].PendReqs, req) {
				count++
			}
		}
		if count >= 3*variables.F+1 {
			pendReqs.PushBack(req)
		}
	}
	for e := rep[variables.Id].ReqQ.Front(); e != nil; e = e.Next() {
		rs := e.Value.(*types.RequestStatus)
		count := 0
		for i := range rep {
			if types.ReqQContains(rep[i].ReqQ, rs) {
				count++
			}
		}
		if count >= 3*variables.F+1 {
			reqQ.PushBack(rs)
		}
	}
}

func apply(x *types.Request) *types.Reply {
	reply := new(types.Reply)
	switch x.Operation.Op {
	case types.ADD:
		rep[variables.Id].RepState = append(rep[variables.Id].RepState, x.Operation.Value)
		logger.OutLogger.Println(rep[variables.Id].RepState)
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
	logger.OutLogger.Println("Flushed Local.")
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
func msg(j int, status ...types.Status) []*types.AcceptedRequest {
	var messages []*types.AcceptedRequest
	//("Msg ReqQ len=",rep[j].ReqQ.Len())
	// Add to messages items that match status
	for item := rep[j].ReqQ.Front(); item != nil; item = item.Next() {
		element := item.Value.(*types.RequestStatus)
		for _, st := range status {
			//("element.St",element.St)
			//("st",st)
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
	if rep[variables.Id].LastExec() != nil {
		return rep[variables.Id].LastExec().Req.Sq
	}
	return 0
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
		if rs.ConFlag == true {
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
		states = append(states, rep[i].RepState)
		for j := range rep {
			if j == i {
				continue
			}
			flag := true
			for _, state := range states {
				if !rep[j].RepState.PrefixRelation(state) {
					flag = false
				}
			}
			if flag {
				states = append(states, rep[j].RepState)
			}
		}
		if len(states) >= d {
			return states
		}
	}
	//for i := range rep {
	//	states = make([]types.RepState, 0)
	//	for j := range rep {
	//		if rep[i].RepState.PrefixRelation(rep[j].RepState) {
	//			states = append(states, rep[j].RepState)
	//		}
	//	}
	//	if len(states) >= d {
	//		return states
	//	}
	//}
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
	if (uint64(lastExec() + Sigma*variables.K)) > MAXINT {
		return true
	}
	return false
}

/**
unsupReq()  returns True if a request exists in reqQ
less than 2f+1 times.
*/
//func unsupReq() bool {
//	if rep[variables.Id].ReqQ.Len() == 0 {
//		return false
//	}
//	for e := rep[variables.Id].ReqQ.Front(); e != nil; e = e.Next() {
//		item := e.Value.(*types.RequestStatus)
//		count := 0
//		for _, rs := range rep {
//			reqs := rs.ReqQ
//			for el := reqs.Front(); el != nil; el = el.Next() {
//				req := el.Value.(*types.AcceptedRequest)
//				if item.Equals(req) {
//					count++
//				}
//			}
//		}
//	}
//	return false
//}

/**
staleRep() returns True if any of double(), unsupReq()
 or staleRep are True.
*/
func staleRep() bool {
	if staleReqSeqn() {
		return true
	}
	//if unsupReq() {
	//	return true
	//}
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
in the rep[i].pendReqs and also appear in the message
queues of at least another 3f + 1 processors
*/
func knownPendReqs() []*types.Request {
	var returnReqs = make([]*types.Request, 0)
	for e := rep[variables.Id].PendReqs.Front(); e != nil; e = e.Next() {
		req := e.Value.(*types.Request)
		count := 0
	outer:
		for _, rs := range rep {
			if rs == rep[variables.Id] {
				continue
			}
			for element := rs.PendReqs.Front(); element != nil; element = element.Next() {
				request := element.Value.(*types.Request)
				if req.Equals(request) {
					count++
					continue outer
				}

			}
			for element := rs.ReqQ.Front(); element != nil; element = element.Next() {
				request := element.Value.(*types.RequestStatus)
				if req.Equals(request.Req.Request) {
					count++
					continue outer
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
		if !statusFlag {
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
	lastCommonExec := LastCommonExec()
	var sq int
	if lastCommonExec == nil {
		sq = 0
	} else {
		sq = lastCommonExec.Req.Sq
	}
	return lastExec() < (sq + 3*variables.K*Sigma)
}

/**
existsPPrepMsg(x, processor)
*/
func existsPPrepMsg(request *types.Request, processor int) bool {
	messages := msg(processor, types.PRE_PREP)
	//("msg() len=",len(messages))
	for _, msg := range messages {
		//("msg", msg)
		//("x", request)
		if msg.Request.Equals(request) {
			return true
		}
	}
	return false
}

/**
unassignedReqs() returns the set of pending requests
for which p[i] has neither seen a PRE-PREP message from
the primary, nor has it seen 3f + 1 processors that have
a PREP message for the same client request.
*/
func unassignedReqs() []*types.Request {
	var returnReqs []*types.Request
	pendReqs := rep[variables.Id].PendReqs
	for e := pendReqs.Front(); e != nil; e = e.Next() {
		req := e.Value.(*types.Request)
		existsPrep := false
		set := knownReqs(types.PREP, types.COMMIT)
		//if variables.Id == 1 {
		//
		//	("Set",len(set))
		//}
		for _, rq := range set {
			if rq.Req.Request.Equals(req) {
				existsPrep = true
				break
			}
		}
		existsPrim := existsPPrepMsg(req, rep[variables.Id].Prim)
		//for e := rep[rep[variables.Id].Prim].ReqQ.Front(); e != nil; e = e.Next() {
		//	if e.Value.(*types.RequestStatus).St == types.PRE_PREP {
		//		logger.OutLogger.Println("Something's fishy.")
		//	}
		//}
		logger.OutLogger.Println("existsPrim", existsPrim)
		if !existsPrim && !existsPrep {
			returnReqs = append(returnReqs, req)
		}
	}
	return returnReqs
}

/**
acceptPPrepReq(x, types.Prim) returns True if there is a pre-prepare message
from the primary Prim for a request x and the request
content is the same for 3f + 1 processors with the same
sequence number and view identifier.
*/
func acceptPPrepReq(x *types.Request, prim int) (bool, int) {

	isKnownPendReq := types.ContainsRequest(x, knownPendReqs())
	logger.OutLogger.Println("isKnownPendReq =", isKnownPendReq)
	if !isKnownPendReq {
		return false, 0
	}
	reqs := rep[prim].ReqQ
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
		return true, req.Sq
	}
	return false, 0
}

/**
committedSet(x)
*/
func committedSet(x *types.AcceptedRequest) []int {
	var committedSet []int
	for i := range rep {
		if types.ContainsAcceptedRequest(x, msg(i, types.COMMIT)) ||
			types.ContainsAcceptedRequest(x, types.GetRequestsFromLog(rep[i].RLog)) {
			committedSet = append(committedSet, i)
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
		if !viewChanged() && AllowService() {
			rep[variables.Id].ViewChanged =
				(rep[variables.Id] != nil) &&
					(GetView(variables.Id) != rep[variables.Id].Prim)
		}
		rep[variables.Id].Prim = GetView(variables.Id)
		if viewChanged() && rep[variables.Id].Prim == variables.Id {
			var processorSet []*types.ReplicaStructure
			for i, rs := range rep {
				if rs.ViewChanged &&
					rep[variables.Id].Prim == GetView(i) {
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
			rep[rep[variables.Id].Prim].ViewChanged == rep[variables.Id].ViewChanged &&
			rep[rep[variables.Id].Prim].Prim == rep[variables.Id].Prim &&
			checkNewVState(rep[variables.Id].Prim) &&
			countCommonPrimary() >= 4*variables.F+1 {

			rep[variables.Id] = rep[rep[variables.Id].Prim]
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
		//rep[variables.Id].ConFlag = len(x.RepState) == 0
		if !rep[variables.Id].ConFlag &&
			(!rep[variables.Id].RepState.PrefixRelation(x.RepState) ||
				rep[variables.Id].RepState.Equals(DEF_STATE().RepState) ||
				delayed()) {
			rep[variables.Id].RepState = x.RepState
		}
		if staleRep() || conflict() {
			logger.OutLogger.Println("Stale Rep = ", staleRep(), "Conflict =", conflict())
			flushLocal()
			//rep[variables.Id] = DEF_STATE()
			needFlush = true
		}
		if flush {
			logger.OutLogger.Println("Flush = ", flush)
			flushLocal()
		}
		for _, el := range knownPendReqs() {
			rep[variables.Id].Enqueue(el)
		}
		//Serviceable View and no conflicts
		if AllowService() && !needFlush {
			if NoViewChange() && !viewChanged() {
				if rep[variables.Id].Prim == variables.Id {
					if config.TestCase != config.BYZANTINE_PRIM && variables.Id == 0 {
						for req := rep[variables.Id].PendReqs.Front(); req != nil; req = req.Next() {
							request := req.Value.(*types.Request)
							isUnassigned := false
							for _, unassigned := range unassignedReqs() {
								if unassigned.Equals(request) {
									isUnassigned = true
								}
							}
							if !isUnassigned {
								continue
							}
							if seqn < lastExec()+Sigma*variables.K {
								// 3-phase commit Replication
								//("Seqn=",seqn)
								if config.TestCase == config.STALE_REQUESTS && variables.Id == 0 {
									request.Operation.Value = 'l'
								}
								reqSt := new(types.RequestStatus)
								reqSt.Req = new(types.AcceptedRequest)
								reqSt.Req.Request = request
								reqSt.Req.View = rep[variables.Id].Prim
								seqn++
								reqSt.Req.Sq = seqn
								reqSt.St = types.PRE_PREP
								rep[variables.Id].Add(reqSt)

								reqSt = new(types.RequestStatus)
								reqSt.Req = new(types.AcceptedRequest)
								reqSt.Req.Request = request
								reqSt.Req.View = rep[variables.Id].Prim
								//seqn++
								reqSt.Req.Sq = seqn
								reqSt.St = types.PREP
								rep[variables.Id].Add(reqSt)
							}
						}
					}
				} else {
					logger.OutLogger.Println("Accepting Sequence Numbers from Primary.")
					logger.OutLogger.Println("KnownPendReqs Len=", len(knownPendReqs()))
					logger.OutLogger.Println("UnassignedReqs Len=", len(unassignedReqs()))
					reqs := types.ExcludeRequests(knownPendReqs(), unassignedReqs())
					logger.OutLogger.Println("Examined Requests length =", len(reqs))
					reqs = types.FilterRequests(reqs, func(request *types.Request) bool {
						//for _, replica := range rep {
						replica := rep[variables.Id]
						for reqSt := replica.ReqQ.Front(); reqSt != nil; reqSt = reqSt.Next() {
							requestStatus := reqSt.Value.(*types.RequestStatus)
							//log.Println("this", requestStatus.Req.Sq)
							//log.Println("that", request.Operation.Value)
							if requestStatus.Req.Request.Equals(request) {
								return true
							}
						}
						//}
						return false

					})
					logger.OutLogger.Println("Examined Requests length =", len(reqs))
					for _, x := range reqs {
						logger.OutLogger.Println("Checking to Accept Request", x)
						acceptReq, sq := acceptPPrepReq(x, rep[variables.Id].Prim)
						if acceptReq {
							rep[variables.Id].Add(
								&types.RequestStatus{
									Req: &types.AcceptedRequest{Request: x, Sq: sq, View: rep[variables.Id].Prim},
									St:  types.PRE_PREP})
							rep[variables.Id].Add(
								&types.RequestStatus{
									Req: &types.AcceptedRequest{Request: x, Sq: sq, View: rep[variables.Id].Prim},
									St:  types.PREP}) // TODO: Check how to add a simple request
						}
					}
				}
				for _, x := range knownReqs(types.PREP) {
					x.St = types.COMMIT

					rep[variables.Id].Remove(x.Req.Request)
				}
				//	if 3f + 1 commited then commit.
				// TODO: Union or Section
				reqQ := types.ToSliceRequestStatus(rep[variables.Id].ReqQ)
				for _, log := range rep[variables.Id].RLog {
					reqQ = types.AppendIfMissingRequestStatus(reqQ, log.Req)
				}
				//reqQ = types.FilterRequestStatus(reqQ, func(rs *types.RequestStatus) bool {
				//	for _, logTuple := range rep[variables.Id].RLog{
				//		if logTuple.Req.Equals(rs.Req){
				//			return false
				//		}
				//	}
				//	return true
				//})

				reqQ = types.FilterRequestStatus(reqQ, func(rs *types.RequestStatus) bool {
					isKnownReq := false
					for _, x := range knownReqs(types.PREP, types.COMMIT) {
						if rs.Equals(x) {
							isKnownReq = true
							break
						}
					}
					return !isKnownReq
				})

				for _, rs := range reqQ {
					x := committedSet(rs.Req)
					//if variables.Id == 0 {
					//	log.Println("lastExec", rs.Req.Sq, lastExec()+1)
					//	log.Println("Is Committed", len(x), len(x) >= 3*variables.F+1)
					//}
					if len(x) >= 3*variables.F+1 && rs.Req.Sq == lastExec()+1 { // TODO Check this part with Chryssis
						if rs.St == types.PRE_PREP {
							log.Fatal("PRE_PREP shouldn't be in commitedSet")
						}
						rr := new(types.RequestReply)
						rr.Req = new(types.Request)
						err := copier.Copy(rr.Req, rs.Req.Request)
						if err != nil {
							logger.ErrLogger.Fatal(err)
						}
						rr.Client = rs.Req.Request.Client
						rr.Rep = new(types.Reply)
						err = copier.Copy(rr.Rep, apply(rs.Req.Request))
						if err != nil {
							logger.ErrLogger.Fatal(err)
						}
						rep[variables.Id].Enqueue(rr)

						ltuple := new(types.LogTuple)
						ltuple.Req = rs.Req
						ltuple.XSet = x
						rep[variables.Id].Enqueue(ltuple)
						rep[variables.Id].Remove(rs.Req.Request)
						rep[variables.Id].Remove(rs.Req) // TODO: Make sure this works properly
					}
				}
			}
		}
		for i := 0; i < variables.N; i++ {
			if i == variables.Id {
				continue
			}
			//if variables.Id==0{
			//	(rep[variables.Id].ReqQ.Len())
			//}
			messenger.SendReplica(rep[variables.Id], i)
		}
		logger.OutLogger.Println("Size of LastReq = ", len(rep[variables.Id].LastReq))
		logger.OutLogger.Println("Size of PendReqs = ", rep[variables.Id].PendReqs.Len())
		logger.OutLogger.Println("Size of ReqQ = ", rep[variables.Id].ReqQ.Len())

		//if variables.Id == 0 {
		//	log.Println("Size of LastReq = ", len(rep[variables.Id].LastReq))
		//}
		for i, lastReq := range rep[variables.Id].LastReq {
			messenger.ReplyClient(lastReq.Rep)
			rep[variables.Id].LastReq = append(rep[variables.Id].LastReq[:i], rep[variables.Id].LastReq[i+1:]...)
		}
	}
}

func countCommonPrimary() int {
	count := 0
	for i := range rep {
		if GetView(i) == rep[variables.Id].Prim {
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
				logger.OutLogger.Println("Updating replica of ", j, "with", replica)
				logger.OutLogger.Println("PendReqs Length =", replica.PendReqs.Len())
				logger.OutLogger.Println("ReqQ Length =", replica.ReqQ.Len())
				rep[j] = replica
			} else {
				logger.OutLogger.Println("Updating State of ", j, "with", replica.RepState)
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
	logger.OutLogger.Println("Ack Received.")
	//_, err := messenger.ServerSockets[cm.Req.Client].Send("", 0)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//messenger.SendRecvSync[cm.Req.Client] <- struct{}{}
	rep[variables.Id].Remove(&types.RequestReply{Req: cm.Req, Rep: nil})
	//log.Println("ReqQ", rep[variables.Id].ReqQ.Len())
	//log.Println("Sq", rep[variables.Id].ReqQ.Front().Value.(*types.RequestStatus).Req.Sq, "id", variables.Id, "lastExec", lastExec())
}

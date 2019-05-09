package helper

import "SSBFT/types"

func ExcludeRequests(src []*types.Request, target []*types.Request) []*types.Request {
	out := make([]*types.Request, 0)
	copy(out, src)
	for i, s := range out {
	inner:
		for _, r := range target {
			if s.Equals(r) {
				out = append(out[:i], out[i+1:]...)
				break inner
			}
		}
	}
	return out
}

func FilterRequests(src []*types.Request, filterFn func(request *types.Request) bool) []*types.Request {
	out := make([]*types.Request, 0)
	for _, s := range src {
		if !filterFn(s) {
			out = append(out, s)
		}
	}
	return out
}

func FilterRequestStatus(src []*types.RequestStatus,
	filterFn func(rs *types.RequestStatus) bool,
) []*types.RequestStatus {
	out := make([]*types.RequestStatus, 0)
	for _, s := range src {
		if !filterFn(s) {
			out = append(out, s)
		}
	}
	return out
}

func RLogEquals(rl1 []*types.LogTuple, rl2 []*types.LogTuple) bool {
	if len(rl1) != len(rl2) {
		return false
	}
	for i, log := range rl1 {
		if !log.Equals(rl2[i]) {
			return false
		}
	}
	return true
}

func LastReqEquals(lr1 []*types.RequestReply, lr2 []*types.RequestReply) bool {
	if len(lr1) != len(lr2) {
		return false
	}
	for i, log := range lr1 {
		if !log.Equals(lr2[i]) {
			return false
		}
	}
	return true
}

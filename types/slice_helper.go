package types

func ExcludeRequests(src []*Request, target []*Request) []*Request {
	out := make([]*Request, 0)
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

func FilterRequests(src []*Request, filterFn func(request *Request) bool) []*Request {
	out := make([]*Request, 0)
	for _, s := range src {
		if !filterFn(s) {
			out = append(out, s)
		}
	}
	return out
}

func FilterRequestStatus(src []*RequestStatus,
	filterFn func(rs *RequestStatus) bool,
) []*RequestStatus {
	out := make([]*RequestStatus, 0)
	for _, s := range src {
		if !filterFn(s) {
			out = append(out, s)
		}
	}
	return out
}

func RLogEquals(rl1 []*LogTuple, rl2 []*LogTuple) bool {
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

func LastReqEquals(lr1 []*RequestReply, lr2 []*RequestReply) bool {
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

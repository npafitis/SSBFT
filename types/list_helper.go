package types

import (
	"container/list"
)

func ListEquals(l1 *list.List, l2 *list.List) bool {
	if l1.Len() != l2.Len() {
		return false
	}
	for el1, el2 := l1.Front(), l2.Front(); el1 != nil || el2 != nil; {
		switch el1.Value.(type) {
		case *AcceptedRequest:
			req1 := el1.Value.(*AcceptedRequest)
			req2 := el2.Value.(*AcceptedRequest)
			if !req1.Equals(req2) {
				return false
			}
			break
		case *Request:
			req1 := el1.Value.(*Request)
			req2 := el2.Value.(*Request)
			if !req1.Equals(req2) {
				return false
			}
			break
		}
	}
	return true
}

func ToSliceRequestStatus(list *list.List) []*RequestStatus {
	res := make([]*RequestStatus, 0)
	for e := list.Front(); e != nil; e = e.Next() {
		res = append(res, e.Value.(*RequestStatus))
	}
	return res
}

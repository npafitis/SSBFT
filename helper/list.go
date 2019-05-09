package helper

import (
	"SSBFT/types"
	"container/list"
)

func ListEquals(l1 *list.List, l2 *list.List) bool {
	if l1.Len() != l2.Len() {
		return false
	}
	for el1, el2 := l1.Front(), l2.Front(); el1 != nil || el2 != nil; {
		switch el1.Value.(type) {
		case *types.AcceptedRequest:
			req1 := el1.Value.(*types.AcceptedRequest)
			req2 := el2.Value.(*types.AcceptedRequest)
			if !req1.Equals(req2){
				return false
			}
			break
		case *types.Request:
			req1 := el1.Value.(*types.Request)
			req2 := el2.Value.(*types.Request)
			if !req1.Equals(req2){
				return false
			}
			break
		}
	}
	return true
}

func ToSliceRequestStatus(list *list.List) []*types.RequestStatus{
	res := make([]*types.RequestStatus,0)
	for e:=list.Front();e!=nil;e=e.Next(){
		res = append(res, e.Value.(*types.RequestStatus))
	}
	return res
}

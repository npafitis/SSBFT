package types

import "math"

type RepState []rune

func (rs RepState) Equals(state RepState) bool {
	if len(rs)!= len(state){
		return false
	}
	for e := range rs {
		if rs[e] != state[e]{
			return false
		}
	}
	return true
}

func (a RepState) PrefixRelation(b RepState) bool {
	var size = int(math.Min(float64(len(a)), float64(len(b))))
	for i := 0; i < size; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (a RepState) GetCommonPrefix(b RepState) RepState {
	var prefix = make(RepState, 0)
	var size = int(math.Min(float64(len(a)), float64(len(b))))
	for i := 0; i < size; i++ {
		if a[i] == b[i] {
			prefix = append(prefix, a[i])
		}
	}
	return prefix
}

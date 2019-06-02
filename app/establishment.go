package app

import (
	"SSBFT/config"
	"SSBFT/logger"
	"SSBFT/types"
	"SSBFT/variables"
	"log"
)

/*
Variables for View Establishment
*/
var views []types.VPair //= make([]types.VPair, variables.N)

var vChange []bool // = make([]bool, variables.N)

func InitializeEstablishment() {
	views = make([]types.VPair, variables.N)
	vChange = make([]bool, variables.N)
}

/*
Constants for View Establishment
*/
//TODO [ ]: values might not be correct
const DF_VIEW_CUR int = 0

const DF_VIEW_NEXT = 0

func RstPair() types.VPair {
	return types.VPair{Cur: DF_VIEW_CUR, Next: DF_VIEW_NEXT}
}

/**
Aliases for View Establishment
*/
func vp(j int) types.VPair {
	return views[j]
}

func phase(j int) types.Phase {
	return GetPhs(j)
}

func NeedReset() bool {
	return staleV(variables.Id) || GetFlush() //TODO [ ]: GetFlush might be wrong
}

func ResetAll() string {
	views[variables.Id] = RstPair()
	AutomatonInit()
	RepRequestReset()
	return "Reset"
}

func ViewChange() {
	logger.OutLogger.Println("Change View")
	//log.Println("Change View", variables.Id)
	vChange[variables.Id] = true
}

func AllowService() bool {
	if config.TestCase == config.NON_SS {
		return true
	}
	return ((len(sameVSet(vp(variables.Id), types.ONE)) +
		len(sameVSet(vp(variables.Id), types.ZERO))) >= 3*variables.F+1) &&
		(phase(variables.Id) == types.ZERO && vp(variables.Id).Cur == vp(variables.Id).Next)
}

func GetView(j int) int {
	if j == variables.Id && phase(variables.Id) == 0 && witnessSeen() {
		if AllowService() {
			return vp(variables.Id).Cur
		} else {
			return -1
		}
	}
	return vp(j).Cur
}

func Automaton(t types.Type, p types.Phase, c int) (val bool, action string) {
	val = false
	switch {
	case t == types.PRED && p == types.ZERO && c == 0:
		for i, v := range views {
			if transitAdopble(i, 0, types.Follow) {
				val = true
				if (v.Next != vp(variables.Id).Next) &&
					v.Next != vp(variables.Id).Cur &&
					val {
					return true, "Adopt this View"
				}
			}
		}
		return false, ""

	case t == types.ACT && p == types.ZERO && c == 0:
		//log.Println("ID:", variables.Id, "Adopt View")
		adopt(vp(variables.Id))
		resetVChange()
		nextPhs()
		break
	case t == types.PRED && p == types.ZERO && c == 1:
		val = changeable() && (establishable(types.ZERO, types.Follow) && vChange[variables.Id])
		//log.Println("ID:", variables.Id, "changeable", changeable(), "establishable", establishable(types.ZERO, types.Follow) && vChange[variables.Id])
		break
	case t == types.ACT && p == types.ZERO && c == 1:
		nextView()
		nextPhs()
		resetVChange()
		break
	case t == types.PRED && p == types.ZERO && c == 2:
		val = transitAdopble(variables.Id, types.ZERO, types.Remain) || vp(variables.Id).Equals(RstPair())
		break
	case t == types.ACT && p == types.ZERO && c == 2:
		//log.Println("ID:",variables.Id,"No Action Phase Zero")
		action = "No Action"
		break
	case t == types.PRED && p == types.ZERO && c == 3:
		val = true
		break
	case t == types.ACT && p == types.ZERO && c == 3:
		//log.Println("ID:", variables.Id, "Reset All Phase Zero")
		action = ResetAll()
		resetVChange()
		break
	case t == types.PRED && p == types.ONE && c == 0:
		for i := range views {
			val = transitAdopble(i, types.ONE, types.Follow)
		}
		val = val && (vp(variables.Id).Cur != vp(variables.Id).Next)
		break
	case t == types.ACT && p == types.ONE && c == 0:
		adopt(vp(variables.Id))
		resetVChange()
		break
	case t == types.PRED && p == types.ONE && c == 1:
		val = establishable(p, types.Follow)
		//log.Println("ID:", variables.Id, "establish view pred", val)
		break
	case t == types.ACT && p == types.ONE && c == 1:
		log.Println("ID:", variables.Id, "Establish Act")
		if vp(variables.Id).Equals(RstPair()) {
			ReplicaFlush()
		}
		establish()
		nextPhs()
		resetVChange()
		break
	case t == types.PRED && p == types.ONE && c == 2:
		val = transitAdopble(variables.Id, p, types.Remain)
		break
	case t == types.ACT && p == types.ONE && c == 2:
		//log.Println("ID:", variables.Id, "No Action Phase One")
		action = "No Action"
		break
	case t == types.PRED && p == types.ONE && c == 3:
		val = true
		break
	case t == types.ACT && p == types.ONE && c == 3:
		//log.Println("ID:", variables.Id, "Reset All Phase One")
		action = ResetAll()
		resetVChange()
		break
	}
	return val, action
}

func AutoMaxCase(p types.Phase) int {
	switch p {
	case types.ZERO:
		return 3
	case types.ONE:
		return 3
	}
	return 3
}

func GetInfo(k int) *types.ViewVChange {
	return &types.ViewVChange{View: types.VPair{Cur: views[k].Cur, Next: views[k].Next}, ViewChange: vChange[k]}
}

func SetInfo(m *types.CoordinationMessage, k int) {
	views[k] = m.Views[k]
}

/*
Macros for View Establishment Algorithm
*/

func legitPhsZero(vp types.VPair) bool {
	return (vp.Cur == vp.Next || vp.Equals(RstPair())) &&
		typeCheck(vp)
}

func legitPhsOne(vp types.VPair) bool {
	return vp.Cur != vp.Next && typeCheck(vp)
}

func typeCheck(vp types.VPair) bool {
	return ((vp.Cur >= 0 && vp.Cur < variables.N) || vp.Cur == -1) &&
		((vp.Next >= 0 && vp.Next < variables.N) || vp.Cur == -1)
	// && vp.Next != -1 && vp.Cur != -1
}

func staleV(k int) bool {
	return (GetPhs(k) == 0 && !legitPhsZero(views[k])) ||
		(phase(k) == 1 && !legitPhsOne(views[k]))
}

func valid(m *types.CoordinationMessage, k int) bool {
	return (m.Phase == 0 && legitPhsZero(m.Views[k])) ||
		(m.Phase == 1 && legitPhsOne(m.Views[k]))
}

func sameVSet(j types.VPair, ph types.Phase) []int {
	var processorSet []int
	for i := 0; i < variables.N; i++ {
		if phase(i) == ph && vp(i).Equals(j) && !staleV(i) {
			processorSet = append(processorSet, i)
		}
	}
	return processorSet
}

func setContains(arr []int, k int) bool {
	for _, el := range arr {
		if k == el {
			return true
		}
	}
	return false
}

//Helper functon Not part of algorithm
func mergeSets(arr1 []int, arr2 []int) []int {
	var set []int
	copy(arr1, set)
	for i := range arr2 {
		set = types.AppendIfMissingInt(set, arr2[i])
	}
	//set := make([]int, 0)
	//set = append(set, arr1...)
	//for _, el := range arr2 {
	//	if !setContains(arr1, el) {
	//		set = append(set, el)
	//	}
	//}
	return set
}

func transitAdopble(j int, ph types.Phase, d types.Mode) bool {
	var set = mergeSets(sameVSet(vp(j), ph), transitSet(j, ph, d))
	return len(set) >= 3*variables.F+1
}

func transitSet(j int, ph types.Phase, d types.Mode) []int {
	var set = make([]int, 0)
	for i := 0; i < variables.N; i++ {
		if phase(i) != ph && transitionCases(j, vp(i), ph, d) && !staleV(i) {
			set = append(set, i)
		}
	}
	return set
}

func transitionCases(j int, vPair types.VPair, ph types.Phase, t types.Mode) bool {
	switch t {
	case types.Remain:
		return vPair.Next == vp(j).Cur
	case types.Follow:
		switch ph {
		case 0:
			return vPair.Next == (vp(j).Cur+1)%variables.N //TODO [ ]: Parentheses might be wrong
		case 1:
			return vPair.Cur == vp(j).Next
		}
	}
	return false
}

func adopt(vPair types.VPair) {
	views[variables.Id].Next = vPair.Cur
}

func establishable(ph types.Phase, mode types.Mode) bool {
	vSet := sameVSet(vp(variables.Id), phase(variables.Id))
	tranSet := transitSet(variables.Id, ph, mode)
	//log.Println("ID:",variables.Id,"vSet", vSet)
	//log.Println("ID:",variables.Id,"tranSet", tranSet)
	return len(vSet)+len(tranSet) >= 4*variables.F+1
}

func changeable() bool {
	setZero := make([]int, 0)
	setOne := make([]int, 0)
	for k := 0; k < variables.N; k++ {
		if phase(k) == types.ZERO &&
			vChange[k] == true &&
			vp(k).Equals(vp(variables.Id)) {
			setZero = types.AppendIfMissingInt(setZero, k)
		} else if phase(k) == types.ONE &&
			views[k].Next == (views[k].Cur+1)%variables.N {
			setOne = types.AppendIfMissingInt(setOne, k)
		}
	}
	return len(setZero)+len(setOne) >= 4*variables.F+1
}

func establish() {
	logger.OutLogger.Println("Establish View")
	log.Println("ID:", variables.Id, "Establish View")
	views[variables.Id].Cur = vp(variables.Id).Next
}

func nextView() {
	//log.Println("ID:", variables.Id, "Next View")
	views[variables.Id].Next = (views[variables.Id].Cur + 1) % variables.N
}

func resetVChange() {
	//log.Println("ID:", variables.Id, "Reset vChange")
	vChange[variables.Id] = false
}

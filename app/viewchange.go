package app

import (
	"SSBFT/app/messenger"
	"SSBFT/config"
	"SSBFT/logger"
	"SSBFT/types"
	"SSBFT/variables"
	"bytes"
	"encoding/gob"
)

var vcm []*types.VCM

func InitializeViewChange() {
	vcm = make([]*types.VCM, variables.N)
	for i := range vcm {
		vcm[i] = new(types.VCM)
	}
}

func defState() *types.VCM {
	return &types.VCM{VStatus: types.OK, Prim: GetView(variables.Id), NeedChange: false, NeedChgSet: make([]int, 0)}
}

/*
Interface for View Change
*/
func NoViewChange() bool {
	if config.TestCase == config.NON_SS {
		return true
	}
	return vcm[variables.Id].VStatus == types.OK
}

func ViewChangeMonitor() {
	go handleVCM()
	for {
		if vcm[variables.Id].Prim != GetView(variables.Id) {
			cleanState()
		}
		vcm[variables.Id].Prim = GetView(variables.Id)
		vcm[variables.Id].NeedChange = Suspected()
		if AllowService() {
			if vcm[variables.Id].Prim == GetView(variables.Id) &&
				vcm[variables.Id].VStatus != types.VChange {
				var set []int
				for i := 0; i < variables.N; i++ {
					if GetView(variables.Id) == GetView(i) &&
						vcm[i].NeedChange {
						set = append(set, i)
					}
				}
				for i := range set {
					vcm[variables.Id].NeedChgSet = types.AppendIfMissingInt(vcm[variables.Id].NeedChgSet, set[i])
				}
				//log.Println("needChgSet len",len(vcm[variables.Id].NeedChgSet))
				//vcm[variables.Id].NeedChgSet = append(vcm[variables.Id].NeedChgSet, set...)
				noServiceCount := 0
				for i := range vcm {
					if vcm[i].VStatus == types.NoService {
						noServiceCount++
					}
				}
				if noServiceCount < 2*variables.F+1 {
					vcm[variables.Id].VStatus = types.OK
				}
				if vcm[variables.Id].VStatus == types.OK &&
					supChange(3*variables.F+1) {
					vcm[variables.Id].VStatus = types.NoService
				} else if supChange(4*variables.F + 1) {
					vcm[variables.Id].VStatus = types.VChange
					ViewChange()
				}
			} else if vcm[variables.Id].Prim == GetView(variables.Id) &&
				vcm[variables.Id].VStatus == types.VChange {
				ViewChange()
			} else {
				cleanState()
			}
		}
		w := new(bytes.Buffer)
		encoder := gob.NewEncoder(w)
		err := encoder.Encode(vcm[variables.Id])
		if err != nil {
			logger.ErrLogger.Fatal(encoder)
		}
		message := types.Message{From: variables.Id, Payload: w.Bytes(), Type: "VCM"}
		messenger.Broadcast(message)
	}
}

/*
Macros for View Change
*/

func cleanState() {
	for i := range vcm {
		vcm[i] = defState()
	}
}

// TODO check correctness here
/*
return True if there is a set of processors size x with the same primary, and this set supports a view
change, and also each member of the set sees an intersection of needChgSet sets of size at least 3f + 1
*/
func supChange(x int) bool {
	for i := 0; i < variables.N; i++ {
		var set []int
		for j := 0; j < variables.N; j++ {
			if vcm[i].Prim == vcm[j].Prim {
				set = append(set, j)
			}
		}
		if len(set) == 0 {
			continue
		}
		var intersection []int
		intersection = vcm[set[0]].NeedChgSet
		for _, j := range set {
			intersection = types.IntersectionInt(intersection, vcm[j].NeedChgSet)
		}
		if len(intersection) >= 3*variables.F+1 && len(set) >= x {
			return true
		}
	}
	return false
}

func handleVCM() {
	for {
		msg := <-messenger.VcmChan
		from := msg.From
		*vcm[from] = msg.Vcm
	}
}

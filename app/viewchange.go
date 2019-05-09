package app

import (
	"SSBFT/app/messenger"
	"SSBFT/logger"
	"SSBFT/types"
	"SSBFT/variables"
	"bytes"
	"encoding/gob"
)

var vcm = make([]*types.VCM, variables.N)

func defState() *types.VCM {
	return &types.VCM{VStatus: types.OK, Prim: GetView(variables.Id), NeedChange: false, NeedChgSet: make([]int, 0)}
}

/*
Interface for View Change
*/
func NoViewChange() bool {
	return vcm[variables.Id].VStatus == types.OK
}

func ViewChangeMonitor() {
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
				vcm[variables.Id].NeedChgSet = append(vcm[variables.Id].NeedChgSet, set...)
				noServiceCount := 0
				for i := range vcm {
					if vcm[i].VStatus == types.OK {
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

func supChange(x int) bool {
	for i := 0; i < variables.N; i++ {
		var set []int
		for j := 0; j < variables.N; j++ {
			if vcm[i].Prim == vcm[j].Prim {
				set = append(set, j)
			}
		}
		pop := 0
		for j := range set {
			pop += len(vcm[j].NeedChgSet)
		}
		if pop >= 3*variables.F+1 && len(set) >= x {
			return true
		}
	}
	return false
}

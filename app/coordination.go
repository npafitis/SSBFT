package app

import (
	"SSBFT/app/messenger"
	"SSBFT/logger"
	"SSBFT/types"
	"SSBFT/variables"
	"bytes"
	"encoding/gob"
)

/**
Coordinating Automaton
*/

/*
Variables for Coordinating Automaton
*/

// Array of phase in {0, 1}
var phs []types.Phase

//witnesses[i] refers to the case where Pi observes that 4f+1
//processors had noticed the most recent value of getInfo(i).
var witnesses []bool

// Set of processors Pj for which Pi received witnesses[j] = true
var witnessSet []int

var echo []*types.AutomatonInfo

/*
Macros for Coordinating Automaton
*/
/**
echoNoWitn(k) checks whether the view and phase that
processor Pk reported about Pi match Pi's view and phase.
*/
func echoNoWitn(k int) bool {
	return vp(variables.Id).Equals(echo[k].View) &&
		phs[variables.Id] == echo[k].Phase &&
		vChange[variables.Id] == echo[k].VChange
}

/*
witnessSeen() is correct when the witnessSet set
is of size greater than 4f + 1
*/
func witnessSeen() bool {
	//log.Println(variables.Id, "witnesses[i]", witnesses[variables.Id])
	return witnesses[variables.Id] && (len(witnessSet)+1) >= 4*variables.F+1
}

/*
nextPhs() proceeds the phase from 0 to 1 and from 1 to
0, also emptying the witnessSet set.
*/
func nextPhs() {
	phs[variables.Id] = (phs[variables.Id] + 1) % 2
	witnesses[variables.Id] = false
	witnessSet = make([]int, 0)
}

/**
Interface Functions for Coordinating Automaton
*/

/*
getPhs(k) returns phsi[k] if called by Pi
*/
func GetPhs(k int) types.Phase {
	return phs[k]
}

/*
init() resets the variables of the coordinating automaton
algorithm to default values
*/
func AutomatonInit() {
	witnessSet = make([]int, 0)
	for j := range phs {
		phs[j] = 0
		witnesses[j] = false
	}
}

func CoordinatingAutomaton() {
	//go handleCoordination()
	for {
		if NeedReset() {
			ResetAll()
		}
		count := 0
		for i := 0; i < variables.N; i++ {
			if echoNoWitn(i) {
				count++
			}
		}
		//log.Println("Count", count)
		witnesses[variables.Id] = count >= 4*variables.F+1
		for i := range witnesses {
			if witnesses[i] {
				witnessSet = types.AppendIfMissingInt(witnessSet, i)
			}
		}
		witnesSeen := witnessSeen()
		//log.Println("id",variables.Id,"vChange", vChange[variables.Id], "witnesSeen", witnesSeen, "|witnesSet|", len(witnessSet))
		if witnesSeen {
			// TODO Dame kati pezei
			c := 0
			automaton, _ := Automaton(types.PRED, phase(variables.Id), c)

			for !automaton && AutoMaxCase(phase(variables.Id)) >= c {
				c++
				automaton, _ = Automaton(types.PRED, phase(variables.Id), c)
			}
			if AutoMaxCase(phase(variables.Id)) >= c {
				_, ret := Automaton(types.ACT, phase(variables.Id), c)

				if ret != "No Action" && ret != "Reset" {
					nextPhs()
				}
			}
		}
		for i := 0; i < variables.N; i++ {
			if i == variables.Id {
				continue
			}
			this := &types.CoordinationMessage{
				Phase:       phase(variables.Id),
				Witness:     witnesses[variables.Id],
				ViewVChange: GetInfo(variables.Id),
				Views:       views, // TODO This is not written in the algorithm description but might be necessary
				LastReport: &types.LastReport{
					Phase:   phase(i),
					Witness: witnesses[i],
					Pair:    GetInfo(i),
				},
			}
			w := new(bytes.Buffer)
			encoder := gob.NewEncoder(w)
			err := encoder.Encode(this)
			if err != nil {
				logger.ErrLogger.Fatal(err)
			}
			message := types.Message{Payload: w.Bytes(), Type: "CoordinationMessage", From: variables.Id}
			//log.Println("Processor", variables.Id, "sending coordination to", i, "Message", this)
			messenger.SendMessage(message, i)
		}
		handleCoordination()
	}
}

func handleCoordination() {
	//for {
	message := <-messenger.CoordChan
	if valid(message.Message, message.From) {
		phs[message.From] = message.Message.Phase
		witnesses[message.From] = message.Message.Witness
		echo[message.From] = &types.AutomatonInfo{
			Phase:   message.Message.Phase,
			Witness: message.Message.Witness,
			View:    message.Message.ViewVChange.View,
			VChange: message.Message.ViewVChange.ViewChange,
		}
		SetInfo(message.Message, message.From)

	}
	//}
}

func InitializeAutomaton() {
	phs = make([]types.Phase, variables.N)
	witnesses = make([]bool, variables.N)
	echo = make([]*types.AutomatonInfo, variables.N)
	for i := range echo {
		echo[i] = &types.AutomatonInfo{
			View:    types.VPair{Cur: 0, Next: 1},
			Phase:   types.ZERO,
			VChange: false,
			Witness: false}
	}
	witnessSet = make([]int, 0)
}

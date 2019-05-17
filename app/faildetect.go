package app

import (
	"SSBFT/app/messenger"
	"SSBFT/types"
	"SSBFT/variables"
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

var beat []int // = make(, variables.N)

var FDSet []int

var cnt []int // = make([]int, variables.N)

var primSusp []bool //= make([]bool, variables.N)

var curCheckReq []*types.Request

func InitializeFailureDetector() {
	beat = make([]int, variables.N)
	FDSet = make([]int, 0)
	cnt = make([]int, variables.N)
	primSusp = make([]bool, variables.N)
	curCheckReq = make([]*types.Request, variables.N)
}

/*
Interface for Failure Detector
*/

func Suspected() bool {
	var count = 0
	for i := 0; i < variables.N; i++ {
		if GetView(i) == GetView(variables.Id) &&
			primSusp[i] == true {
			count++
		}
	}
	return count >= 3*variables.F+1
}

/*
Macros for Failure Detector
*/
func reset() {
	for i := 0; i < variables.N; i++ {
		primSusp[i] = false
		beat[i] = 0
		cnt[i] = 0
	}
	curCheckReq = make([]*types.Request, 0)
}

func Monitor() {
	go handleToken()
	go func() {
		for {
			time.Sleep(time.Second)
			token := new(types.Token)
			token.FDSet = 0
			token.PrimSusp = primSusp[variables.Id]
			message := new(types.Message)
			w := new(bytes.Buffer)
			encoder := gob.NewEncoder(w)
			err := encoder.Encode(token)
			if err != nil {
				log.Fatal(err)
			}
			message.Payload = w.Bytes()
			message.Type = "Token"
			message.From = variables.Id
			messenger.Broadcast(*message)
		}
	}()
}

func handleToken() {
	for {
		msg := <-messenger.TokenChan
		token := msg.Token
		j := msg.From
		// Responsiveness Check
		beat[variables.Id] = 0
		beat[j] = beat[variables.Id]
		for i := 0; i < variables.N; i++ {
			if i == j || i == variables.Id {
				continue
			}
			beat[i]++
		}
		set := make([]int, 0)
		for i := 0; i < variables.N; i++ {
			if beat[i] < variables.T {
				set = append(set, i)
			}
		}
		FDSet = set

		//	Primary Progress Check
		if variables.Prim != GetView(variables.Id) {
			reset()
		}
		variables.Prim = GetView(variables.Id)
		if AllowService() && NoViewChange() {
			if j == variables.Prim {
				isNotPending := false
				pendReqs, _ := GetPendReqs()
				if len(curCheckReq) == 0 {
					isNotPending = true
				}
			loop:
				for _, req := range curCheckReq {
					for _, r := range pendReqs {
						if req.Equals(r) {
							isNotPending = true
							break loop
						}
					}
				}
				if isNotPending {
					cnt[j], curCheckReq = 0, pendReqs
				} else {
					cnt[variables.Id]++
				}
			} else if variables.Prim == GetView(j) {
				primSusp[j] = token.PrimSusp
			}
			for i := range cnt {
				if i == variables.Prim {
					continue
				}
				cnt[i] = 0
			}
			if variables.Id == variables.Prim {
				cnt[variables.Id] = 0
			}
			if !primSusp[variables.Id] {
				primIsFD := false
				for i := range FDSet {
					if FDSet[i] == variables.Prim {
						primIsFD = true
					}
				}
				primSusp[variables.Id] = !primIsFD && cnt[variables.Id] > variables.T
			}
		} else if !AllowService() {
			reset()
		}
	}
}

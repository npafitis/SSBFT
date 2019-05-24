package config

import (
	"SSBFT/variables"
	"strconv"
)

var RepAddresses map[string]int
var ReqAddresses map[string]int
var ServerAddresses map[string]int
var ResponseAddresses map[string]int

func InitialiseIP(n int) {
	RepAddresses = make(map[string]int, n)
	ReqAddresses = make(map[string]int, n)
	ServerAddresses = make(map[string]int, variables.K)
	ResponseAddresses = make(map[string]int, variables.K)
	for i := 0; i < n; i++ {
		RepAddresses["tcp://localhost:"+strconv.Itoa(4000+variables.Id*100+i)] = i
		ReqAddresses["tcp://localhost:"+strconv.Itoa(4000+i*100+variables.Id)] = i
	}
	for i := 0; i < variables.K; i++ {
		ServerAddresses["tcp://*:"+strconv.Itoa(7000+variables.Id*100+i)] = i
		ResponseAddresses["tcp://*:"+strconv.Itoa(10000+variables.Id*100+i)] = i
	}
}

func GetRepAddress(id int) string {
	for key, value := range RepAddresses {
		if value == id {
			return key
		}
	}
	return ""
}

func GetResponseAddress(id int) string {
	for key, value := range ResponseAddresses {
		if value == id {
			return key
		}
	}
	return ""
}
func GetServerAddress(id int) string {
	for key, value := range ServerAddresses {
		if value == id {
			return key
		}
	}
	return ""
}

func GetReqAddress(id int) string {
	for key, value := range ReqAddresses {
		if value == id {
			return key
		}
	}
	return ""
}

package config

import (
	"SSBFT/variables"
	"strconv"
)

var addresses = []string{
	"192.36.94.2",
	"141.22.213.35",
	"139.30.241.191",
	"132.227.123.14",
	"129.242.19.196",
	"141.24.249.131",
	"130.192.157.138",
	"141.22.213.34",
	"192.33.193.18",
	"192.33.193.16",
	"131.246.19.201",
	"155.185.54.249",
	"128.232.103.202",
	"195.251.248.180",
	"194.42.17.164",
	"128.232.103.201",
	"193.1.201.27",
	"193.226.19.30",
	"132.65.240.103",
	"193.1.201.26",
	"129.16.20.70",
	"129.16.20.71",
	"195.113.161.13",
}

var RepAddressesIp map[string]int
var ReqAddressesIp map[string]int
var ServerAddressesIp map[string]int
var ResponseAddressesIp map[string]int

func InitialiseIp(n int) {
	RepAddressesIp = make(map[string]int, n)
	ReqAddressesIp = make(map[string]int, n)
	ServerAddressesIp = make(map[string]int, variables.K)
	ResponseAddressesIp = make(map[string]int, variables.K)
	for i := 0; i < n; i++ {
		RepAddressesIp["tcp://*:"+strconv.Itoa(4000+i)] = i
		ReqAddressesIp["tcp://"+addresses[i]+":"+strconv.Itoa(4000+i)] = i
	}
	for i := 0; i < variables.K; i++ {
		ServerAddressesIp["tcp://*:"+strconv.Itoa(7000+variables.Id*100+i)] = i
		ResponseAddressesIp["tcp://*:"+strconv.Itoa(10000+variables.Id*100+i)] = i
	}
}

func GetRepAddress(id int) string {
	for key, value := range RepAddressesIp {
		if value == id {
			return key
		}
	}
	return ""
}

func GetResponseAddress(id int) string {
	for key, value := range ResponseAddressesIp {
		if value == id {
			return key
		}
	}
	return ""
}
func GetServerAddress(id int) string {
	for key, value := range ServerAddressesIp {
		if value == id {
			return key
		}
	}
	return ""
}

func GetReqAddress(id int) string {
	for key, value := range ReqAddressesIp {
		if value == id {
			return key
		}
	}
	return ""
}

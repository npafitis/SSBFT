package messenger

import (
	"SSBFT/config"
	"SSBFT/logger"
	"SSBFT/types"
	"SSBFT/variables"
	"bytes"
	"encoding/gob"
	"github.com/pebbe/zmq4"
	"strings"
)

var (
	Context *zmq4.Context

	RcvSockets map[int]*zmq4.Socket

	SndSockets map[int]*zmq4.Socket

	serverSockets map[int]*zmq4.Socket

	CoordChan = make(chan struct {
		Message *types.CoordinationMessage
		From    int
	}, 100)

	VcmChan = make(chan struct {
		Vcm  types.VCM
		From int
	}, 100)

	TokenChan = make(chan struct {
		Token types.Token
		From  int
	}, 100)

	RequestChan = make(chan *types.ClientMessage, 100)

	ReplicaChan = make(chan struct {
		Rep  *types.ReplicaStructure
		From int
	}, 100)
)

func Initialise() {
	Context, err := zmq4.NewContext()
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	serverSockets = make(map[int]*zmq4.Socket, variables.K)
	for i := 0; i < variables.K; i++ {
		serverSockets[i], err = Context.NewSocket(zmq4.REP)
		if err != nil {
			logger.ErrLogger.Fatal(err)
		}
		serverAddr := config.GetServerAddress(i)
		err = serverSockets[i].Bind(serverAddr)
		if err != nil {
			logger.ErrLogger.Fatal(err)
		}
	}
	RcvSockets = make(map[int]*zmq4.Socket)
	SndSockets = make(map[int]*zmq4.Socket)
	for i := 0; i < variables.N; i++ {
		if i == variables.Id {
			continue
		}
		RcvSockets[i], err = Context.NewSocket(zmq4.REP)
		if err != nil {
			logger.ErrLogger.Fatal(err)
		}
		rcvAddr := strings.Replace(
			config.GetRepAddress(i),
			"localhost",
			"*",
			1)
		err = RcvSockets[i].Bind(rcvAddr)
		if err != nil {
			logger.ErrLogger.Fatal(err, " "+rcvAddr)
		}
		logger.OutLogger.Println("Binded on ", rcvAddr)
		sndAddr := config.GetReqAddress(i)
		SndSockets[i], err = Context.NewSocket(zmq4.REQ)
		if err != nil {
			logger.ErrLogger.Fatal(err)
		}
		err = SndSockets[i].Connect(sndAddr)
		if err != nil {
			logger.ErrLogger.Fatal(err)
		}
		logger.OutLogger.Println("Connected to ", sndAddr)
	}
}

func Broadcast(message types.Message) {
	for i := 0; i < variables.N; i++ {
		if i == variables.Id {
			continue
		}
		SendMessage(message, i)
	}
}

func SendMessage(message types.Message, to int) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(message)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	_, err = SndSockets[to].SendBytes(w.Bytes(), 0)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	_, err = SndSockets[to].Recv(0)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
}

func SendReplica(replica *types.ReplicaStructure, to int) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(replica)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	message := types.NewMessage(w.Bytes(), "ReplicaStructure")
	SendMessage(message, to)
}

func Subscribe() {
	for i := 0; i < variables.K; i++ {
		go func(i int) {
			for {
				message, err := serverSockets[i].RecvBytes(0)
				if err != nil {
					logger.ErrLogger.Fatal(err)
				}
				go handleRequest(message)
			}
		}(i)
	}
	for i := 0; i < variables.N; i++ {
		if i == variables.Id {
			continue
		}
		go func(i int) {
			for {
				message, err := RcvSockets[i].RecvBytes(0)
				if err != nil {
					logger.ErrLogger.Fatal(err)
				}
				go handleMessage(message)
				_, err = RcvSockets[i].Send("OK.", 0)
				if err != nil {
					logger.ErrLogger.Fatal(err)
				}
			}
		}(i)
	}
}

func handleRequest(msg []byte) {
	cm := new(types.ClientMessage)
	buffer := bytes.NewBuffer(msg)
	decoder := gob.NewDecoder(buffer)
	err := decoder.Decode(&cm)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	RequestChan <- cm
}

func ReplyClient(reply *types.Reply) {
	to := reply.Client
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(reply)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	message := types.NewMessage(w.Bytes(), "Reply")
	w = new(bytes.Buffer)
	encoder = gob.NewEncoder(w)
	err = encoder.Encode(message)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	_, err = serverSockets[to].SendBytes(w.Bytes(), 0)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
}

func handleMessage(msg []byte) {
	message := new(types.Message)
	buffer := bytes.NewBuffer([]byte(msg))
	decoder := gob.NewDecoder(buffer)
	err := decoder.Decode(&message)
	if err != nil {
		logger.ErrLogger.Fatal(err)
	}
	switch message.Type {
	case "CoordinationMessage":
		coordination := new(types.CoordinationMessage)
		err = decoder.Decode(&coordination)
		if err != nil {
			logger.ErrLogger.Fatal(err)
		}
		CoordChan <- struct {
			Message *types.CoordinationMessage
			From    int
		}{Message: coordination, From: message.From}
		break
	case "VCM":
		vcm := new(types.VCM)
		err = decoder.Decode(&vcm)
		if err != nil {
			logger.ErrLogger.Fatal(err)
		}
		VcmChan <- struct {
			Vcm  types.VCM
			From int
		}{Vcm: *vcm, From: message.From}
		break
	case "Token":
		token := new(types.Token)
		err = decoder.Decode(&token)
		if err != nil {
			logger.ErrLogger.Fatal(err)
		}
		TokenChan <- struct {
			Token types.Token
			From  int
		}{Token: *token, From: message.From}
		break
	case "ReplicaStructure":
		replica := new(types.ReplicaStructure)
		err = decoder.Decode(&replica)
		if err != nil {
			logger.ErrLogger.Fatal(err)
		}
		ReplicaChan <- struct {
			Rep  *types.ReplicaStructure
			From int
		}{Rep: replica, From: message.From}
	}
}

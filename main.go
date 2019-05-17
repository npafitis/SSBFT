package main

import (
	"SSBFT/app"
	"SSBFT/app/messenger"
	"SSBFT/config"
	"SSBFT/logger"
	"SSBFT/variables"
	"log"
	"os"
	"os/signal"
	"strconv"
)

func Initialise(id int, n int, t int, k int) {
	variables.Initialise(id, n, t, k)
	logger.InitialiseLogger()
	config.InitialiseIP(n)
	messenger.InitialiseMessenger()
	app.InitializeAutomaton()
	app.InitializeViewChange()
	app.InitializeFailureDetector()
	app.InitializeEstablishment()
	app.InitializeReplication()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			for i := 0; i < n; i++ {
				if i == id {
					continue
				}
				messenger.RcvSockets[i].Close()
				messenger.SndSockets[i].Close()
			}
			os.Exit(0)
		}
	}()
}

func main() {
	done := make(chan interface{})
	args := os.Args[1:]
	if len(args) < 4 {
		log.Fatal("Arguments should be 'ssbft <id> <n> <f>...")
	}
	id, _ := strconv.Atoi(args[0])
	n, _ := strconv.Atoi(args[1])
	t, _ := strconv.Atoi(args[2])
	k, _ := strconv.Atoi(args[3])

	Initialise(id, n, t, k)
	messenger.Subscribe()

	go messenger.TransmitMessages()
	go app.ByzantineReplication()
	go app.ViewChangeMonitor()
	go app.CoordinatingAutomaton()

	_ = <-done
}

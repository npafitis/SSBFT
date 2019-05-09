package main

import (
	"SSBFT/app"
	"SSBFT/app/messenger"
	"SSBFT/config"
	"SSBFT/logger"
	"SSBFT/variables"
	"os"
	"os/signal"
	"strconv"
)

func Initialise(id int, n int) {
	variables.Initialise(id, n)
	logger.Initialise()
	config.Initialise(n)
	messenger.Initialise()
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
	if len(args) < 2 {
		logger.ErrLogger.Fatal("Arguments should be 'ssbft <id> <n> <f>...")
	}
	id, _ := strconv.Atoi(args[0])
	n, _ := strconv.Atoi(args[1])

	Initialise(id, n)
	messenger.Subscribe()

	go app.ByzantineReplication()
	go app.ViewChangeMonitor()
	go app.CoordinatingAutomaton()

	_ = <-done
}

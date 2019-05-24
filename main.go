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

func Initialise(id int, n int, t int, k int, scenario config.Scenario) {
	variables.Initialise(id, n, t, k)

	config.InitialiseIP(n)
	config.InitialiseScenario(scenario)

	logger.InitialiseLogger()
	logger.OutLogger.Println(
		"N", variables.N,
		"ID", variables.Id,
		"F", variables.F,
		"Threshold T", variables.T,
		"Client Size", variables.K,
	)

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
	if len(args) < 5 {
		log.Fatal("Arguments should be 'ssbft <id> <n> <f> <k> <scenario>")
	}
	id, _ := strconv.Atoi(args[0])
	n, _ := strconv.Atoi(args[1])
	t, _ := strconv.Atoi(args[2])
	k, _ := strconv.Atoi(args[3])
	tmp, _ := strconv.Atoi(args[4])
	scenario := config.Scenario(tmp)

	Initialise(id, n, t, k, scenario)
	messenger.Subscribe()

	go messenger.TransmitMessages()
	go app.FailDetector()
	go app.ByzantineReplication()
	go app.ViewChangeMonitor()
	go app.CoordinatingAutomaton()
	_ = <-done
}

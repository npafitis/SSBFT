package logger

import (
	"SSBFT/config"
	"SSBFT/variables"
	"log"
	"os"
	"strconv"
	"time"
)

var OutLogger *log.Logger

var ErrLogger *log.Logger

func InitialiseLogger() {
	outFolder := "./logs/"
	errFolder := "./logs/"
	switch config.TestCase {
	case config.NORMAL:
		outFolder += "normal/out/"
		errFolder += "normal/err/"
		break
	case config.STALE_VIEWS:
		outFolder += "stale_views/out/"
		errFolder += "stale_views/err/"
		break
	case config.STALE_STATES:
		outFolder += "stale_states/out/"
		errFolder += "stale_states/err/"
		break
	case config.STALE_REQUESTS:
		outFolder += "stale_requests/out/"
		errFolder += "stale_requests/err/"
		break
	case config.BYZANTINE_PRIM:
		outFolder += "byzantine_prim/out/"
		errFolder += "byzantine_prim/err/"
		break
	case config.NON_SS:
		outFolder += "non_ss/out/"
		errFolder += "non_ss/err/"
	}
	output := outFolder + "output_" + strconv.Itoa(variables.Id) + "_" + time.Now().UTC().String() + ".txt"
	errorf := errFolder + "err_" + strconv.Itoa(variables.Id) + "_" + time.Now().UTC().String() + ".txt"
	outFile, err := os.Create(output)
	if err != nil {
		log.Fatal(err)
	}
	errFile, err := os.Create(errorf)
	if err != nil {
		log.Fatal(err)
	}
	OutLogger = log.New(outFile, "", 0)
	ErrLogger = log.New(errFile, "", 0)
	OutLogger.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	ErrLogger.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

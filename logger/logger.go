package logger

import (
	"SSBFT/variables"
	"log"
	"os"
	"strconv"
	"time"
)

var OutLogger *log.Logger

var ErrLogger *log.Logger

func Initialise() {
	output := "./logs/output_" + strconv.Itoa(variables.Id) + "_" + time.Now().UTC().String() + ".txt"
	errorf := "./logs/err_" + strconv.Itoa(variables.Id) + "_" + time.Now().UTC().String() + ".txt"
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

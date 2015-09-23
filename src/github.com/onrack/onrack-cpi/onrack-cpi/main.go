package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"encoding/json"
	"io/ioutil"

	"github.com/onrack/onrack-cpi/bosh"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/cpi"
)

func exitWithError(err error) {
	fmt.Println(bosh.BuildErrorResponse(err, false, ""))
	os.Exit(1)
}

func exitWithResult(result interface{}) {
	fmt.Println(bosh.BuildResultResponse(result, ""))
	os.Exit(0)
}

func main() {
	log.SetOutput(ioutil.Discard)
	loggingEnabled := os.Getenv("ONRACK_CPI_ENABLE_LOGGING")

	if strings.ToLower(loggingEnabled) == "true" {
		log.SetOutput(os.Stderr)
	}

	configPath := flag.String("configPath", "", "Path to configuration file")
	flag.Parse()

	file, err := os.Open(*configPath)
	defer file.Close()

	if err != nil {
		exitWithError(err)
	}

	fileBody, err := ioutil.ReadAll(file)
	if err != nil {
		exitWithError(err)
	}

	var cpiConfig config.Cpi
	err = json.Unmarshal(fileBody, &cpiConfig)
	if err != nil {
		exitWithError(err)
	}

	reqBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		exitWithError(err)
	}

	req := bosh.CpiRequest{}
	err = json.Unmarshal(reqBytes, &req)
	if err != nil {
		exitWithError(err)
	}

	implemented, err := cpi.ImplementsMethod(req.Method)
	if err != nil {
		exitWithError(err)
	}

	if !implemented {
		exitWithError(fmt.Errorf("Method: %s is not implemented", req.Method))
	}

	switch req.Method {
	case cpi.CREATE_STEMCELL:
		cid, err := cpi.CreateStemcell(cpiConfig, req.Arugments)
		if err != nil {
			exitWithError(fmt.Errorf("Error running CreateStemcell: %s", err))
		}
		exitWithResult(cid)
	case cpi.DELETE_STEMCELL:
		err = cpi.DeleteStemcell(cpiConfig, req.Arugments)
		if err != nil {
			exitWithError(fmt.Errorf("Error running DeleteStemcell: %s", err))
		}
		exitWithResult("")
	default:
		exitWithError(fmt.Errorf("Unexpected command: %s dispatched...aborting", req.Method))
	}
}

package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"

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
	logLevel := os.Getenv("ONRACK_CPI_LOG_LEVEL")
	if logLevel != "" {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	switch logLevel {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "FATAL":
		log.SetLevel(log.FatalLevel)
	}

	configPath := flag.String("configPath", "", "Path to configuration file")
	flag.Parse()

	file, err := os.Open(*configPath)
	defer file.Close()

	if err != nil {
		log.Error(fmt.Sprintf("unable to open configuration file %s", err))
		exitWithError(err)
	}

	cpiConfig, err := config.New(file)
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
		cid, err := cpi.CreateStemcell(cpiConfig, req.Arguments)
		if err != nil {
			exitWithError(fmt.Errorf("Error running CreateStemcell: %s", err))
		}
		exitWithResult(cid)
	case cpi.CREATE_VM:
		vmcid, err := cpi.CreateVM(cpiConfig, req.Arguments)
		if err != nil {
			exitWithError(fmt.Errorf("Error running CreateVM: %s", err))
		}
		exitWithResult(vmcid)
	case cpi.DELETE_STEMCELL:
		err = cpi.DeleteStemcell(cpiConfig, req.Arguments)
		if err != nil {
			exitWithError(fmt.Errorf("Error running DeleteStemcell: %s", err))
		}
		exitWithResult("")
	case cpi.DELETE_VM:
		err = cpi.DeleteVM(cpiConfig, req.Arguments)
		if err != nil {
			exitWithError(fmt.Errorf("Error running DeleteVM: %s", err))
		}
		exitWithResult("")
	case cpi.SET_VM_METADATA:
		exitWithResult("")
	case cpi.HAS_VM:
		hasVM, err := cpi.HasVM(cpiConfig, req.Arguments)
		if err != nil {
			exitWithError(fmt.Errorf("Error running HasVM: %s", err))
		}
		exitWithResult(hasVM)
	case cpi.CONFIGURE_NETWORKS:
		err := cpi.ConfigureNetworks(cpiConfig, req.Arguments)
		if err != nil {
			exitWithError(fmt.Errorf("Error running ConfigureNetworks: %s", err))
		}
	default:
		exitWithError(fmt.Errorf("Unexpected command: %s dispatched...aborting", req.Method))
	}
}

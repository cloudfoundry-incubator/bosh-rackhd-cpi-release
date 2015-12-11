package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	log "github.com/Sirupsen/logrus"

	"encoding/json"
	"io/ioutil"

	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/cpi"
)

var responseLogBuffer *bytes.Buffer

func exitWithDefaultError(err error) {
	fmt.Println(bosh.BuildDefaultErrorResponse(err, false, responseLogBuffer.String()))
	responseLogBuffer.Reset()
	os.Exit(1)
}

func exitWithNotSupportedError(err error) {
	fmt.Println(bosh.BuildErrorResponse(err, bosh.NotSupportedErrorType, false, responseLogBuffer.String()))
	responseLogBuffer.Reset()
	os.Exit(1)
}

func exitWithResult(result interface{}) {
	fmt.Println(bosh.BuildResultResponse(result, responseLogBuffer.String()))
	responseLogBuffer.Reset()
	os.Exit(0)
}

func main() {
	responseLogBuffer = new(bytes.Buffer)
	multiWriter := io.MultiWriter(os.Stderr, responseLogBuffer)
	logLevel := os.Getenv("RACKHD_CPI_LOG_LEVEL")
	log.SetOutput(multiWriter)

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
		exitWithDefaultError(err)
	}

	cpiConfig, err := config.New(file)
	if err != nil {
		exitWithDefaultError(err)
	}

	reqBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		exitWithDefaultError(err)
	}

	req := bosh.CpiRequest{}
	err = json.Unmarshal(reqBytes, &req)
	if err != nil {
		exitWithDefaultError(err)
	}

	implemented, err := cpi.ImplementsMethod(req.Method)
	if err != nil {
		exitWithDefaultError(err)
	}

	if !implemented {
		exitWithNotSupportedError(fmt.Errorf("Method: %s is not implemented", req.Method))
	}

	switch req.Method {
	case cpi.CREATE_STEMCELL:
		cid, err := cpi.CreateStemcell(cpiConfig, req.Arguments)
		if err != nil {
			exitWithDefaultError(fmt.Errorf("Error running CreateStemcell: %s", err))
		}
		exitWithResult(cid)
	case cpi.CREATE_VM:
		vmcid, err := cpi.CreateVM(cpiConfig, req.Arguments)
		if err != nil {
			exitWithDefaultError(fmt.Errorf("Error running CreateVM: %s", err))
		}
		exitWithResult(vmcid)
	case cpi.DELETE_STEMCELL:
		err = cpi.DeleteStemcell(cpiConfig, req.Arguments)
		if err != nil {
			exitWithDefaultError(fmt.Errorf("Error running DeleteStemcell: %s", err))
		}
		exitWithResult("")
	case cpi.DELETE_VM:
		err = cpi.DeleteVM(cpiConfig, req.Arguments)
		if err != nil {
			exitWithDefaultError(fmt.Errorf("Error running DeleteVM: %s", err))
		}
		exitWithResult("")
	case cpi.SET_VM_METADATA:
		err := cpi.SetVMMetadata(cpiConfig, req.Arguments)
		if err != nil {
			exitWithDefaultError(fmt.Errorf("Error running SetVMMetadata: %s", err))
		}
		exitWithResult("")
	case cpi.HAS_VM:
		hasVM, err := cpi.HasVM(cpiConfig, req.Arguments)
		if err != nil {
			exitWithDefaultError(fmt.Errorf("Error running HasVM: %s", err))
		}
		exitWithResult(hasVM)
	default:
		exitWithDefaultError(fmt.Errorf("Unexpected command: %s dispatched...aborting", req.Method))
	}
}

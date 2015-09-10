package main

import (
	"os"
	"flag"

	"io/ioutil"
	"github.com/onrack/onrack-cpi/cli"
	"github.com/onrack/onrack-cpi/cpi"
	"github.com/onrack/onrack-cpi/config"
	"log"
	"encoding/json"
)


func main() {
	configPath := flag.String("configPath", "", "Path to configuration file")
	flag.Parse()

	file, err := os.Open(*configPath)
	if err != nil {
		log.Fatalf("Error opening file %s", err)
	}
	defer file.Close()

	fileBody, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("Error reading file %s", err)
	}

	var cpiConfig config.Cpi
	err = json.Unmarshal(fileBody, &cpiConfig)
	if err != nil {
		log.Fatalf("Error unmarshalling cpi config %s", err)
	}

	reqBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("Error reading stdin %s", err)
	}

	command, extInput, err := cli.ParseCommand(reqBytes)
	if err != nil {
		log.Fatalf("Error parsing command %s", err)
	}

	switch command {
	case cpi.CREATE_STEMCELL:
		cpi.CreateStemcell(cpiConfig, extInput)
	default:
		log.Fatal("Should not get here")
	}

}
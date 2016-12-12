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
  log.Error(err)
  fmt.Println(bosh.BuildDefaultErrorResponse(err, false, responseLogBuffer.String()))
  responseLogBuffer.Reset()
  os.Exit(1)
}

func exitWithNotImplementedError(err error) {
  fmt.Println(bosh.BuildErrorResponse(err, bosh.NotImplementedErrorType, false, responseLogBuffer.String()))
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
  default:
    log.SetLevel(log.DebugLevel)
  }

  configPath := flag.String("configPath", "", "Path to configuration file")
  flag.Parse()

  file, err := os.Open(*configPath)
  defer file.Close()

  if err != nil {
    log.Error(fmt.Sprintf("unable to open configuration file %s", err))
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

  cpiConfig, err := config.New(file, req)
  if err != nil {
    exitWithDefaultError(err)
  }

  implemented, err := cpi.ImplementsMethod(req.Method)
  if err != nil {
    exitWithDefaultError(err)
  }

  if !implemented {
    exitWithNotImplementedError(fmt.Errorf("Method: %s is not implemented", req.Method))
  }

  switch req.Method {
  case bosh.CREATE_STEMCELL:
    cid, err := cpi.CreateStemcell(cpiConfig, req.Arguments)
    if err != nil {
      exitWithDefaultError(fmt.Errorf("Error running CreateStemcell: %s", err))
    }
    exitWithResult(cid)
  case bosh.CREATE_VM:
    vmcid, err := cpi.CreateVM(cpiConfig, req.Arguments)
    if err != nil {
      exitWithDefaultError(fmt.Errorf("Error running CreateVM: %s", err))
    }
    exitWithResult(vmcid)
  case bosh.DELETE_STEMCELL:
    err = cpi.DeleteStemcell(cpiConfig, req.Arguments)
    if err != nil {
      exitWithDefaultError(fmt.Errorf("Error running DeleteStemcell: %s", err))
    }
    exitWithResult("")
  case bosh.DELETE_VM:
    err = cpi.DeleteVM(cpiConfig, req.Arguments)
    if err != nil {
      exitWithDefaultError(fmt.Errorf("Error running DeleteVM: %s", err))
    }
    exitWithResult("")
  case bosh.SET_VM_METADATA:
    err := cpi.SetVMMetadata(cpiConfig, req.Arguments)
    if err != nil {
      exitWithDefaultError(fmt.Errorf("Error running SetVMMetadata: %s", err))
    }
    exitWithResult("")
  case bosh.HAS_VM:
    hasVM, err := cpi.HasVM(cpiConfig, req.Arguments)
    if err != nil {
      exitWithDefaultError(fmt.Errorf("Error running HasVM: %s", err))
    }
    exitWithResult(hasVM)
  case bosh.CREATE_DISK:
    diskCID, err := cpi.CreateDisk(cpiConfig, req.Arguments)
    if err != nil {
      exitWithDefaultError(fmt.Errorf("Error running CreateDisk: %s", err))
    }
    exitWithResult(diskCID)
  case bosh.DELETE_DISK:
    err := cpi.DeleteDisk(cpiConfig, req.Arguments)
    if err != nil {
      exitWithDefaultError(fmt.Errorf("Error running DeleteDisk: %s", err))
    }
    exitWithResult("")
  case bosh.ATTACH_DISK:
    err := cpi.AttachDisk(cpiConfig, req.Arguments)
    if err != nil {
      exitWithDefaultError(fmt.Errorf("Error running AttachDisk: %s", err))
    }
    exitWithResult("")
  case bosh.DETACH_DISK:
    err := cpi.DetachDisk(cpiConfig, req.Arguments)
    if err != nil {
      exitWithDefaultError(fmt.Errorf("Error running DetachDisk: %s", err))
    }
    exitWithResult("")
  case bosh.HAS_DISK:
    diskExists, err := cpi.HasDisk(cpiConfig, req.Arguments)
    if err != nil {
      exitWithDefaultError(fmt.Errorf("Error running HasDisk: %s", err))
    }
    exitWithResult(diskExists)
  case bosh.GET_DISKS:
    diskCIDs, err := cpi.GetDisks(cpiConfig, req.Arguments)
    if err != nil {
      exitWithDefaultError(fmt.Errorf("Error running GetDisks: %s", err))
    }
    exitWithResult(diskCIDs)
  default:
    exitWithDefaultError(fmt.Errorf("Unexpected command: %s dispatched  .aborting", req.Method))
  }
}

package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"os/exec"
	"os"
	httpclient "github.com/cppforlife/baremetal_cpi/utils/httpclient"
	"bytes"
	"io/ioutil"
	"strings"
	"fmt"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type CreateStemcell struct {
	logger boshlog.Logger
}

type CreateStemcellCloudProps struct{}

func NewCreateStemcell(logger boshlog.Logger) CreateStemcell {
	return CreateStemcell{logger}
}

func (a CreateStemcell) Run(imagePath string, _ CreateStemcellCloudProps) (StemcellCID, error) {
	logTag := "create_stemcell"
	a.logger.Info(logTag, "Extracting stemcell from '%s'", imagePath)

   	os.Mkdir("tmp", os.FileMode(0755))
	command := fmt.Sprintf("tar -C %s -xzf %s 2>&1", "tmp/", imagePath)
	_, err:= runCommand(command)
	if err != nil {
		bosherr.WrapErrorf(err, "Error extracting image '%s'", imagePath)
	}

	a.logger.Info(logTag, "Extracted stemcell")
	file, err := os.Open("tmp/image-disk1.vmdk")
	if (err != nil) {
		bosherr.WrapErrorf(err, "Error opening file")
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if (err != nil) {
		bosherr.WrapErrorf(err, "Error getting file info")
	}
	fileSize := fileStat.Size()
	a.logger.Info(logTag, "File Size is '%d'", fileSize)

	body := ioutil.NopCloser(file)
	client := httpclient.NewHTTPClient(httpclient.DefaultClient, a.logger)
	resp, err := client.Put("endpoint", body, fileSize)

	if err != nil {
		bosherr.WrapErrorf(err, "Error uploading stemcell")
	}

	a.logger.Info(logTag, "Succeeded uploading stemcell '%s'", resp.Status)
	responseBody, _ := ioutil.ReadAll(resp.Body)
	uuid := string(responseBody)
	a.logger.Info(logTag, "UUID '%s'  \n", uuid)

    return StemcellCID(uuid), nil
}

func runCommand(cmd string) (string, error) {
	var stdout bytes.Buffer
	shCmd := exec.Command("sh", "-c", cmd)
	shCmd.Stdout = &stdout
	if err := shCmd.Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
}

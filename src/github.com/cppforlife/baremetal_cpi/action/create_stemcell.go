package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"fmt"
	"os/exec"
	"os"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	httpclient "github.com/cloudfoundry/bosh-utils/httpclient"
	"bytes"
	"io/ioutil"
	"strings"
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

	client := httpclient.NewHTTPClient(httpclient.DefaultClient, a.logger)
	resp, err := client.Put("upload_path", nil)

	if err != nil {
		bosherr.WrapErrorf(err, "Error uploading stemcell")
	}
	defer resp.Body.Close()

	a.logger.Info(logTag, "Succeeded uploading stemcell '%s'", resp.Status)
	responseBody, _ := ioutil.ReadAll(resp.Body)
	uuid := string(responseBody)
	fmt.Printf("UUID %s  \n", uuid)

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
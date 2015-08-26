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
	uuidGen "github.com/cloudfoundry/bosh-utils/uuid"
)

type CreateStemcell struct {
	APIServer string
	logger boshlog.Logger
}

type CreateStemcellCloudProps struct{}

func NewCreateStemcell(APIServer string, logger boshlog.Logger) CreateStemcell {
	return CreateStemcell{APIServer, logger}
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

	uuid, err := uuidGen.NewGenerator().Generate()
	if (err != nil) {
		bosherr.WrapErrorf(err, "Error generating UUID")
	}

	url := fmt.Sprintf("http://%s:8080/api/common/files/%s", a.APIServer, uuid)
	body := ioutil.NopCloser(file)
	client := httpclient.NewHTTPClient(httpclient.DefaultClient, a.logger)

	resp, err := client.Put(url, body, fileSize)
	if err != nil {
		bosherr.WrapErrorf(err, "Error uploading stemcell")
	}

	a.logger.Info(logTag, "Succeeded uploading stemcell '%s'", resp.Status)
	responseBody, _ := ioutil.ReadAll(resp.Body)
	stemcell_uuid := string(responseBody)
	a.logger.Info(logTag, "UUID '%s'  \n", stemcell_uuid)

	//TODO verify what uuid is needed later in the api. This is returned from the server
	// and appended to the uuid we've generated above. Format: localuuid_remoteuuid
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

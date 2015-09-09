package cpi
import (
	"errors"
	"os"
	"os/exec"
	"io/ioutil"
	"fmt"
	"net/http"
	"github.com/nu7hatch/gouuid"
	"log"
)

func CreateStemcell(config Config, args ExternalInput) (string, error) {
	imagePath := ""

	// extract into function here
	log.Printf("Extracting stemcell from '%s'", imagePath)

	os.Mkdir("tmp", os.FileMode(0755))
	tarCmd := exec.Command("tar", "-C tmp/", "-xzf", fmt.Sprintf("%s", imagePath))
	tarCmdOutput, err := tarCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Error extracting image '%s': %s, commnad output was %s", imagePath, err, tarCmdOutput)
	}

	//a.logger.Info(logTag, "Extracted stemcell")
	file, err := os.Open("tmp/image-disk1.vmdk") // coupling between the vSphere stemcell format
	if (err != nil) {
		return "", errors.New("Error opening file")
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if (err != nil) {
		return "", errors.New("Error getting file info")
	}
	fileSize := fileStat.Size()
	log.Printf("File Size is '%d'", fileSize)

	// end extract into function

	uuid, err := uuid.NewV4()
	if (err != nil) {
		return "", errors.New("Error generating UUID")
	}

	url := fmt.Sprintf("http://%s:8080/api/common/files/%s", config.ApiServer, uuid.String())
	body := ioutil.NopCloser(file)

	request, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return "", err
	}

	request.ContentLength = fileSize
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	log.Printf("Succeeded uploading stemcell '%s'", resp.Status)
	responseBody, _ := ioutil.ReadAll(resp.Body)
	stemcell_uuid := string(responseBody)
	log.Printf("UUID '%s'  \n", stemcell_uuid)

	//TODO returning locally generated uuid because create_vm needs filename. delete_stemcellneeds the uuid.


	if imagePath == "" {
		panic("you're not done!")
	}
	return uuid.String(), nil

}

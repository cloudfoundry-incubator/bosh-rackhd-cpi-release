package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	httpclient "github.com/cppforlife/baremetal_cpi/utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

type DeleteStemcell struct {
	APIServer string
	logger boshlog.Logger
	logTag string
}

func NewDeleteStemcell(APIServer string,logger boshlog.Logger) DeleteStemcell {
	return DeleteStemcell{
		APIServer: APIServer,
		logger: logger,
		logTag: "delete_stemcell",
	}
}

func (a DeleteStemcell) Run(stemcellCID StemcellCID) (interface{}, error) {
	url := fmt.Sprintf("http://%s:8080/api/common/files/metadata/%s", a.APIServer, stemcellCID)
	client := httpclient.NewHTTPClient(httpclient.DefaultClient, a.logger)

	resp, err := client.Get(url)
	if err != nil {
		bosherr.WrapErrorf(err, "Error uploading stemcell")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		bosherr.WrapErrorf(err, "Error getting response body")
	}
	a.logger.Info(a.logTag, "Status is  '%s'", resp.Status)
	var metadata FileMetadata
	err = json.Unmarshal(body, &metadata)
	if err != nil {
		return nil, bosherr.WrapError(err, "Unmarshalling File Metadata")
	}

	for _, each := range metadata {
		a.logger.Info(a.logTag, "File uuid is '%d'", each.UUID)
		deleteUrl := fmt.Sprintf("http://%s:8080/api/common/files/%s", a.APIServer, each.UUID)
		resp, err = client.Delete(deleteUrl)
		if err != nil {
			bosherr.WrapErrorf(err, "Error deleting stemcell")
		}
	}
	return nil, nil
}

type FileMetadata []struct {
	UUID string `json:"uuid"`
}
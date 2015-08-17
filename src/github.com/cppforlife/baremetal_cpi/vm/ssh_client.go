package vm

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	gossh "golang.org/x/crypto/ssh"
)

type SSHClient struct {
	client  *gossh.Client
	uuidGen boshuuid.Generator
	logger  boshlog.Logger
}

func NewSSHClient(
	client *gossh.Client,
	uuidGen boshuuid.Generator,
	logger boshlog.Logger,
) SSHClient {
	return SSHClient{client: client, uuidGen: uuidGen, logger: logger}
}

func (c SSHClient) ExecuteScript(name string, contents string, runInBg ...bool) error {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return bosherr.WrapError(err, "Generating script file id")
	}

	// todo should not be in /tmp
	scriptPath := fmt.Sprintf("/tmp/execute-script-%s-%s.sh", name, string(id))

	err = c.uploadAscii(bytes.NewBufferString(contents), scriptPath)
	if err != nil {
		return bosherr.WrapError(err, "Uploading script")
	}

	_, err = c.run("/bin/chmod", "+x", scriptPath)
	if err != nil {
		return bosherr.WrapError(err, "Making script executable")
	}

	if len(runInBg) == 0 {
		_, err = c.run(scriptPath)
	} else if runInBg[0] == true {
		err = c.runInBg(scriptPath)
	}

	if err != nil {
		return bosherr.WrapErrorf(err,
			"Executing script name '%s' at path '%s'", name, scriptPath)
	}

	return nil
}

func (c SSHClient) UploadConfig(path string, contents []byte) error {
	return c.uploadAscii(bytes.NewBuffer(contents), filepath.Join("/var/stemcell", path))
}

func (c SSHClient) DownloadConfig(path string) ([]byte, error) {
	return c.run("/bin/cat", filepath.Join("/var/stemcell", path))
}

func (c SSHClient) UploadStemcell(path, dstPath string) error {
	// todo use fs?
	fileReader, err := os.Open(path)
	if err != nil {
		return bosherr.WrapError(err, "Opening file for uploading")
	}

	defer fileReader.Close()

	return c.uploadTar(fileReader, dstPath)
}

func (c SSHClient) uploadAscii(contents io.Reader, dstPath string) error {
	return c.upload(contents, fmt.Sprintf("/bin/cat > '%s'", dstPath))
}

// uploadTar assumes that dstPath directory must *not* be present
func (c SSHClient) uploadTar(contents io.Reader, dstPath string) error {
	_, err := c.run("/bin/mkdir", dstPath)
	if err != nil {
		return bosherr.WrapError(err, "Making script executable")
	}

	return c.upload(contents, fmt.Sprintf("/bin/tar xzf - -C '%s'", dstPath))
}

func (c SSHClient) upload(contents io.Reader, cmd string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return bosherr.WrapError(err, "Creating new SSH session")
	}

	defer session.Close()

	uploadCh := make(chan error)

	uploadWriter, err := session.StdinPipe()
	if err != nil {
		return bosherr.WrapError(err, "Creating new SSH stdin pipe")
	}

	go func() {
		_, err = io.Copy(uploadWriter, contents) // todo n?
		if err != nil {
			uploadCh <- err
		}

		uploadCh <- uploadWriter.Close()
	}()

	var stdout, stderr bytes.Buffer

	session.Stdout = &stdout
	session.Stderr = &stderr

	// todo escape path
	scpErr := session.Run(cmd)

	// Wait for upload
	uploadErr := <-uploadCh

	if scpErr != nil {
		return bosherr.WrapErrorf(scpErr,
			"Running scp for uploading stdout: '%s' stderr: '%s'", stdout.String(), stderr.String())
	}

	if uploadErr != nil {
		return bosherr.WrapError(uploadErr, "Writing to scp for uploading")
	}

	return nil
}

func (c SSHClient) run(path string, args ...string) ([]byte, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating new SSH session")
	}

	defer session.Close()

	var stdout, stderr bytes.Buffer

	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(fmt.Sprintf("%s %s", path, strings.Join(args, " "))) // todo wtf escape
	if err != nil {
		return nil, bosherr.WrapErrorf(err,
			"Running command stdout: '%s' stderr: '%s'", stdout.String(), stderr.String())
	}

	return stdout.Bytes(), nil
}

func (c SSHClient) runInBg(path string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return bosherr.WrapError(err, "Creating new SSH session")
	}

	defer session.Close()

	err = session.Start(path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Running background command")
	}

	return nil
}

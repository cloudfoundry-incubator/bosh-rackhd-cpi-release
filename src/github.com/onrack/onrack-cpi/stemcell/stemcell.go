package stemcell

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"sync"
)

type Stemcell struct {
	stemcellPath string
	workDir      string
	fileHandle   *os.File
	extractOnce  sync.Once
	cleanOnce    sync.Once
}

func New(path string) *Stemcell {
	return &Stemcell{
		stemcellPath: path,
	}
}

func (s *Stemcell) Extract() (*os.File, error) {
	var err error

	s.extractOnce.Do(func() {
		var tarCmdOutput []byte
		var tarCmd *exec.Cmd

		s.workDir, err = ioutil.TempDir("", "stemcell")
		if err != nil {
			log.Printf("")
			return
		}

		log.Printf("Extracting stemcell from '%s'", s.stemcellPath)
		os.Mkdir(s.workDir, os.FileMode(0755))
		tarCmd = exec.Command("tar", "-C", s.workDir, "-xzf", fmt.Sprintf("%s", s.stemcellPath))
		tarCmdOutput, err = tarCmd.CombinedOutput()
		if err != nil {
			log.Printf("Error extracting image '%s': %s, command output was %s", s.stemcellPath, err, tarCmdOutput)
			return
		}

		s.fileHandle, err = os.Open(path.Join(s.workDir, "image-disk1.vmdk"))
		if err != nil {
			log.Printf("Error opening file")
			return
		}
	})

	return s.fileHandle, err
}

func (s *Stemcell) GetWorkDir() string {
	return s.workDir
}

func (s *Stemcell) CleanUp() error {
	var err error

	s.cleanOnce.Do(func() {
		err = os.RemoveAll(s.workDir)
		if err != nil {
			log.Printf("Error removing stemcell work directory")
			return
		}
		s.workDir = ""
	})

	return err
}

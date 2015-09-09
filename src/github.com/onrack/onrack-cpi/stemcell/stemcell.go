package stemcell

type Stemcell struct {
	stemcellPath string
	workDir      string
}

func New(path string) *Stemcell {
	return &Stemcell{
		stemcellPath: path,
	}
}
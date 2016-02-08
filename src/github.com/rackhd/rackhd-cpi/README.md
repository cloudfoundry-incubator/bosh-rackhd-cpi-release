For running the local tests:

Allow direnv to load this .envrc
Install the ginkgo CLI `go get github.com/onsi/ginkgo/ginkgo`
Install the matcher library `go get github.com/onsi/gomega`
`cd src/github.com/rackhd/rackhd-cpi/`
Set variable `export RACKHD_API_HOST=rackhd.server.ip`
Set variable `export RACKHD_API_PORT=8080`
Run tests with `ginkgo -r`

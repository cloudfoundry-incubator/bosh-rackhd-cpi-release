export ON_RACK_API_URI="192.168.252.131"

apt-get install -y direnv

cd bosh-external-cpi/
direnv allow .

cd src/github.com/onrack/onrack-cpi
ginkgo -r

echo "Test suite complete."

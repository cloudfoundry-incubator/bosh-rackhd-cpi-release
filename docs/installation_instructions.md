# Bosh Cloud Provider Interface for RackHD

## Prerequisites

You must have [installed RackHD](http://rackhd.readthedocs.org/en/latest/).

## Installation steps

### Enable MonorailAccess

You must first enable MonorailAccess. Browse to `https://yourrackhd.ip/rest/v1/api.html` and find the `POST /__Internal__/MonorailAccess/{value}` and input Enable into the text box and click "Try it out!"

If you then `GET /__Internal__/MonorailAccess/` it should return "enable."

You should then be able to query the server for its catalog of nodes with `curl "RACKHD_API_URL/api/common/nodes" | jq .` (or whatever browser/JSON viewer you choose to use.)

### Download Packages

Install BOSH init: <https://bosh.io/docs/install-bosh-init.html>

Download a BOSH OpenStack stemcell for Ubuntu Trusty: <https://bosh.io/stemcells/bosh-openstack-kvm-ubuntu-trusty-go_agent-raw>

Download the RackHD CPI release: <https://github.com/EMC-CMD/bosh-rackhd-cpi-release>

Then, build a release:

```
$ cd bosh-rackhd-cpi-release
$ bosh create release
```

### Create Public Key

`$ ssh-keygen -f ssh-key`

After prompting you for a passphrase, the command creates a public key, `ssh-key.pub`, and a matching private key, `ssh-key`, in your current directory. Later, if you want to log into provisioned machines on your network, you can specify the key explicitly with `ssh -i /path/to/director-ssh-key ...`, or you can configure a tool like `ssh-agent` to manage your keys for you.

### Prepare Manifest for Deployment

Create a file like `redis.yml` below. Update the path to the CPI release file, IP addresses, and public key top match your configuration.

```
name: redis

releases:
- name: rackhd-cpi
  url: file:///path/to/release/rackhd-cpi.tgz
- name: redis
  url: https://bosh.io/d/github.com/cloudfoundry-community/redis-boshrelease

networks:
- name: default
  type: manual

  subnets:
  - range: 192.168.1.0/24
    gateway: 192.168.1.1
    static: [192.168.1.2]

resource_pools:
- name: default
  network: default
  cloud_properties:
    public_key: "YOUR PUBLIC KEY"
  stemcell:
    name: bosh-openstack-kvm-ubuntu-trusty-go_agent-raw
    version: latest
    url: https://bosh.io/d/stemcells/bosh-openstack-kvm-ubuntu-trusty-go_agent-raw

compilation:
  workers: 1
  network: default
  reuse_compilation_vms: true
  cloud_properties:
    public_key: "YOUR PUBLIC KEY"

update:
  canaries: 1
  max_in_flight: 1
  canary_watch_time: 1000 - 30000
  update_watch_time: 1000 - 30000

jobs:
- name: redis
  instances: 1
  resource_pool: default
  networks:
  - name: default
    static_ips: [192.168.1.2]
  templates:
  - {name: redis, release: redis}
  properties:
    redis:
      password: r3d!s
      port: 6379

cloud_provider:
  template: {name: rackhd-cpi, release: rackhd-cpi}

  mbus: https://mbus:mbus-password@192.168.1.2:6868

  properties:
    rackhd-cpi:
      api_url: "RACKHD_ENDPOINT_URL"
      agent:
        mbus: "https://mbus:mbus-password@0.0.0.0:6868"
        blobstore:
          provider: local
          options: {blobstore_path: /var/vcap/micro_bosh/data/cache}

```

### Deploy Redis

`$ bosh-init deploy redis.yml`

Redis should be running on one of the nodes now. You can verify this by installing `redis-cli` and connecting to the deployment

```
$ redis-cli -h 192.168.1.2
redis> set FOO BAR
OK
redis> get FOO
BAR
```  

##How to deploy Cloud Foundry on Bare-Metal Machine(s)
Deploying Cloud Foundry on Bare-Metal Machine(s) require 3 steps:

1. Set up RackHD
2. Set up Bosh Director with RackHD CPI
3. Deploying Cloud Foundry

In this tutorial, we will do the example with vSphere, but the same steps can be applied with other IaaS or Bare-Metal environment. Let's begin!

##Setting up RackHD
The first step is to set up [RackHD](https://github.com/rackhd), an open source solution that manage your Bare-Metal environment. Please reference the [RackHD documentation](http://rackhd.readthedocs.org/) if you would like to learn more about RackHD. 

In a nutshell, RackHD provides DHCP service and picks up iPXE boot signal from Bare-Metal machines. Once RackHD receives signal from machines, a new entry will be added to the RackHD database. RackHD will then be able to fully control the machines through IPMI and AMT. (See diagram below.)

![Diagram 1](./img/cf-installation-1.jpg)

Please make sure that there is no other DHCP service connected to the nodes. The machines should be getting IP addresses from the DHCP service from RackHD server. In our environment, the nodes and the RackHD server are connected by a simple network switch. 

Now that you have a high level understanding of how RackHD works, let's install RackHD! You would need a empty machine (virtual or bare-metal) with Ubuntu installed that's connected to the network switch. In my case, I have just used vagrant to bring up a new Virtual Machine:

```
vagrant init ubuntu/wily64
vagrant up
```

Great! Now you should have an empty box with Ubuntu installed. Let's install the prerequisites of RackHD:

```
#!/bin/bash 
set -­e

sudo apt­-get ­-y update
sudo apt­-get ­-y dist­upgrade 
sudo apt-­get -­y autoremove
sudo apt­-get -­y install nodejs nodejs­legacy npm

#runtime dependencies
sudo apt-­get -­y install rabbitmq­server mongodb isc­dhcp­server
sudo apt-­get ­-y install snmp ipmitool ansible amtterm apt­mirror libkrb5­dev unzip

#Ubuntu 15.04 or later: use upstart instead of systemd 
sudo apt­-get -­y install upstart­sysv
sudo update-­initramfs ­-u

# AMTTool TNG
wget http://downloads.sourceforge.net/project/amttool­tng/1.7/amttool 
sudo chmod 755 ./amttool
sudo mv amttool /usr/bin/amttool­tng

#compile dependencies
sudo apt­-get -­y install git openssh­server pbuilder dh­make ubuntu­dev­tools devscripts
```

Now that the prerequisites are installed, we are ready to build RackHD from the latest source and use the `HWIMOBUILD` script to make debian packages for the components of RackHD. Then we install each of the debian packages by `dpkg`. 

```
RACKHD_INSTALL_DIR=~; cd $RACKHD_INSTALL_DIR 
git clone https://github.com/RackHD/RackHD 
RACKHD_PROJECT_DIR=${RACKHD_INSTALL_DIR}/RackHD
cd $RACKHD_PROJECT_DIR
git submodule update ­­init ­­recursive
git submodule foreach git pull origin master

sudo touch /etc/default/on­http
sudo touch /etc/default/on­dhcp­proxy 
sudo touch /etc/default/on­taskgraph 
sudo touch /etc/default/on­syslog 
sudo touch /etc/default/on­tftp

cd ${RACKHD_PROJECT_DIR}/on­http 
./HWIMO­BUILD
sudo dpkg -­i ./on­http_*.deb

cd ${RACKHD_PROJECT_DIR}/on­dhcp­proxy 
./HWIMO­BUILD
sudo dpkg ­i ./on­dhcp­proxy_*.deb

cd ${RACKHD_PROJECT_DIR}/on­taskgraph 
./HWIMO­BUILD
sudo dpkg ­i ./on­taskgraph_*.deb

cd ${RACKHD_PROJECT_DIR}/on­syslog 
./HWIMO­BUILD
sudo dpkg ­i ./on­syslog_*.deb


cd ${RACKHD_PROJECT_DIR}/on­tftp 
./HWIMO­BUILD
sudo dpkg ­i ./on­tftp_*.deb

sudo reboot
```

After reboot, ssh into the machine. Modify /etc/default/isc-dhcp-server.conf to look like
`https://github.com/RackHD/RackHD/blob/master/packer/ansible/roles/isc­dhcp­server/files/isc-dhcp-server` by adding
```
INTERFACES="eth1"
```
* `eth1` is the network interface that you use to talk to your node.

Then, modify /etc/dhcp/dhcp.conf to look like
`https://github.com/RackHD/RackHD/blob/master/packer/ansible/roles/isc-dhcp-server/files/dhcpd.conf` by adding
```
# RackHD added lines
deny duplicates;

ignore-client-uids true;

subnet 172.31.128.0 netmask 255.255.252.0 {
 range 172.31.128.2 172.31.131.254;
 # Use this option to signal to the PXE client that we are doing proxy DHCP
 option vendor-class-identifier "PXEClient";
}
```
Make sure to edit subnet accordingly. This is the subnet of rackhd server and its node.

Now, download `https://github.com/RackHD/RackHD/blob/master/packer/ansible/roles/monorail/files/config.json` and put it under `/opt/monorail/`. This is your rackhd-server config file, please edit its network accordingly.

Last but not least, follow instructions under `https://github.com/RackHD/RackHD/blob/master/packer/ansible/roles/images/tasks/main.yml` to download files from `https://bintray.com/artifact/download/rackhd/binary/builds/` and place them under `/home/vagrant/src/on-http/static/http/common/` and `/home/vagrant/src/on-tftp/static/tftp/` folders accordingly.

Finally, restart RackHD server components using
```
sudo service on-http restart
sudo service on-tftp restart
sudo service on-taskgraph restart
sudo service on-dhcp-proxy restart
sudo service on-syslog restart
```

To see rackhd's logs, you can go under `/var/log/upstart/`. They are `on-http.log`, `on-taskgraph.log`, `on-http.log`, `on-tftp.log`, and `on-syslog.log`.

If you run into problems during the RackHD installation process, don't panic! There is a slack channel that can help you out. The RackHD community is very responsive and help! You can get invitation to the Slack channel by coming [here](http://community.emccode.com).

Now that we have the RackHD server ready. We are ready to discover our nodes. At the time of writing, RackHD can only be accessed through RESTful web services and we have developed a command line for our day-to-day activities. 

## Setting up RackHD CLI
The RackHD CLI is open source and hosted in [https://github.com/EMC-CMD/rackhd-cli](https://github.com/EMC-CMD/rackhd-cli). 

In your workspace, you can clone the source for the RackHD CLI:

```
git clone https://github.com/EMC-CMD/rackhd-cli.git
```

Once you have the source code, you can then go into the directory and build the gem:

```
gem build rackhd-cli.gemspec
gem install rackhd-cli-x.y.z.gem
```

After that, the `rack` command should be available for you to use. For example, running `rack nodes` should tell you that there is no node registered to the RackHD server yet. 

```
ID                       | NAME                      | CID | STATUS | DISK CID | ACTIVE WORKFLOW
-------------------------|---------------------------|-----|--------|----------|----------------
```

## Setting up Nodes on RackHD server
Now that we have the RackHD server and the command line tool installed, we can start to TURN ON your machines. Make sure your machines has IPMI or other similar protocols (ie. AMT) that RackHD supports and have `Network Boot` turned on. 

You can run `watch rack nodes` and then turn on your machines. As the machines register themselves to RackHD, you should be able to set nodes being added, like this:

```
ID                       | NAME                      | CID | STATUS | DISK CID | ACTIVE WORKFLOW
-------------------------|---------------------------|-----|--------|----------|----------------
56c4a84de4dc603837faa636 | c0:3f:d5:60:51:b9 (node1) | n/a | n/a    | n/a      | n/a
56c4a84de4dc603837faa637 | c0:3f:d5:63:fd:b0 (node0) | n/a | n/a    | n/a      | n/a
56c4a861e4dc603837faa638 | c0:3f:d5:63:fe:13 (node2) | n/a | n/a    | n/a      | n/a
```

## Setting up Bosh Director with RackHD CPI
Please refer to this [manual](https://github.com/cloudfoundry-incubator/bosh-rackhd-cpi-release/wiki/Deploy%20a%20BOSH%20Director%20with%20the%20RackHD%20CPI) for creating a Bosh Director with RackHD CPI.

After deploying the Bosh Director, your environment should look similar to this:
![Diagram 2](./img/cf-installation-2.jpg)

## Setting up Cloud Foundry on Bare-Metal environment
Download source code for CF Release: 

```
git clone https://github.com/cloudfoundry/cf-release.git
```

Then create the CF release:

```
bosh create release
```

Upload the CF release to the director

```
bosh upload release <path to cf-release file>
```

Download the OpenStack stemcell from [bosh.io](http://bosh.io) and then upload it to the director

```
bosh upload stemcell <path to stemcell>
```

Modify the example manifest. In this example, I have combined all of CF components into just 3 machines. The Router is on the CF2 machine. There could be other deployment set up as well, such as running on the all components in one single box. 

On a side note, the RackHD server and Bosh Director can also run in a virtual environment. 

```
---
name: cf
director_uuid: <DIRECTOR UUID>

releases:
- {name: cf, version: latest}

networks:
- name: cf_private
  type: manual
  subnets:
  - range: 10.0.4.0/22
    gateway: 10.0.4.1
    dns: [10.0.4.1]
    static: [10.0.4.2 - 10.0.5.254]
    cloud_properties:
      subnet: subnet-a5138898

- name: cf_public
  type: manual
  subnets:
  - range: 10.0.0.0/22
    gateway: 10.0.0.1
    dns: [10.0.0.1]
    reserved: [10.0.0.1 - 10.0.0.255]
    static: [10.0.1.2 - 10.0.1.254]
    cloud_properties:
      subnet: subnet-126ef52f
      security_groups:
        - victor-sg

- name: elastic
  type: vip
  cloud_properties: {}

resource_pools:
- name: small_z1
  network: cf_private
  stemcell:
    name: bosh-aws-xen-hvm-ubuntu-trusty-go_agent
    version: latest
  cloud_properties:
    availability_zone: us-east-1e
    instance_type: c3.large

compilation:
  workers: 3
  network: cf_private
  reuse_compilation_vms: true
  cloud_properties:
    availability_zone: us-east-1e
    instance_type: c3.large

update:
  canaries: 1
  max_in_flight: 1
  serial: false
  canary_watch_time: 30000-600000
  update_watch_time: 5000-600000

jobs:
- name: giant1_z1
  instances: 1
  resource_pool: small_z1
  persistent_disk: 102400
  templates:
  - {name: nats, release: cf}
  - {name: nats_stream_forwarder, release: cf}
  - {name: etcd, release: cf}
  - {name: etcd_metrics_server, release: cf}
  - {name: hm9000, release: cf}
  - {name: postgres, release: cf}

  - {name: metron_agent, release: cf}
  - {name: route_registrar, release: cf}
  networks:
  - name: cf_private
    static_ips: [10.0.4.103]
  properties:
    etcd_metrics_server:
      nats:
        machines: [10.0.4.103]
        password: password
        username: nats
    route_registrar:
      routes:
      - name: hm9000
        port: 5155
        uris:
        - "hm9000.system.cf.local"

- name: giant2_z1
  instances: 1
  resource_pool: small_z1
  persistent_disk: 102400
  templates:
  - {name: doppler, release: cf}
  - {name: loggregator_trafficcontroller, release: cf}
  - {name: gorouter, release: cf}
  - {name: cloud_controller_ng, release: cf}
  - {name: cloud_controller_worker, release: cf}
  - {name: cloud_controller_clock, release: cf}
  - {name: uaa, release: cf}
  - {name: metron_agent, release: cf}
  - {name: route_registrar, release: cf}
  networks:
  - name: cf_private
    static_ips: [10.0.4.102]
  properties:
    traffic_controller: {zone: z1}
    dropsonde: {enabled: true}
    doppler: {zone: z1}
    doppler_endpoint:
      shared_secret: password
    route_registrar:
      routes:
      - name: api
        port: 9022
        uris:
        - "api.system.cf.local"
      - name: doppler
        port: 8081
        uris:
        - "doppler.system.cf.local"
      - name: uaa
        port: 8080
        uris:
        - "uaa.system.cf.local"
        - "*.uaa.system.cf.local"
        - "login.system.cf.local"
        - "*.login.system.cf.local"
      - name: loggregator
        port: 8080
        uris:
          - "loggregator.system.cf.local"
    login:
      catalina_opts: -Xmx768m -XX:MaxPermSize=256m
    uaa:
      admin:
        client_secret: password
      batch:
        password: password
        username: batch_user
      cc:
        client_secret: password
      scim:
        userids_enabled: true
        users:
        - admin|password|scim.write,scim.read,openid,cloud_controller.admin,doppler.firehose,routing.router_groups.read
    uaadb:
      address: 10.0.4.103
      databases:
      - {name: uaadb, tag: uaa}
      db_scheme: postgresql
      port: 5524
      roles:
      - {name: uaaadmin, password: password, tag: admin}

- name: runner_z1
  instances: 1
  resource_pool: small_z1
  templates:
  - {name: dea_next, release: cf}
  - {name: dea_logging_agent, release: cf}
  - {name: metron_agent, release: cf}
  networks:
  - name: cf_private
  properties:
    dea_next: {zone: z1}

properties:
  networks: {apps: cf_private}
  app_domains: [system.cf.local]
  cc:
    allow_app_ssh_access: false
    bulk_api_password: password
    db_encryption_key: password
    default_running_security_groups: [public_networks, dns]
    default_staging_security_groups: [public_networks, dns]
    install_buildpacks:
    - {name: java_buildpack, package: buildpack_java}
    - {name: ruby_buildpack, package: buildpack_ruby}
    - {name: nodejs_buildpack, package: buildpack_nodejs}
    - {name: go_buildpack, package: buildpack_go}
    - {name: python_buildpack, package: buildpack_python}
    - {name: php_buildpack, package: buildpack_php}
    - {name: staticfile_buildpack, package: buildpack_staticfile}
    - {name: binary_buildpack, package: buildpack_binary}
    internal_api_password: password
    quota_definitions:
      default:
        memory_limit: 102400
        non_basic_services_allowed: true
        total_routes: 1000
        total_services: -1
    security_group_definitions:
    - name: public_networks
      rules:
      - {destination: 0.0.0.0-9.255.255.255, protocol: all}
      - {destination: 11.0.0.0-169.253.255.255, protocol: all}
      - {destination: 169.255.0.0-172.15.255.255, protocol: all}
      - {destination: 172.32.0.0-192.167.255.255, protocol: all}
      - {destination: 192.169.0.0-255.255.255.255, protocol: all}
    - name: dns
      rules:
      - {destination: 0.0.0.0/0, ports: '53', protocol: tcp}
      - {destination: 0.0.0.0/0, ports: '53', protocol: udp}
    srv_api_uri: https://api.system.cf.local
    staging_upload_password: password
    staging_upload_user: staging_upload_user
  ccdb:
    address: 10.0.4.103
    databases:
    - {name: ccdb, tag: cc}
    db_scheme: postgres
    port: 5524
    roles:
    - {name: ccadmin, password: password, tag: admin}
  databases:
    databases:
    - {name: ccdb, tag: cc, citext: true}
    - {name: uaadb, tag: uaa, citext: true}
    port: 5524
    roles:
    - {name: ccadmin, password: password, tag: admin}
    - {name: uaaadmin, password: password, tag: admin}
  dea_next:
    advertise_interval_in_seconds: 5
    heartbeat_interval_in_seconds: 10
    memory_mb: 33996
  description: Cloud Foundry sponsored by Pivotal
  domain: system.cf.local
  etcd:
    machines: [10.0.4.103]
    peer_require_ssl: false
    require_ssl: false
  hm9000:
    url: https://hm9000.system.cf.local
    port: 5155
  logger_endpoint:
    port: 4443
  loggregator:
    maxRetainedLogMessages: 100
    debug: false
    outgoing_dropsonde_port: 8081
    etcd:
      machines: [10.0.4.103]
  loggregator_endpoint:
    shared_secret: password
  metron_agent:
    zone: z1
    deployment: minimal-aws
  metron_endpoint:
    shared_secret: password
  nats:
    machines: [10.0.4.103]
    password: password
    port: 4222
    user: nats
  ssl:
    skip_cert_verify: true
  system_domain: system.cf.local
  system_domain_organization: default_organization
  uaa:
    clients:
      cc_routing:
        authorities: routing.router_groups.read
        authorized-grant-types: client_credentials
        secret: password
      cf:
        id: cf
        override: true
        authorized-grant-types: implicit,password,refresh_token
        scope: cloud_controller.read,cloud_controller.write,openid,password.write,cloud_controller.admin,scim.read,scim.write,doppler.firehose,uaa.user,routing.router_groups.read
        authorities: uaa.none
        access-token-validity: 600
        refresh-token-validity: 2592000
        autoapprove: true
      cloud_controller_username_lookup:
        authorities: scim.userids
        authorized-grant-types: client_credentials
        secret: password
      doppler:
        authorities: uaa.resource
        override: true
        secret: password
      gorouter:
        authorities: routing.routes.read
        authorized-grant-types: client_credentials,refresh_token
        secret: password
      tcp_emitter:
        authorities: routing.routes.write,routing.routes.read
        authorized-grant-types: client_credentials,refresh_token
        secret: password
      tcp_router:
        authorities: routing.routes.read
        authorized-grant-types: client_credentials,refresh_token
        secret: password
      login:
        authorities: oauth.login,scim.write,clients.read,notifications.write,critical_notifications.write,emails.write,scim.userids,password.write
        authorized-grant-types: authorization_code,client_credentials,refresh_token
        override: true
        redirect-uri: https://login.system.cf.local
        scope: openid,oauth.approvals
        secret: password
      notifications:
        authorities: cloud_controller.admin,scim.read
        authorized-grant-types: client_credentials
        secret: password
    jwt:
      signing_key: |
        -----BEGIN RSA PRIVATE KEY-----
        <RSA PRIVATE KEY>
        -----END RSA PRIVATE KEY-----
      verification_key: |
        -----BEGIN PUBLIC KEY-----
		<PUBLIC KEY>
        -----END PUBLIC KEY-----
    ssl:
      port: -1
    url: https://uaa.system.cf.local
# code_snippet cf-minimal-aws end
# The previous line helps maintain current documentation at http://docs.cloudfoundry.org.
```

After setting up the manifest, all you need to do is to run:

```
bosh deploy
```

If all goes well, your environment should now look like this:
![Diagram 3](./img/cf-installation-3.jpg)

###Further Reading
For more information on how to set up RackHD: [rackhd.readthedocs.org](rackhd.readthedocs.org)

###Questions
There is a slack channel. 

- Slack Organization: `cloudfoundry.slack.com`
- Slack Channel: `#bosh-rackhd-cpi`

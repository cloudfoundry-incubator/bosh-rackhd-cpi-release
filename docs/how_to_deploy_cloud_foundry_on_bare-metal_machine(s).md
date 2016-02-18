###How to deploy Cloud Foundry on Bare-Metal Machine(s)
Deploying Cloud Foundry on Bare-Metal Machine(s) require 3 steps:

1. Set up RackHD
2. Set up Bosh Director with RackHD CPI
3. Deploying Cloud Foundry

In this tutorial, we will do the example with vSphere, but the same steps can be applied with other IaaS or Bare-Metal environment. Let's begin!

###Setting up RackHD
The first step is to set up [RackHD](https://github.com/rackhd), an open source solution that manage your Bare-Metal environment. Please reference the [RackHD documentation](http://rackhd.readthedocs.org/) if you would like to learn more about RackHD. 

In a nutshell, RackHD provides DHCP service and picks up iPXE boot signal from Bare-Metal machines. Once RackHD receives signal from machines, a new entry will be added to the RackHD database. RackHD will then be able to fully control the machines through IPMI and AMT. (See diagram 1)

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
sudo apt­-get ­-y updatesudo apt­-get ­-y dist­upgrade sudo apt-­get -­y autoremovesudo apt­-get -­y install nodejs nodejs­legacy npm#runtime dependenciessudo apt-­get -­y install rabbitmq­server mongodb isc­dhcp­serversudo apt-­get ­-y install snmp ipmitool ansible amtterm apt­mirror libkrb5­dev unzip#Ubuntu 15.04 or later: use upstart instead of systemd sudo apt­-get -­y install upstart­sysvsudo update-­initramfs ­-u
# AMTTool TNGwget http://downloads.sourceforge.net/project/amttool­tng/1.7/amttool sudo chmod 755 ./amttoolsudo mv amttool /usr/bin/amttool­tng
#compile dependenciessudo apt­-get -­y install git openssh­server pbuilder dh­make ubuntu­dev­tools devscripts
```

Now that the prerequisites are installed, we are ready to build RackHD from the latest source and use the `HWIMOBUILD` script to make debian packages for the components of RackHD. Then we install each of the debian packages by `dpkg`. 

```
RACKHD_INSTALL_DIR=~; cd $RACKHD_INSTALL_DIR 
git clone https://github.com/RackHD/RackHD 
RACKHD_PROJECT_DIR=${RACKHD_INSTALL_DIR}/RackHD
cd $RACKHD_PROJECT_DIRgit submodule update ­­init ­­recursivegit submodule foreach git pull origin mastersudo touch /etc/default/on­httpsudo touch /etc/default/on­dhcp­proxy 
sudo touch /etc/default/on­taskgraph 
sudo touch /etc/default/on­syslog 
sudo touch /etc/default/on­tftpcd ${RACKHD_PROJECT_DIR}/on­http 
./HWIMO­BUILDsudo dpkg -­i ./on­http_*.deb
cd ${RACKHD_PROJECT_DIR}/on­dhcp­proxy ./HWIMO­BUILDsudo dpkg ­i ./on­dhcp­proxy_*.deb
cd ${RACKHD_PROJECT_DIR}/on­taskgraph 
./HWIMO­BUILDsudo dpkg ­i ./on­taskgraph_*.deb
cd ${RACKHD_PROJECT_DIR}/on­syslog 
./HWIMO­BUILDsudo dpkg ­i ./on­syslog_*.deb
cd ${RACKHD_PROJECT_DIR}/on­tftp 
./HWIMO­BUILDsudo dpkg ­i ./on­tftp_*.debsudo reboot```

After reboot, you should have a RackHD server running. If you run into problems during the RackHD installation process, don't panic! There is a slack channel that can help you out. The RackHD community is very responsive and help! You can get invitation to the Slack channel by coming [here](http://community.emccode.com).

Now that we have the RackHD server ready. We are ready to discover our nodes. At the time of writing, RackHD can only be accessed through RESTful web services and we have developed a command line for our day-to-day activities. 

### Setting up RackHD CLI
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

### Setting up Nodes on RackHD server
Now that we have the RackHD server and the command line tool installed, we can start to TURN ON your machines. Make sure your machines has IPMI or other similar protocols (ie. AMT) that RackHD supports and have `Network Boot` turned on. 

You can run `watch rack nodes` and then turn on your machines. As the machines register themselves to RackHD, you should be able to set nodes being added, like this:

```
ID                       | NAME                      | CID | STATUS | DISK CID | ACTIVE WORKFLOW
-------------------------|---------------------------|-----|--------|----------|----------------
56c4a84de4dc603837faa636 | c0:3f:d5:60:51:b9 (node1) | n/a | n/a    | n/a      | n/a
56c4a84de4dc603837faa637 | c0:3f:d5:63:fd:b0 (node0) | n/a | n/a    | n/a      | n/a
56c4a861e4dc603837faa638 | c0:3f:d5:63:fe:13 (node2) | n/a | n/a    | n/a      | n/a
```

### Setting up Bosh Director with RackHD CPI
We can use `bosh-init` command line tool to install a bosh director in one of your nodes or in a virtual environment. In this example, we will be using `bosh-init` to deploy a bosh director in one of the Bare-Metal nodes. For more information about bosh-init, you can reference the tutorial [here](https://bosh.io/docs/using-bosh-init.html).

Let's follow these steps:

1. [Install bosh-init command line tool](https://bosh.io/docs/install-bosh-init.html)
2. [Install bosh command line tool](https://bosh.io/docs/bosh-cli.html)
3. [Download OpenStack Stemcell](http://bosh.io/stemcells/bosh-openstack-kvm-ubuntu-trusty-go_agent)

Once you have `bosh` and `bosh-init` set up, we can start setting up the Bosh Director!



###Further Reading
For more information on how to set up RackHD: 

###Questions
There is a slack channel. 

- Slack Organization: `cloudfoundry.slack.com`
- Slack Channel: `#bosh-rackhd-cpi`
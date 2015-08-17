package vm

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type SSHVM struct {
	id string

	sshClient       SSHClient
	agentEnvService AgentEnvService

	exists bool

	logger boshlog.Logger
}

func NewSSHVM(
	id string,
	sshClient SSHClient,
	agentEnvService AgentEnvService,
	logger boshlog.Logger,
	exists bool,
) SSHVM {
	return SSHVM{
		id: id,

		sshClient:       sshClient,
		agentEnvService: agentEnvService,

		exists: exists,

		logger: logger,
	}
}

func (vm SSHVM) ID() string { return vm.id }

func (vm SSHVM) SetUpStemcell(path string) error {
	err := vm.stopAgentAndRemoveStemcell()
	if err != nil {
		return bosherr.WrapError(err, "Deleting stemcell")
	}

	err = vm.sshClient.UploadStemcell(path, "/var/stemcell")
	if err != nil {
		return bosherr.WrapError(err, "Uploading stemcell tarball")
	}

	return nil
}

func (vm SSHVM) StartAgent() error {
	// todo leaking stemcell location
	prepScript := `#!/bin/bash
set -e -x
path=/tmp/baremetal_cpi-cgroups

mkdir -p $path
mountpoint -q $path || mount -t cgroup none $path

mkdir -p $path/chroot
cat $path/cpuset.cpus > $path/chroot/cpuset.cpus
cat $path/cpuset.mems > $path/chroot/cpuset.mems

# todo mounts are not idempotent
mkdir -p /var/stemcell/dev
mountpoint -q /var/stemcell/dev     || mount -n --bind /dev /var/stemcell/dev
mountpoint -q /var/stemcell/dev/pts || mount -n --bind /dev/pts /var/stemcell/dev/pts

mkdir -p /var/stemcell/proc
mountpoint -q /var/stemcell/proc || mount -n --bind /proc /var/stemcell/proc

# Make sure ssh service is not killed
rm -f /var/stemcell/etc/service/ssh
`

	err := vm.sshClient.ExecuteScript("start-runit", prepScript)
	if err != nil {
		return bosherr.WrapError(err, "Preparing for BOSH Agent")
	}

	// Places everything into a cgroup
	startScript := `#!/bin/bash
set -e -x
path=/tmp/baremetal_cpi-cgroups
echo $$ > $path/chroot/tasks
exec chroot /var/stemcell env -i \
	PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
	bash -e -c "/usr/sbin/runsvdir-start"
`

	err = vm.sshClient.ExecuteScript("start-runit", startScript, true)
	if err != nil {
		return bosherr.WrapError(err, "Running BOSH Agent")
	}

	return nil
}

func (vm SSHVM) Delete() error {
	return vm.stopAgentAndRemoveStemcell()
}

func (vm SSHVM) stopAgentAndRemoveStemcell() error {
	// Kills everything in a cgroup
	killScript := `#!/bin/bash
set -e -x
path=/tmp/baremetal_cpi-cgroups
if [ -d $path ]; then
	pids=$(cat $path/chroot/tasks)
	if [ ! -z "$pids" ]; then
		kill $pids
		sleep 5
		kill -9 $pids
	fi
	until [ -z "$(cat $path/chroot/tasks)" ]; do sleep 0.1; done
fi

mountpoint -q /var/stemcell/dev/pts && umount /var/stemcell/dev/pts
mountpoint -q /var/stemcell/dev     && umount /var/stemcell/dev
mountpoint -q /var/stemcell/proc    && umount /var/stemcell/proc

rm -rf /var/stemcell
`

	return vm.sshClient.ExecuteScript("kill-everything", killScript)
}

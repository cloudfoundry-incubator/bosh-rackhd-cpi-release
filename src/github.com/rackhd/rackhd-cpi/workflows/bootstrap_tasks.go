package workflows

import "github.com/rackhd/rackhd-cpi/rackhdapi"

const (
	SetPxeBootTaskName      = "Task.Obm.Node.PxeBoot"
	RebootNodeTaskName      = "Task.Obm.Node.Reboot"
	BootstrapUbuntuTaskName = "Task.Linux.Bootstrap.Ubuntu"
)

var bootstrapUbuntuTaskTemplate = []byte(`
	{
	  "friendlyName": "Bootstrap Ubuntu",
	  "injectableName": "Task.Linux.Bootstrap.Ubuntu",
	  "implementsTask": "Task.Base.Linux.Bootstrap",
	  "options": {
			"kernelFile": "vmlinuz-3.13.0-32-generic",
			"initrdFile": "initrd.img-3.13.0-32-generic",
			"kernelUri": "{{ api.server }}/2.0/{{ options.kernelFile }}",
			"initrdUri": "{{ api.server }}/2.0/{{ options.initrdFile }}",
			"basefs": "common/base.trusty.3.13.0-32-generic.squashfs.img",
			"overlayfs": "common/discovery.overlay.cpio.gz",
			"profile": "linux.ipxe",
			"comport": "ttyS0"
	  },
	  "properties": {
	    "os": {
	      "linux": {
	        "distribution": "ubuntu",
	        "release": "trusty",
	        "kernel": "3.13.0-32-generic"
	      }
	    }
	  }
	}`)

var setPxeBootTemplate = []byte(`
	{
  	"friendlyName": "Set Node Pxeboot",
  	"implementsTask": "Task.Base.Obm.Node",
  	"injectableName": "Task.Obm.Node.PxeBoot",
  	"options": {
    	"action": "setBootPxe",
    	"obmServiceName": "ipmi-obm-service"
  	},
  	"properties": {
    	"power": {}
   	}
	}`)

var rebootNodeTemplate = []byte(`
	{
	  "friendlyName": "Reboot Node",
	  "implementsTask": "Task.Base.Obm.Node",
	  "injectableName": "Task.Obm.Node.Reboot",
	  "options": {
	    "action": "reboot",
	    "obmServiceName": "ipmi-obm-service"
	  },
	  "properties": {
	    "power": {
	      "state": "reboot"
	    }
	  }
	}`)

type bootstrapUbuntuTaskOptions struct {
	Kernelversion string `json:"kernelversion"`
	Kernel        string `json:"kernel"`
	Initrd        string `json:"initrd"`
	Basefs        string `json:"basefs"`
	Overlayfs     string `json:"overlayfs"`
	Profile       string `json:"profile"`
	Comport       string `json:"comport"`
}

type bootstrapUbuntuTaskOptionsContainer struct {
	Options bootstrapUbuntuTaskOptions `json:"options"`
}

type bootstrapUbuntuTaskProperties struct {
	OS boostrapUbuntuTaskOsProperties `json:"os"`
}

type boostrapUbuntuTaskOsProperties struct {
	Linux boostrapUbuntuTaskOsLinuxProperties `json:"linux"`
}

type boostrapUbuntuTaskOsLinuxProperties struct {
	Distribution string `json:"distribution"`
	Release      string `json:"release"`
	Kernel       string `json:"kernel"`
}

type bootstrapUbuntuTaskPropertiesContainer struct {
	Properties bootstrapUbuntuTaskProperties `json:"properties"`
}

type bootstrapUbuntuTask struct {
	*rackhdapi.TaskStub
	*bootstrapUbuntuTaskOptionsContainer
	*bootstrapUbuntuTaskPropertiesContainer
}

type obmServiceOptions struct {
	Action         string `json:"action"`
	ObmServiceName string `json:"obmServiceName"`
}

type setPxeBootTaskOptionsContainer struct {
	Options obmServiceOptions `json:"options"`
}

type setPxeBootTaskProperties struct {
	Power map[string]string `json:"power"`
}

type setPxeBootTaskPropertiesContainer struct {
	Properties setPxeBootTaskProperties `json:"properties"`
}

type setPxeBootTask struct {
	*rackhdapi.TaskStub
	*setPxeBootTaskOptionsContainer
	*setPxeBootTaskPropertiesContainer
}

type rebootNodeTaskOptionsContainer struct {
	Options obmServiceOptions `json:"options"`
}

type rebootNodeTaskProperties struct {
	Power map[string]string `json:"power"`
}

type rebootNodeTaskPropertiesContainer struct {
	Properties rebootNodeTaskProperties `json:"properties"`
}

type rebootNodeTask struct {
	*rackhdapi.TaskStub
	*rebootNodeTaskOptionsContainer
	*rebootNodeTaskPropertiesContainer
}

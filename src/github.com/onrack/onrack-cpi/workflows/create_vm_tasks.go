package workflows

func BootstrapUbuntuTaskTemplate() []byte { return bootstrapUbuntuTaskTemplate }

var bootstrapUbuntuTaskTemplate []byte = []byte(`
	{
	  "friendlyName": "Boot into microkernel",
	  "injectableName": "Task.Linux.Bootstrap.Ubuntu",
	  "implementsTask": "Task.Base.Linux.Bootstrap",
	  "options": {
	    "kernelversion": "vmlinuz-3.13.0-32-generic",
	    "kernel": "common/vmlinuz-3.13.0-32-generic",
	    "initrd": "common/initrd.img-3.13.0-32-generic",
	    "basefs": "common/base.trusty.3.13.0-32.squashfs.img",
	    "overlayfs": "common/overlayfs_all_files.cpio.gz",
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

type bootstrapUbuntuTaskTemplateOptions struct {
}

func SetPxeBootTemplate() []byte { return setPxeBootTemplate }

var setPxeBootTemplate []byte = []byte(`
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

func RebootNodeTemplate() []byte { return rebootNodeTemplate }

var rebootNodeTemplate []byte = []byte(`
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

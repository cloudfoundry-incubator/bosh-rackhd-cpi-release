package workflows

// Required Tasks
const (
	BootstrapUbuntuTaskName         string = "Task.Linux.Bootstrap.Ubuntu"
	BootstrapUbuntuTaskTemplatePath string = "../templates/rackhd_bootstrap_ubuntu_task.json"
	RebootNodeTaskName              string = "Task.Obm.Node.Reboot"
	RebootNodeTaskTemplatePath      string = "../templates/rackhd_reboot_task.json"
	SetPxeRebootTaskName            string = "Task.Obm.Node.PxeBoot"
	SetPxeRebootTaskTemplatePath    string = "../templates/rackhd_set-pxe-boot_task.json"
)

// GetRequiredTasks returns all required tasks name
func GetRequiredTasks() map[string]string {
	return map[string]string{
		BootstrapUbuntuTaskName: BootstrapUbuntuTaskTemplatePath,
		RebootNodeTaskName:      RebootNodeTaskTemplatePath,
		SetPxeRebootTaskName:    SetPxeRebootTaskTemplatePath,
	}
}

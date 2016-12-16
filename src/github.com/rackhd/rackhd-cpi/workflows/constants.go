package workflows

// Generated Bosh Graphs
const (
	ReserveGraphName         string = "Graph.BOSH.Node.Reserve"
	ReserveGraphTemplatePath string = "../templates/reserve_node_workflow.json"

	ProvisionGraphName         string = "Graph.BOSH.Node.Provision"
	ProvisionGraphTemplatePath string = "../templates/provision_node_workflow.json"

	DeprovisionGraphName         string = "Graph.BOSH.Node.Deprovision"
	DeprovisionGraphTemplatePath string = "../templates/deprovision_node_workflow.json"
)

// Generated Bosh Tasks
const (
	ReserveTaskName         string = "Task.BOSH.Node.Reserve"
	ReserveTaskTemplatePath string = "../templates/reserve_node_task.json"

	ProvisionTaskName         string = "Task.BOSH.Node.Provision"
	ProvisionTaskTemplatePath string = "../templates/provision_node_task.json"

	DeprovisionTaskName         string = "Task.BOSH.Node.Deprovision"
	DeprovisionTaskTemplatePath string = "../templates/deprovision_node_task.json"

	SetIDTaskName         string = "Task.BOSH.Node.SetCID"
	SetIDTaskTemplatePath string = "../templates/set_id_task.json"
)

// Required RackHD Tasks
const (
	BootstrapUbuntuTaskName         string = "Task.Linux.Bootstrap.Ubuntu"
	BootstrapUbuntuTaskTemplatePath string = "../templates/rackhd_bootstrap_ubuntu_task.json"

	RebootNodeTaskName         string = "Task.Obm.Node.Reboot"
	RebootNodeTaskTemplatePath string = "../templates/rackhd_reboot_task.json"

	SetPxeRebootTaskName         string = "Task.Obm.Node.PxeBoot"
	SetPxeRebootTaskTemplatePath string = "../templates/rackhd_set-pxe-boot_task.json"
)

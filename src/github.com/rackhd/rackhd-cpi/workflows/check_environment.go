package workflows



// GetRequiredTasks returns all required tasks name
func GetRequiredTasks() map[string]string {
	return map[string]string{
		BootstrapUbuntuTaskName: BootstrapUbuntuTaskTemplatePath,
		RebootNodeTaskName:      RebootNodeTaskTemplatePath,
		SetPxeRebootTaskName:    SetPxeRebootTaskTemplatePath,
	}
}

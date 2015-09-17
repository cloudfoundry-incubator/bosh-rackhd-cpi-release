package workflows_test

import (
	"encoding/json"
	"fmt"
	"github.com/onrack/onrack-cpi/workflows"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackhttp"

	"io/ioutil"
	"net/http"
	"github.com/nu7hatch/gouuid"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func expectedWorkflow(uuid string) (expectedWorkflow onrackhttp.Workflow){
	expectedWorkflow = onrackhttp.Workflow{
		FriendlyName: "CF CreateReserve VM",
		InjectableName: fmt.Sprintf("Graph.CF.CreateReserveVM.%s", uuid),
		Options: onrackhttp.Options{
				BootstrapUbuntu: map[string]string{"overlayfs": "common/overlayfs_all_files.cf.cpio.gz"},
				Defaults: onrackhttp.Defaults{
					AgentSettingsFile: "",
					AgentSettingsPath: "",
					Cid: "",
					DownloadDir: "/opt/downloads",
					RegistrySettingsFile: "",
					RegistrySettingsPath: "",
					StemcellFile: "",
				},
		},
		Tasks: []onrackhttp.WorkflowTask{
			onrackhttp.WorkflowTask{
				Label: "set-boot-pxe",
				TaskName: "Task.Obm.Node.PxeBoot",
				IgnoreFailure: true,
			},
			onrackhttp.WorkflowTask{
				Label: "reboot",
				TaskName: "Task.Obm.Node.Reboot",
				WaitOn: map[string]string{
					"set-boot-pxe": "finished",
				},
			},
			onrackhttp.WorkflowTask{
				Label: "bootstrap-ubuntu",
				TaskName: "Task.Linux.Bootstrap.Ubuntu",
				WaitOn: map[string]string{
					"reboot": "succeeded",
				},
			},
			onrackhttp.WorkflowTask{
				Label: "reserve-node",
				TaskName: fmt.Sprintf("Task.Os.Reserve.CF.VM.%s", uuid),
				WaitOn: map[string]string{
					"bootstrap-ubuntu": "succeeded",
				},
			},
			onrackhttp.WorkflowTask{
				Label: "provision-node",
				TaskName: fmt.Sprintf("Task.Os.Install.CF.Stemcell.%s", uuid),
				WaitOn: map[string]string{
					"reserve-node": "succeeded",
				},
			},
			onrackhttp.WorkflowTask{
				Label: "shell-reboot",
				TaskName: "Task.ProcShellReboot",
				WaitOn: map[string]string{
					"provision-node": "finished",
				},
			},
		},
	}
	return
}

func expectedReserveNodeTask(uuid string) (expectedReserveNodeTask onrackhttp.Task){
	expectedReserveNodeTask = onrackhttp.Task{
		FriendlyName: "Provision Node",
		ImplementsTask: "Task.Base.Linux.Commands",
		InjectableName: fmt.Sprintf("Task.Os.Install.CF.Stemcell.%s", uuid),
		Options: map[string]interface{} {
			"agentSettingsFile": "",
			"agentSettingsMd5Uri": "{{ api.files }}/md5/{{ options.agentSettingsFile }}/latest",
			"agentSettingsPath": "",
			"agentSettingsUri": "{{ api.files }}/{{ options.agentSettingsFile }}/latest",
			"device": "/dev/sda",
      "downloadDir": "/opt/downloads",
      "registrySettingsFile": "",
      "registrySettingsMd5Uri": "{{ api.files }}/md5/{{ options.registrySettingsFile }}/latest",
      "registrySettingsPath": "",
      "registrySettingsUri": "{{ api.files }}/{{ options.registrySettingsFile }}/latest",
      "stemcellFile": "",
      "stemcellFileMd5Uri": "{{ api.files }}/md5/{{ options.stemcellFile }}/latest",
      "stemcellUri": "{{ api.files }}/{{ options.stemcellFile }}/latest",
		},
		Properties: map[string]interface{}{},
	}
	return
}

func expectedProvisionNodeTask(uuid string) (expectedProvisionNodeTask onrackhttp.Task){
	expectedProvisionNodeTask = onrackhttp.Task{
		FriendlyName: "Reserve Node",
		ImplementsTask: "Task.Base.Linux.Commands",
		InjectableName: fmt.Sprintf("Task.Os.Reserve.CF.VM.%s", uuid),
		Options: map[string]interface{} {
			"cid": nil,
		},
		Properties: map[string]interface{}{},
	}
	return
}

var _ = Describe("Workflows", func() {
	var apiServer string
	var cpiConfig config.Cpi
	var fakeUUID *uuid.UUID
	var nodeID string

	BeforeEach(func (){
	  apiServer = os.Getenv("ON_RACK_API_URI")
	  cpiConfig = config.Cpi{ApiServer: apiServer}
	  fakeUUID, _ = uuid.NewV4()
	  nodeID = os.Getenv("ON_RACK_NODE_ID")
	})

	Describe("PublishCreateVMWorkflow", func (){
		It("returns", func(){
			err := workflows.PublishCreateVMWorkflow(cpiConfig, fakeUUID.String())
			Expect(err).ToNot(HaveOccurred())

			url := fmt.Sprintf("http://%s:8080/api/1.1/workflows/library", cpiConfig.ApiServer)
			resp, err := http.Get(url)
			Expect(err).ToNot(HaveOccurred())

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			var workflowLibrary []onrackhttp.Workflow
			err = json.Unmarshal(body, &workflowLibrary)
			Expect(err).ToNot(HaveOccurred())

			expectedWorkflow := expectedWorkflow(fakeUUID.String())
			taskLibrary, err := onrackhttp.RetrieveTasks(cpiConfig)
			Expect(err).ToNot(HaveOccurred())

			for i := range taskLibrary {
				delete(taskLibrary[i].Options, "commands")
			}
			Expect(taskLibrary).To(ContainElement(expectedProvisionNodeTask(fakeUUID.String())))
			Expect(taskLibrary).To(ContainElement(expectedReserveNodeTask(fakeUUID.String())))

			Expect(workflowLibrary).To(ContainElement(expectedWorkflow))
		})
	})
})

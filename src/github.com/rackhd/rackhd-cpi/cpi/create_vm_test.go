package cpi

/*
	Rather than creating a separate cpi_test package, this suite is part of
	the cpi package itself in order to test node selection functionality without
	exporting these methods for testing. Please be careful as this suite will have
	access to all unexported functions and variables in the cpi package. You have
	been warned

	- The ghost in the air ducts
*/

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
)

func loadNodes(nodePath string) []rackhdapi.Node {
	dummyResponseFile, err := os.Open(nodePath)
	Expect(err).ToNot(HaveOccurred())
	defer dummyResponseFile.Close()

	dummyResponseBytes, err := ioutil.ReadAll(dummyResponseFile)
	Expect(err).ToNot(HaveOccurred())

	nodes := []rackhdapi.Node{}
	err = json.Unmarshal(dummyResponseBytes, &nodes)
	Expect(err).ToNot(HaveOccurred())

	return nodes
}

func loadWorkflowsResponse(assetPath string) []byte {
	dummyResponseFile, err := os.Open(assetPath)
	Expect(err).ToNot(HaveOccurred())
	defer dummyResponseFile.Close()

	workflowsResponse, err := ioutil.ReadAll(dummyResponseFile)
	Expect(err).ToNot(HaveOccurred())

	return workflowsResponse
}

func loadNodeCatalog(nodeCatalogPath string) rackhdapi.NodeCatalog {
	dummyCatalogfile, err := os.Open(nodeCatalogPath)
	Expect(err).ToNot(HaveOccurred())
	defer dummyCatalogfile.Close()

	b, err := ioutil.ReadAll(dummyCatalogfile)
	Expect(err).ToNot(HaveOccurred())

	nodeCatalog := rackhdapi.NodeCatalog{}

	err = json.Unmarshal(b, &nodeCatalog)
	Expect(err).ToNot(HaveOccurred())
	return nodeCatalog
}

var _ = Describe("The VM Creation Workflow", func() {
	var server *ghttp.Server
	var jsonReader *strings.Reader
	var cpiConfig config.Cpi

	BeforeEach(func() {
		server = ghttp.NewServer()
		serverURL, err := url.Parse(server.URL())
		Expect(err).ToNot(HaveOccurred())
		jsonReader = strings.NewReader(fmt.Sprintf(`{"apiserver":"%s", "agent":{"blobstore": {"provider":"local","some": "options"}, "mbus":"localhost"}, "max_create_vm_attempts":1}`, serverURL.Host))
		cpiConfig, err = config.New(jsonReader)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("parsing director input", func() {
		It("return the fields given a valid config", func() {
			jsonInput := []byte(`[
				"4149ba0f-38d9-4485-476f-1581be36f290",
				"vm-478585",
				{"public_key": "MTIzNA=="},
				{
						"private": {
								"type": "dynamic"
						}
				},
				[],
				{}]`)
			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())
			agentID, vmCID, publicKey, networks, err := parseCreateVMInput(extInput)
			Expect(agentID).To(Equal("4149ba0f-38d9-4485-476f-1581be36f290"))
			Expect(vmCID).To(Equal("vm-478585"))
			Expect(publicKey).To(Equal("1234"))
			Expect(networks).ToNot(BeEmpty())
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error if passed an unexpected type for network configuration", func() {
			jsonInput := []byte(`[
				"4149ba0f-38d9-4485-476f-1581be36f290",
				"vm-478585",
				{},
				"aint-gonna-work-network",
				[],
				{}]`)

			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())

			_, _, _, _, err = parseCreateVMInput(extInput)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("network config has unexpected type in: string. Expecting a map"))
		})

		It("returns an error if more than one network is provided", func() {
			jsonInput := []byte(`[
    		"4149ba0f-38d9-4485-476f-1581be36f290",
    		"vm-478585",
    		{},
    		{
        		"private": {
            		"type": "dynamic"
        		},
        		"private2": {
            		"type": "dynamic"
        		}
    		},
    		[],
    		{}]`)

			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())

			_, _, _, _, err = parseCreateVMInput(extInput)
			Expect(err).To(MatchError("config error: Only one network supported, provided length: 2"))
		})

		It("defaults to manual network if network type is not defined", func() {
			jsonInput := []byte(`[
				"4149ba0f-38d9-4485-476f-1581be36f290",
				"vm-478585",
				{},
				{
					"private": {
						"ip": "10.0.0.2",
						"netmask": "255.255.255.0",
						"gateway": "10.0.0.1"
					}
				},
				[],
				{}]`)

			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())

			_, _, _, networks, err := parseCreateVMInput(extInput)
			Expect(err).ToNot(HaveOccurred())
			Expect(networks).ToNot(BeEmpty())
			Expect(networks["private"].NetworkType).To(Equal(bosh.ManualNetworkType))
		})

		It("returns an error if Agent ID is empty", func() {
			jsonInput := []byte(`[
    		"",
    		"vm-478585",
    		{},
    		{
        		"private": {
            		"type": "dynamic"
        		},
        		"private2": {
            		"type": "dynamic"
        		}
    		},
    		[],
    		{}]`)

			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())

			_, _, _, _, err = parseCreateVMInput(extInput)
			Expect(err).To(MatchError("agent id cannot be empty"))
		})

		It("returns an error if Agent ID is of an unexpected type", func() {
			jsonInput := []byte(`[
				{},
				"vm-478585",
				{},
				{
						"private": {
								"type": "dynamic"
						},
						"private2": {
								"type": "dynamic"
						}
				},
				[],
				{}]`)

			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())

			_, _, _, _, err = parseCreateVMInput(extInput)
			Expect(err).To(MatchError("agent id has unexpected type: map[string]interface {}. Expecting a string"))
		})

		It("return an error if public key is of an unexpected type", func() {
			jsonInput := []byte(`[
				"4149ba0f-38d9-4485-476f-1581be36f290",
				"vm-478585",
				{"public_key": 1234},
				{
						"private": {
								"type": "dynamic"
						}
				},
				[],
				{}]`)
			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())
			_, _, _, _, err = parseCreateVMInput(extInput)
			Expect(err).To(MatchError("public key has unexpected type: float64. Expecting a string"))
		})

		Context("when specifying manual networking", func() {
			It("returns an error if IP is not set", func() {
				jsonInput := []byte(`[
				"4149ba0f-38d9-4485-476f-1581be36f290",
				"vm-478585",
				{},
				{
						"private": {
								"type": "manual",
								"netmask": "255.255.255.0",
								"gateway": "10.0.0.1"
						}
				},
				[],
				{}]`)

				var extInput bosh.MethodArguments
				err := json.Unmarshal(jsonInput, &extInput)
				Expect(err).ToNot(HaveOccurred())

				_, _, _, _, err = parseCreateVMInput(extInput)
				Expect(err).To(MatchError("config error: ip must be specified for manual network"))
			})

			It("returns an error if gateway is not set", func() {
				jsonInput := []byte(`[
				"4149ba0f-38d9-4485-476f-1581be36f290",
				"vm-478585",
				{},
				{
						"private": {
								"type": "manual",
								"netmask": "255.255.255.0",
								"ip": "10.0.0.5"
						}
				},
				[],
				{}]`)

				var extInput bosh.MethodArguments
				err := json.Unmarshal(jsonInput, &extInput)
				Expect(err).ToNot(HaveOccurred())

				_, _, _, _, err = parseCreateVMInput(extInput)
				Expect(err).To(MatchError("config error: gateway must be specified for manual network"))
			})

			It("returns an error if netmask is not set", func() {
				jsonInput := []byte(`[
				"4149ba0f-38d9-4485-476f-1581be36f290",
				"vm-478585",
				{},
				{
						"private": {
								"type": "manual",
								"ip": "10.0.0.4",
								"gateway": "10.0.0.1"
						}
				},
				[],
				{}]`)

				var extInput bosh.MethodArguments
				err := json.Unmarshal(jsonInput, &extInput)
				Expect(err).ToNot(HaveOccurred())

				_, _, _, _, err = parseCreateVMInput(extInput)
				Expect(err).To(MatchError("config error: netmask must be specified for manual network"))
			})
		})
	})

	Describe("unreserving a node", func() {
		It("return a node with reserved flag unset", func() {
			apiServerIP := fmt.Sprintf("%s:8080", os.Getenv("RACKHD_API_URI"))
			Expect(apiServerIP).ToNot(BeEmpty())
			c := config.Cpi{ApiServer: apiServerIP}

			nodes, err := rackhdapi.GetNodes(c)
			Expect(err).ToNot(HaveOccurred())
			targetNodeID := nodes[0].ID
			log.Info(fmt.Sprintf("targetNodeId: %s", targetNodeID))
			err = rackhdapi.ReleaseNode(c, targetNodeID)
			Expect(err).ToNot(HaveOccurred())
			nodeURL := fmt.Sprintf("http://%s/api/common/nodes/%s", c.ApiServer, targetNodeID)

			resp, err := http.Get(nodeURL)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			nodeBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			var node rackhdapi.Node
			err = json.Unmarshal(nodeBytes, &node)
			Expect(err).ToNot(HaveOccurred())
			Expect(node.Status).To(Equal(rackhdapi.Available))
		})
	})

	Describe("trying to reserve a node without an ephemeral disk", func() {
		It("blocks the node", func() {
			expectedNodes := loadNodes("../spec_assets/dummy_two_node_response.json")
			expectedNodesData, err := json.Marshal(expectedNodes)
			Expect(err).ToNot(HaveOccurred())
			firstExpectedNodeCatalog := loadNodeCatalog("../spec_assets/dummy_no_ephemeral_disk_catalog_response.json")
			firstExpectedNodeCatalogData, err := json.Marshal(firstExpectedNodeCatalog)
			Expect(err).ToNot(HaveOccurred())
			secondExpectedNodeCatalog := loadNodeCatalog("../spec_assets/dummy_node_catalog_response.json")
			secondExpectedNodeCatalogData, err := json.Marshal(secondExpectedNodeCatalog)
			Expect(err).ToNot(HaveOccurred())

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/common/nodes"),
					ghttp.RespondWith(http.StatusOK, expectedNodesData),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", expectedNodes[0].ID)),
					ghttp.RespondWith(http.StatusOK, []byte("[]")),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/common/nodes/%s/catalogs/ohai", expectedNodes[0].ID)),
					ghttp.RespondWith(http.StatusOK, firstExpectedNodeCatalogData),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", fmt.Sprintf("/api/common/nodes/%s", expectedNodes[0].ID)),
					ghttp.VerifyJSON(fmt.Sprintf(`{"status": "%s", "status_reason": "%s"}`, "blocked", "Node has missing disks")),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", expectedNodes[1].ID)),
					ghttp.RespondWith(http.StatusOK, []byte("[]")),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/common/nodes/%s/catalogs/ohai", expectedNodes[1].ID)),
					ghttp.RespondWith(http.StatusOK, secondExpectedNodeCatalogData),
				),
			)

			_, err = tryReservation(
				cpiConfig,
				"agentID",
				blockNodesWithoutEphemeralDisk,
				func(config.Cpi) (string, error) { return "", nil },
				func(config.Cpi, string, string) error { return nil },
			)
		})
	})

	Describe("building the BOSH agent networking spec", func() {
		It("returns an error if no active networks can be found", func() {
			dummyCatalogfile, err := os.Open("../spec_assets/dummy_node_catalog_all_interface_down_response.json")
			Expect(err).ToNot(HaveOccurred())
			defer dummyCatalogfile.Close()

			b, err := ioutil.ReadAll(dummyCatalogfile)
			Expect(err).ToNot(HaveOccurred())

			nodeCatalog := rackhdapi.NodeCatalog{}

			err = json.Unmarshal(b, &nodeCatalog)
			Expect(err).ToNot(HaveOccurred())

			prevSpec := bosh.Network{}

			_, err = attachMAC(nodeCatalog.Data.NetworkData.Networks, prevSpec)
			Expect(err).To(MatchError("node has no active network"))
		})

		It("returns an error if multiple active networks are found", func() {
			dummyCatalogfile, err := os.Open("../spec_assets/dummy_node_catalog_multiple_interface_up_response.json")
			Expect(err).ToNot(HaveOccurred())
			defer dummyCatalogfile.Close()

			b, err := ioutil.ReadAll(dummyCatalogfile)
			Expect(err).ToNot(HaveOccurred())

			nodeCatalog := rackhdapi.NodeCatalog{}

			err = json.Unmarshal(b, &nodeCatalog)
			Expect(err).ToNot(HaveOccurred())

			prevSpec := bosh.Network{}

			_, err = attachMAC(nodeCatalog.Data.NetworkData.Networks, prevSpec)
			Expect(err).To(MatchError("node has 2 active networks"))
		})

		Context("when using manual networking", func() {
			It("copies the fields we are interested in passing on to the agent", func() {
				dummyCatalogfile, err := os.Open("../spec_assets/dummy_node_catalog_response.json")
				Expect(err).ToNot(HaveOccurred())
				defer dummyCatalogfile.Close()

				b, err := ioutil.ReadAll(dummyCatalogfile)
				Expect(err).ToNot(HaveOccurred())

				nodeCatalog := rackhdapi.NodeCatalog{}

				err = json.Unmarshal(b, &nodeCatalog)
				Expect(err).ToNot(HaveOccurred())

				prevSpec := bosh.Network{
					NetworkType: bosh.DynamicNetworkType,
					Netmask:     "255.255.255.0",
					Gateway:     "10.0.0.1",
					IP:          "10.0.0.5",
					Default:     []string{"dns", "gateway"},
					DNS:         []string{"8.8.8.8"},
				}

				newSpec, err := attachMAC(nodeCatalog.Data.NetworkData.Networks, prevSpec)
				Expect(err).ToNot(HaveOccurred())
				Expect(prevSpec.NetworkType).To(Equal(newSpec.NetworkType))
				Expect(prevSpec.Netmask).To(Equal(newSpec.Netmask))
				Expect(prevSpec.Gateway).To(Equal(newSpec.Gateway))
				Expect(prevSpec.IP).To(Equal(newSpec.IP))
				Expect(prevSpec.Default).To(Equal(newSpec.Default))
				Expect(prevSpec.DNS).To(Equal(newSpec.DNS))
				Expect(newSpec.CloudProperties).To(BeEmpty())

			})

			It("attaches MAC address information from the RackHD API", func() {
				nodeCatalog := loadNodeCatalog("../spec_assets/dummy_node_catalog_response.json")

				prevSpec := bosh.Network{}

				netSpec, err := attachMAC(nodeCatalog.Data.NetworkData.Networks, prevSpec)
				Expect(err).ToNot(HaveOccurred())
				Expect(netSpec.MAC).To(Equal("00:1e:67:c4:e1:a0"))
			})
		})
	})

	Context("when specifying dynamic networking", func() {
		It("creates the networking spec without cloud_properties", func() {
			jsonInput := []byte(`[
				"4149ba0f-38d9-4485-476f-1581be36f290",
				"vm-478585",
				{},
				{
						"private": {
								"type": "dynamic",
								"cloud_properties": { "option": "not-passed-to-agent" }
						}
				},
				[],
				{}]`)

			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())

			testSpec := bosh.Network{
				NetworkType: bosh.DynamicNetworkType,
			}
			_, _, _, netSpec, err := parseCreateVMInput(extInput)
			Expect(err).ToNot(HaveOccurred())
			Expect(netSpec).To(Equal(map[string]bosh.Network{"private": testSpec}))
		})
	})

	Describe("selecting an available node", func() {
		It("returns an error if there are no free nodes available", func() {
			nodes := loadNodes("../spec_assets/dummy_all_reserved_nodes_response.json")
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodes[0].ID)),
					ghttp.RespondWith(http.StatusOK, []byte("[]")),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodes[1].ID)),
					ghttp.RespondWith(http.StatusOK, []byte("[]")),
				),
			)

			_, err := randomSelectAvailableNode(cpiConfig, nodes)

			Expect(err).To(MatchError("all nodes have been reserved"))
		})

		It("selects a free node for provisioning", func() {
			nodes := loadNodes("../spec_assets/dummy_two_node_response.json")
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodes[0].ID)),
					ghttp.RespondWith(http.StatusOK, []byte("[]")),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodes[1].ID)),
					ghttp.RespondWith(http.StatusOK, []byte("[]")),
				),
			)

			rackHDID, err := randomSelectAvailableNode(cpiConfig, nodes)

			Expect(err).ToNot(HaveOccurred())
			Expect(rackHDID).To(Equal("55e79ea54e66816f6152fff9"))
		})

		It("return an error if all nodes are created vms with cids", func() {
			nodes := loadNodes("../spec_assets/dummy_all_nodes_are_vms.json")
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodes[0].ID)),
					ghttp.RespondWith(http.StatusOK, []byte("[]")),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodes[1].ID)),
					ghttp.RespondWith(http.StatusOK, []byte("[]")),
				),
			)

			_, err := randomSelectAvailableNode(cpiConfig, nodes)

			Expect(err).To(MatchError("all nodes have been reserved"))
		})
	})

	Describe("retrying node reservation", func() {
		It("return a node if selection is successful", func() {
			jsonReader := strings.NewReader(`{"apiserver":"localhost", "agent":{"blobstore": {"provider":"local","some": "options"}, "mbus":"localhost"}, "max_create_vm_attempts":3}`)
			c, err := config.New(jsonReader)
			Expect(err).ToNot(HaveOccurred())
			nodeID, err := tryReservation(
				c,
				"agentID",
				func(config.Cpi) error { return nil },
				func(config.Cpi) (string, error) { return "node-1234", nil },
				func(config.Cpi, string, string) error { return nil },
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(nodeID).To(Equal("node-1234"))
		})

		It("returns an error if selection continually fails", func() {
			jsonReader := strings.NewReader(`{"apiserver":"localhost", "agent":{"blobstore": {"provider":"local","some": "options"}, "mbus":"localhost"}, "max_create_vm_attempts":3}`)
			c, err := config.New(jsonReader)
			Expect(err).ToNot(HaveOccurred())

			nodeID, err := tryReservation(
				c,
				"agentID",
				func(config.Cpi) error { return nil },
				func(config.Cpi) (string, error) { return "node-1234", nil },
				func(config.Cpi, string, string) error { return errors.New("error") },
			)
			Expect(err).To(MatchError("unable to reserve node"))
			Expect(nodeID).To(Equal(""))
		})

		It("retries and eventually returns a node when selection is successful", func() {
			jsonReader := strings.NewReader(`{"apiserver":"localhost", "agent":{"blobstore": {"provider":"local","some": "options"}, "mbus":"localhost"}, "max_create_vm_attempts":3}`)
			c, err := config.New(jsonReader)
			Expect(err).ToNot(HaveOccurred())

			tries := 0
			flakeySelectionFunc := func(config.Cpi) (string, error) {
				if tries < 2 {
					tries++
					return "", errors.New("")
				}
				return "node-1234", nil
			}
			nodeID, err := tryReservation(
				c,
				"agentID",
				func(config.Cpi) error { return nil },
				flakeySelectionFunc,
				func(config.Cpi, string, string) error { return nil },
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(nodeID).To(Equal("node-1234"))
		})

		It("cleans up reservation flag after failing to reserve", func() {
			apiServerIP := fmt.Sprintf("%s:8080", os.Getenv("RACKHD_API_URI"))
			Expect(apiServerIP).ToNot(BeEmpty())
			jsonReader := strings.NewReader(fmt.Sprintf(`{"apiserver":"%s", "agent":{"blobstore": {"provider":"local","some": "options"}, "mbus":"localhost"}, "max_create_vm_attempts":1}`, apiServerIP))
			c, err := config.New(jsonReader)
			Expect(err).ToNot(HaveOccurred())

			var testNodeID string
			flakeyReservationFunc := func(c config.Cpi, agentID string, nodeID string) error {
				testNodeID = nodeID
				url := fmt.Sprintf("http://%s/api/common/nodes/%s", c.ApiServer, nodeID)

				reserveFlag := `{"status" : "reserved"}`
				body := ioutil.NopCloser(strings.NewReader(reserveFlag))
				defer body.Close()

				request, err := http.NewRequest("PATCH", url, body)
				Expect(err).ToNot(HaveOccurred())

				request.Header.Set("Content-Type", "application/json")
				request.ContentLength = int64(len(reserveFlag))

				resp, err := http.DefaultClient.Do(request)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(200))

				nodeURL := fmt.Sprintf("http://%s/api/common/nodes/%s", c.ApiServer, testNodeID)
				nodeResp, err := http.Get(nodeURL)
				Expect(err).ToNot(HaveOccurred())

				nodeBytes, err := ioutil.ReadAll(nodeResp.Body)
				Expect(err).ToNot(HaveOccurred())
				defer nodeResp.Body.Close()

				var node rackhdapi.Node
				err = json.Unmarshal(nodeBytes, &node)
				Expect(err).ToNot(HaveOccurred())
				Expect(node.Status).To(Equal(rackhdapi.Reserved))

				return errors.New("fake error doing reservation")
			}

			_, err = selectNodeFromRackHD(c)
			Expect(err).ToNot(HaveOccurred())

			_, err = tryReservation(
				c,
				"agentID",
				func(config.Cpi) error { return nil },
				selectNodeFromRackHD,
				flakeyReservationFunc,
			)

			Expect(err).To(MatchError("unable to reserve node"))
			nodeURL := fmt.Sprintf("http://%s/api/common/nodes/%s", c.ApiServer, testNodeID)
			resp, err := http.Get(nodeURL)
			Expect(err).ToNot(HaveOccurred())

			nodeBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()

			var node rackhdapi.Node
			err = json.Unmarshal(nodeBytes, &node)
			Expect(err).ToNot(HaveOccurred())
			Expect(node.Status).To(Equal(rackhdapi.Available))
		})
	})

	Describe("when a node has an active workflow", func() {
		It("skips the node", func() {
			rawWorkflow := loadWorkflowsResponse("../spec_assets/dummy_workflow_response.json")
			httpResponse := []byte(fmt.Sprintf("[%s]", string(rawWorkflow)))
			var expectedResponse []rackhdapi.WorkflowResponse
			err := json.Unmarshal(httpResponse, &expectedResponse)
			Expect(err).ToNot(HaveOccurred())

			nodeID := "5665a65a0561790005b77b85"
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodeID)),
					ghttp.RespondWith(http.StatusOK, httpResponse),
				),
			)

			nodes := loadNodes("../spec_assets/dummy_one_node_running_workflow.json")
			_, err = randomSelectAvailableNode(cpiConfig, nodes)
			Expect(err).To(MatchError("all nodes have been reserved"))
		})
	})
})

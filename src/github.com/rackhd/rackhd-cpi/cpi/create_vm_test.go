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
	"os"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/helpers"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
)

var _ = Describe("The VM Creation Workflow", func() {
	var server *ghttp.Server
	var jsonReader *strings.Reader
	var cpiConfig config.Cpi
	var request bosh.CpiRequest
	var allowFilter Filter

	BeforeEach(func() {
		server, jsonReader, cpiConfig, request = helpers.SetUp(bosh.CREATE_VM)

		allowFilter = Filter{
			data:   nil,
			method: AllowAnyNodeMethod,
		}
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
				["diskCID"],
				{}]`)
			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)

			Expect(err).ToNot(HaveOccurred())
			agentID, vmCID, publicKey, networks, diskCIDs, err := parseCreateVMInput(extInput)
			Expect(err).ToNot(HaveOccurred())
			Expect(agentID).To(Equal("4149ba0f-38d9-4485-476f-1581be36f290"))
			Expect(vmCID).To(Equal("vm-478585"))
			Expect(diskCIDs).To(Equal("diskCID"))
			Expect(publicKey).To(Equal("1234"))
			Expect(networks).ToNot(BeEmpty())
		})

		It("returns an error if passed an unexpected type for network configuration", func() {
			jsonInput := []byte(`[
				"4149ba0f-38d9-4485-476f-1581be36f290",
				"vm-478585",
				{},
				{
						"private": {
								"type": "dynamic"
						}
				},
				"aint gon work disks",
				{}]`)

			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())

			_, _, _, _, _, err = parseCreateVMInput(extInput)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("disk config has unexpected type in: string. Expecting an array"))
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

			_, _, _, _, _, err = parseCreateVMInput(extInput)
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

			_, _, _, _, _, err = parseCreateVMInput(extInput)
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

			_, _, _, networks, _, err := parseCreateVMInput(extInput)
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

			_, _, _, _, _, err = parseCreateVMInput(extInput)
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

			_, _, _, _, _, err = parseCreateVMInput(extInput)
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
			_, _, _, _, _, err = parseCreateVMInput(extInput)
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

				_, _, _, _, _, err = parseCreateVMInput(extInput)
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

				_, _, _, _, _, err = parseCreateVMInput(extInput)
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

				_, _, _, _, _, err = parseCreateVMInput(extInput)
				Expect(err).To(MatchError("config error: netmask must be specified for manual network"))
			})
		})
	})

	Describe("unreserving a node", func() {
		It("return a node with reserved flag unset", func() {
			apiServer, err := helpers.GetRackHDHost()
			Expect(err).ToNot(HaveOccurred())

			c := config.Cpi{ApiServer: apiServer}

			nodes, err := rackhdapi.GetNodes(c)
			Expect(err).ToNot(HaveOccurred())
			targetNodeID := nodes[0].ID
			log.Info(fmt.Sprintf("targetNodeId: %s", targetNodeID))
			err = rackhdapi.ReleaseNode(c, targetNodeID)
			Expect(err).ToNot(HaveOccurred())
			nodeURL := fmt.Sprintf("%s/api/common/nodes/%s", c.ApiServer, targetNodeID)

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
			Expect(err).To(MatchError("error attaching MAC address: node has no active network"))
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
			Expect(err).To(MatchError("error attaching MAC address: node has 2 active networks"))
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
				nodeCatalog := helpers.LoadNodeCatalog("../spec_assets/dummy_node_catalog_response.json")

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
			_, _, _, netSpec, _, err := parseCreateVMInput(extInput)
			Expect(err).ToNot(HaveOccurred())
			Expect(netSpec).To(Equal(map[string]bosh.Network{"private": testSpec}))
		})
	})

	Describe("selecting an available node", func() {
		It("returns an error if there are no free nodes available", func() {
			nodes := helpers.LoadNodes("../spec_assets/dummy_all_reserved_nodes_response.json")

			node0HttpResponse, err := json.Marshal(nodes[0])
			Expect(err).ToNot(HaveOccurred())
			node1HttpResponse, err := json.Marshal(nodes[1])
			Expect(err).ToNot(HaveOccurred())

			server.RouteToHandler("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodes[0].ID),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodes[0].ID)),
					ghttp.RespondWith(http.StatusNoContent, []byte{}),
				),
			)

			server.RouteToHandler("GET", fmt.Sprintf("/api/common/nodes/%s", nodes[0].ID),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/common/nodes/%s", nodes[0].ID)),
					ghttp.RespondWith(http.StatusOK, node0HttpResponse),
				),
			)

			server.RouteToHandler("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodes[1].ID),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodes[1].ID)),
					ghttp.RespondWith(http.StatusNoContent, []byte{}),
				),
			)

			server.RouteToHandler("GET", fmt.Sprintf("/api/common/nodes/%s", nodes[1].ID),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/common/nodes/%s", nodes[1].ID)),
					ghttp.RespondWith(http.StatusOK, node1HttpResponse),
				),
			)

			_, err = randomSelectAvailableNode(cpiConfig, nodes, allowFilter)
			Expect(err).To(MatchError("all nodes have been reserved"))
		})

		Context("with a disk CID", func() {
			It("selects the node with the disk CID", func() {
				nodes := helpers.LoadNodes("../spec_assets/dummy_create_vm_with_disk_response.json")

				nodesResponse, err := json.Marshal(nodes)
				Expect(err).ToNot(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("/api/common/nodes")),
						ghttp.RespondWith(http.StatusOK, nodesResponse),
					),
				)

				node, err := SelectNodeFromRackHD(cpiConfig, "disk-1234", allowFilter)

				Expect(err).ToNot(HaveOccurred())
				Expect(node.ID).To(Equal("5665a65a0561790005b77b85"))
			})
		})

		It("selects a free node for provisioning", func() {
			nodes := helpers.LoadNodes("../spec_assets/dummy_two_node_response.json")

			node0HttpResponse, err := json.Marshal(nodes[0])
			Expect(err).ToNot(HaveOccurred())
			node1HttpResponse, err := json.Marshal(nodes[1])
			Expect(err).ToNot(HaveOccurred())

			server.RouteToHandler("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodes[0].ID),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodes[0].ID)),
					ghttp.RespondWith(http.StatusNoContent, []byte{}),
				),
			)

			server.RouteToHandler("GET", fmt.Sprintf("/api/common/nodes/%s", nodes[0].ID),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/common/nodes/%s", nodes[0].ID)),
					ghttp.RespondWith(http.StatusOK, node0HttpResponse),
				),
			)

			server.RouteToHandler("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodes[1].ID),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodes[1].ID)),
					ghttp.RespondWith(http.StatusNoContent, []byte{}),
				),
			)

			server.RouteToHandler("GET", fmt.Sprintf("/api/common/nodes/%s", nodes[1].ID),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/common/nodes/%s", nodes[1].ID)),
					ghttp.RespondWith(http.StatusOK, node1HttpResponse),
				),
			)

			node, err := randomSelectAvailableNode(cpiConfig, nodes, allowFilter)
			Expect(err).ToNot(HaveOccurred())
			Expect(node.ID).To(Equal("55e79ea54e66816f6152fff9"))
		})

		It("return an error if all nodes are created vms with cids", func() {
			nodes := helpers.LoadNodes("../spec_assets/dummy_all_nodes_are_vms.json")

			node0HttpResponse, err := json.Marshal(nodes[0])
			Expect(err).ToNot(HaveOccurred())
			node1HttpResponse, err := json.Marshal(nodes[1])
			Expect(err).ToNot(HaveOccurred())

			server.RouteToHandler("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodes[0].ID),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodes[0].ID)),
					ghttp.RespondWith(http.StatusNoContent, []byte{}),
				),
			)

			server.RouteToHandler("GET", fmt.Sprintf("/api/common/nodes/%s", nodes[0].ID),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/common/nodes/%s", nodes[0].ID)),
					ghttp.RespondWith(http.StatusOK, node0HttpResponse),
				),
			)

			server.RouteToHandler("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodes[1].ID),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodes[1].ID)),
					ghttp.RespondWith(http.StatusNoContent, []byte{}),
				),
			)

			server.RouteToHandler("GET", fmt.Sprintf("/api/common/nodes/%s", nodes[1].ID),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/common/nodes/%s", nodes[1].ID)),
					ghttp.RespondWith(http.StatusOK, node1HttpResponse),
				),
			)

			_, err = randomSelectAvailableNode(cpiConfig, nodes, allowFilter)
			Expect(err).To(MatchError("all nodes have been reserved"))
		})
	})

	Describe("retrying node reservation", func() {
		It("return a node if selection is successful", func() {
			cpiConfig.MaxReserveNodeAttempts = 3
			nodeID, err := TryReservation(
				cpiConfig,
				"",
				func(config.Cpi, string, Filter) (rackhdapi.Node, error) { return rackhdapi.Node{ID: "node-1234"}, nil },
				func(config.Cpi, rackhdapi.Node) error { return nil },
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(nodeID).To(Equal("node-1234"))
		})

		It("returns an error if selection continually fails", func() {
			cpiConfig.MaxReserveNodeAttempts = 3
			nodeID, err := TryReservation(
				cpiConfig,
				"",
				func(config.Cpi, string, Filter) (rackhdapi.Node, error) { return rackhdapi.Node{ID: "node-1234"}, nil },
				func(config.Cpi, rackhdapi.Node) error { return errors.New("error") },
			)
			Expect(err).To(MatchError("unable to reserve node: error"))
			Expect(nodeID).To(Equal(""))
		})

		It("retries and eventually returns a node when selection is successful", func() {
			cpiConfig.MaxReserveNodeAttempts = 3
			tries := 0
			flakeySelectionFunc := func(config.Cpi, string, Filter) (rackhdapi.Node, error) {
				if tries < 2 {
					tries++
					return rackhdapi.Node{}, errors.New("")
				}
				return rackhdapi.Node{ID: "node-1234"}, nil
			}
			nodeID, err := TryReservation(
				cpiConfig,
				"",
				flakeySelectionFunc,
				func(config.Cpi, rackhdapi.Node) error { return nil },
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(nodeID).To(Equal("node-1234"))
		})

		It("cleans up reservation flag after receive timeout error on reserve function", func() {
			apiServer, err := helpers.GetRackHDHost()
			Expect(err).ToNot(HaveOccurred())

			jsonReader := strings.NewReader(fmt.Sprintf(`{"apiserver":"%s", "agent":{"blobstore": {"provider":"local","some": "options"}, "mbus":"localhost", "disks":{"system": "/dev/sda"}}, "max_reserve_node_attempts":1}`, apiServer))
			c, err := config.New(jsonReader, request)
			Expect(err).ToNot(HaveOccurred())

			var testNodeID string
			flakeyReservationFunc := func(c config.Cpi, node rackhdapi.Node) error {
				testNodeID = node.ID
				url := fmt.Sprintf("%s/api/common/nodes/%s", c.ApiServer, node.ID)

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

				nodeURL := fmt.Sprintf("%s/api/common/nodes/%s", c.ApiServer, testNodeID)
				nodeResp, err := http.Get(nodeURL)
				Expect(err).ToNot(HaveOccurred())

				nodeBytes, err := ioutil.ReadAll(nodeResp.Body)
				Expect(err).ToNot(HaveOccurred())
				defer nodeResp.Body.Close()

				var expectedNode rackhdapi.Node
				err = json.Unmarshal(nodeBytes, &expectedNode)
				Expect(err).ToNot(HaveOccurred())
				Expect(expectedNode.Status).To(Equal(rackhdapi.Reserved))

				return errors.New("Timed out running workflow: AWorkflow on node: 12345")
			}

			_, err = SelectNodeFromRackHD(c, "", allowFilter)
			Expect(err).ToNot(HaveOccurred())

			_, err = TryReservation(
				c,
				"",
				SelectNodeFromRackHD,
				flakeyReservationFunc,
			)

			Expect(err).To(MatchError("unable to reserve node: Timed out running workflow: AWorkflow on node: 12345"))
			nodeURL := fmt.Sprintf("%s/api/common/nodes/%s", c.ApiServer, testNodeID)
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
			rawWorkflow := helpers.LoadJSON("../spec_assets/dummy_workflow_response.json")
			httpWorkflowsResponse := []byte(fmt.Sprintf("%s", string(rawWorkflow)))
			var expectedResponse rackhdapi.WorkflowResponse
			err := json.Unmarshal(httpWorkflowsResponse, &expectedResponse)
			Expect(err).ToNot(HaveOccurred())

			httpNodeResponse := helpers.LoadJSON("../spec_assets/dummy_one_node_response.json")

			nodeID := "5665a65a0561790005b77b85"
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodeID)),
					ghttp.RespondWith(http.StatusOK, httpWorkflowsResponse),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/common/nodes/%s", nodeID)),
					ghttp.RespondWith(http.StatusOK, httpNodeResponse),
				),
			)

			nodes := helpers.LoadNodes("../spec_assets/dummy_one_node_running_workflow.json")
			_, err = randomSelectAvailableNode(cpiConfig, nodes, allowFilter)
			Expect(err).To(MatchError("all nodes have been reserved"))
		})
	})

	Describe("when a node does not have obmsettings", func() {
		It("skips the node", func() {
			rawWorkflow := helpers.LoadJSON("../spec_assets/dummy_workflow_response.json")
			httpWorkflowsResponse := []byte(fmt.Sprintf("[%s]", string(rawWorkflow)))
			var expectedResponse []rackhdapi.WorkflowResponse
			err := json.Unmarshal(httpWorkflowsResponse, &expectedResponse)
			Expect(err).ToNot(HaveOccurred())

			httpNodeResponse := helpers.LoadJSON("../spec_assets/dummy_one_node_without_obmsettings_response.json")

			nodeID := "5665a65a0561790005b77b85"
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodeID)),
					ghttp.RespondWith(http.StatusNoContent, []byte{}),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/common/nodes/%s", nodeID)),
					ghttp.RespondWith(http.StatusOK, httpNodeResponse),
				),
			)

			nodes := helpers.LoadNodes("../spec_assets/dummy_one_node_running_workflow.json")
			_, err = randomSelectAvailableNode(cpiConfig, nodes, allowFilter)
			Expect(err).To(MatchError("all nodes have been reserved"))
		})
	})

	Describe("reserving multiple nodes simultaneously", func() {
		XIt("works", func() {
			var wg sync.WaitGroup

			CallTryReservation := func(c config.Cpi, nodes []rackhdapi.Node) {
				defer GinkgoRecover()
				_, err := TryReservation(
					c,
					"",
					SelectNodeFromRackHD,
					ReserveNodeFromRackHD,
				)
				Expect(err).ToNot(HaveOccurred())
				defer wg.Done()
			}

			apiServer, err := helpers.GetRackHDHost()
			Expect(err).ToNot(HaveOccurred())

			c := config.Cpi{ApiServer: apiServer, MaxReserveNodeAttempts: 5, RunWorkflowTimeoutSeconds: 4 * 60}

			nodes, err := rackhdapi.GetNodes(c)
			Expect(err).ToNot(HaveOccurred())

			times := 3
			for i := 0; i < times; i++ {
				wg.Add(1)
				go CallTryReservation(c, nodes)
			}

			wg.Wait()
		})
	})
})

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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/helpers"
	"github.com/rackhd/rackhd-cpi/models"
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

	Describe("parseCreateVMInput", func() {
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
        ["nodeid-uuid"],
        {}]`)
			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)

			Expect(err).ToNot(HaveOccurred())
			agentID, vmCID, publicKey, networks, nodeID, err := parseCreateVMInput(extInput)
			Expect(err).ToNot(HaveOccurred())
			Expect(agentID).To(Equal("4149ba0f-38d9-4485-476f-1581be36f290"))
			Expect(vmCID).To(Equal("vm-478585"))
			Expect(nodeID).To(Equal("nodeid")) //the diskCID == nodeID in this case because we want to know which node we need to pick for the disk
			Expect(publicKey).To(Equal("1234"))
			Expect(networks).To(Equal(map[string]bosh.Network{"private": bosh.Network{
				NetworkType: bosh.DynamicNetworkType,
			}}))
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

	Describe("building the BOSH agent networking spec", func() {
		It("returns an error if no active networks can be found", func() {
			dummyCatalogfile, err := os.Open("../spec_assets/dummy_node_catalog_all_interface_down_response.json")
			Expect(err).ToNot(HaveOccurred())
			defer dummyCatalogfile.Close()

			b, err := ioutil.ReadAll(dummyCatalogfile)
			Expect(err).ToNot(HaveOccurred())

			nodeCatalog := models.NodeCatalog{}

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

			nodeCatalog := models.NodeCatalog{}

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

				nodeCatalog := models.NodeCatalog{}

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
				Expect(netSpec.MAC).To(Equal("52:54:be:ef:fd:e0"))
			})
		})
	})

	Describe("SelectNodeFromRackHD", func() {
		Context("with a nodeID", func() {
			It("selects the node with the nodeID", func() {
				node := helpers.LoadNode("../spec_assets/dummy_create_vm_with_disk_response.json")

				nodeResponse, err := json.Marshal(node)
				Expect(err).ToNot(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/nodes/5665a65a0561790005b77b85")),
						ghttp.RespondWith(http.StatusOK, nodeResponse),
					),
				)

				node, err = SelectNodeFromRackHD(cpiConfig, "5665a65a0561790005b77b85", allowFilter)
				Expect(err).ToNot(HaveOccurred())
				Expect(node.ID).To(Equal("5665a65a0561790005b77b85"))
			})
		})

		Context("When a node has no unavailable or blocked tag and no running workflow", func() {
			It("selects the node", func() {
				//Once upon a time, there were 3 nodes from tag_nodes_all.json
				//The node 583f2dec08a459ab6085a867 was unavailable
				helpers.AddHandler(server, "GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", models.Unavailable), 200, helpers.LoadJSON("../spec_assets/tag_nodes_reserved.json"))
				//Node same 583f2dec08a459ab6085a867 was blocked
				helpers.AddHandler(server, "GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", models.Blocked), 200, helpers.LoadJSON("../spec_assets/tag_nodes_blocked.json"))

				//Among 3 nodes, only node 583f2dec08a459ab6085a867 and 583f2ddd08a459ab6085a85a are computes node
				helpers.AddHandlerWithParam(server, "GET", "/api/2.0/nodes", "type=compute", 200, helpers.LoadJSON("../spec_assets/tag_nodes_all.json"))

				//So it should return [583f2ddd08a459ab6085a85a] as available nodes because 583f2dec08a459ab6085a867 is blocked and unavailable

				//Now it check if 583f2ddd08a459ab6085a85a has active workflow
				helpers.AddHandlerWithParam(server, "GET", fmt.Sprintf("/api/2.0/nodes/583f2ddd08a459ab6085a85a/workflows"), "active=true", 200, []byte("[]"))

				//Met all condition for available node, it returns 583f2ddd08a459ab6085a85a
				node, err := SelectNodeFromRackHD(cpiConfig, "", allowFilter)
				Expect(err).ToNot(HaveOccurred())
				Expect(node.ID).To(Equal("583f2ddd08a459ab6085a85a"))
			})
		})
	})

	Describe("randomSelectNodeWithoutWorkflow", func() {
		Context("when a node has an active workflow", func() {
			It("skips the node", func() {
				rawWorkflow := helpers.LoadJSON("../spec_assets/dummy_has_workflow_response.json")

				nodeID := "5665a65a0561790005b77b85"
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/nodes/%s/workflows", nodeID), "active=true"),
						ghttp.RespondWith(http.StatusOK, rawWorkflow),
					),
				)

				nodes := helpers.LoadNodes("../spec_assets/dummy_one_node_running_workflow.json")
				_, err := randomSelectNodeWithoutWorkflow(cpiConfig, nodes, allowFilter)
				Expect(err).To(MatchError("all nodes have been reserved"))
			})
		})

		Context("two nodes, one has active workflow", func() {
			It("selects the node that has no active workflow", func() {
				nodes := helpers.LoadNodes("../spec_assets/dummy_all_nodes_are_vms.json")
				workflowResponse := helpers.LoadJSON("../spec_assets/dummy_has_workflow_response.json")

				server.RouteToHandler("GET", fmt.Sprintf("/api/2.0/nodes/%s/workflows", nodes[1].ID),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/nodes/%s/workflows", nodes[1].ID)),
						ghttp.RespondWith(http.StatusOK, workflowResponse),
					),
				)

				server.RouteToHandler("GET", fmt.Sprintf("/api/2.0/nodes/%s/workflows", nodes[0].ID),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/nodes/%s/workflows", nodes[0].ID)),
						ghttp.RespondWith(http.StatusOK, []byte("[]")),
					),
				)

				node, err := randomSelectNodeWithoutWorkflow(cpiConfig, nodes, allowFilter)
				Expect(err).ToNot(HaveOccurred())
				Expect(node.ID).To(Equal("55e79ea54e66816f6152fff9"))
			})
		})
	})

	Describe("retrying node reservation", func() {
		It("return a node if selection is successful", func() {
			cpiConfig.MaxReserveNodeAttempts = 3
			nodeID, err := TryReservation(
				cpiConfig,
				"",
				func(config.Cpi, string, Filter) (models.Node, error) { return models.Node{ID: "node-1234"}, nil },
				func(config.Cpi, string) error { return nil },
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(nodeID).To(Equal("node-1234"))
		})

		It("returns an error if selection continually fails", func() {
			cpiConfig.MaxReserveNodeAttempts = 3
			nodeID, err := TryReservation(
				cpiConfig,
				"",
				func(config.Cpi, string, Filter) (models.Node, error) { return models.Node{ID: "node-1234"}, nil },
				func(config.Cpi, string) error { return errors.New("error") },
			)
			Expect(err).To(MatchError("unable to reserve node: error"))
			Expect(nodeID).To(Equal(""))
		})

		It("retries and eventually returns a node when selection is successful", func() {
			cpiConfig.MaxReserveNodeAttempts = 3
			tries := 0
			flakeySelectionFunc := func(config.Cpi, string, Filter) (models.Node, error) {
				if tries < 2 {
					tries++
					return models.Node{}, errors.New("")
				}
				return models.Node{ID: "node-1234"}, nil
			}
			nodeID, err := TryReservation(
				cpiConfig,
				"",
				flakeySelectionFunc,
				func(config.Cpi, string) error { return nil },
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(nodeID).To(Equal("node-1234"))
		})

		It("cleans up reservation flag after receive timeout error on reserve function", func() {
			apiServer, err := helpers.GetRackHDHost()
			Expect(err).ToNot(HaveOccurred())

			jsonReader := strings.NewReader(fmt.Sprintf(`{"api_url":"%s", "agent":{"blobstore": {"provider":"local","some": "options"}, "mbus":"localhost", "disks":{"system": "/dev/sda"}}, "max_reserve_node_attempts":1}`, apiServer))
			c, err := config.New(jsonReader, request)
			Expect(err).ToNot(HaveOccurred())

			var testNodeID string
			flakeyReservationFunc := func(c config.Cpi, nodeID string) error {
				testNodeID = nodeID

				err = rackhdapi.CreateTag(c, nodeID, models.Unavailable)
				Expect(err).ToNot(HaveOccurred())

				tags, err := rackhdapi.GetTags(c, nodeID)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(tags)).To(Equal(1))
				Expect(tags[0]).To(Equal(models.Unavailable))

				return fmt.Errorf("Timed out running workflow: AWorkflow on node: %s", testNodeID)
			}

			_, err = SelectNodeFromRackHD(c, "", allowFilter)
			Expect(err).ToNot(HaveOccurred())

			_, err = TryReservation(
				c,
				"",
				SelectNodeFromRackHD,
				flakeyReservationFunc,
			)

			Expect(err).To(MatchError(fmt.Sprintf("unable to reserve node: Timed out running workflow: AWorkflow on node: %s", testNodeID)))
			tags, err := rackhdapi.GetTags(c, testNodeID)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(tags)).To(Equal(0))
		})
	})

	Context("reserving multiple nodes simultaneously", func() {
		XIt("works", func() {
			var wg sync.WaitGroup

			CallTryReservation := func(c config.Cpi, nodes []models.Node) {
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

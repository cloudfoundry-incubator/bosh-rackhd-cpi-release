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
	"io/ioutil"
	"os"

	"github.com/onrack/onrack-cpi/bosh"
	"github.com/onrack/onrack-cpi/onrackhttp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("The VM Creation Workflow", func() {
	Describe("parsing director input", func() {
		It("return the fields given a valid config", func() {
			jsonInput := []byte(`[
				"4149ba0f-38d9-4485-476f-1581be36f290",
				"vm-478585",
				{"public_key": "1234"},
				{
						"private": {
								"type": "dynamic"
						}
				},
				[],
				{}]`)
			var extInput bosh.ExternalInput
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

			var extInput bosh.ExternalInput
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

			var extInput bosh.ExternalInput
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())

			_, _, _, _, err = parseCreateVMInput(extInput)
			Expect(err).To(MatchError("config error: Only one network supported, provided length: 2"))
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

			var extInput bosh.ExternalInput
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

			var extInput bosh.ExternalInput
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
			var extInput bosh.ExternalInput
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

				var extInput bosh.ExternalInput
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

				var extInput bosh.ExternalInput
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

				var extInput bosh.ExternalInput
				err := json.Unmarshal(jsonInput, &extInput)
				Expect(err).ToNot(HaveOccurred())

				_, _, _, _, err = parseCreateVMInput(extInput)
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

			nodeCatalog := onrackhttp.NodeCatalog{}

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

			nodeCatalog := onrackhttp.NodeCatalog{}

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

				nodeCatalog := onrackhttp.NodeCatalog{}

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

			It("attaches MAC address information from the OnRack API", func() {
				dummyCatalogfile, err := os.Open("../spec_assets/dummy_node_catalog_response.json")
				Expect(err).ToNot(HaveOccurred())
				defer dummyCatalogfile.Close()

				b, err := ioutil.ReadAll(dummyCatalogfile)
				Expect(err).ToNot(HaveOccurred())

				nodeCatalog := onrackhttp.NodeCatalog{}

				err = json.Unmarshal(b, &nodeCatalog)
				Expect(err).ToNot(HaveOccurred())

				prevSpec := bosh.Network{}

				netSpec, err := attachMAC(nodeCatalog.Data.NetworkData.Networks, prevSpec)
				Expect(err).ToNot(HaveOccurred())
				Expect(netSpec.MAC).To(Equal("00:1E:67:C4:E1:A0"))
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

			var extInput bosh.ExternalInput
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
			dummyResponseFile, err := os.Open("../spec_assets/dummy_all_reserved_nodes_response.json")
			Expect(err).ToNot(HaveOccurred())
			defer dummyResponseFile.Close()

			dummyResponseBytes, err := ioutil.ReadAll(dummyResponseFile)
			Expect(err).ToNot(HaveOccurred())

			nodes := []onrackhttp.Node{}
			err = json.Unmarshal(dummyResponseBytes, &nodes)
			Expect(err).ToNot(HaveOccurred())

			_, err = selectNonReservedNode(nodes)
			Expect(err).To(MatchError("all nodes have been reserved"))
		})

		It("selects a free node for provisioning", func() {
			dummyResponseFile, err := os.Open("../spec_assets/dummy_two_node_response.json")
			Expect(err).ToNot(HaveOccurred())
			defer dummyResponseFile.Close()

			dummyResponseBytes, err := ioutil.ReadAll(dummyResponseFile)
			Expect(err).ToNot(HaveOccurred())

			nodes := []onrackhttp.Node{}
			err = json.Unmarshal(dummyResponseBytes, &nodes)
			Expect(err).ToNot(HaveOccurred())

			onRackID, err := selectNonReservedNode(nodes)
			Expect(err).ToNot(HaveOccurred())
			Expect(onRackID).To(Equal("55e79ea54e66816f6152fff9"))

		})
	})
})

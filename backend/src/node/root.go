/*
 * ************************** BEGIN LICENSE BLOCK ******************************
 *
 * Copyright © 2024 Christian Stolze
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * ************************** END LICENSE BLOCK ********************************
 */

package node

/*

The node package handles all the functionality of the Renderhive nodes. The node
types are:

    (1) Render nodes
    (2) Client nodes
    (3) Mediator nodes (not implemented)

*/

import (

	// standard

	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	// "os"
	// "time"

	// external

	"github.com/spf13/cobra"

	// internal
	. "renderhive/globals"
	"renderhive/hedera"
	"renderhive/logger"
)

// Node data of the node running this service app instance
type NodeData struct {
	ID   int    // Renderhive Node ID given by the Renderhive Smart Contract
	Name string // A name for the node, which makes it human identifiable

	// Configuration
	ClientNode    bool // True, if the node acts as a client node
	RenderNode    bool // True, if the node acts as a render node
	HederaAccount struct {
		AccountID string // ID of the node's Hedera account
		PublicKey string // ID of the node's Hedera account
	}
}

// Define the JSON data structure for the node data
type NodeDataJSON struct {
	ID            int    `json:"ID"`
	Name          string `json:"Name"`
	ClientNode    bool   `json:"ClientNode"`
	RenderNode    bool   `json:"RenderNode"`
	HederaAccount struct {
		AccountID string `json:"AccountID"`
		PublicKey string `json:"PublicKey"`
	} `json:"HederaAccount"`
}

// Render data of the node running this service app instance
type RenderData struct {

	// Render requests and offers
	Offer    *RenderOffer           // Render offer provided by this node (if any)
	Requests map[int]*RenderRequest // Render jobs requested by this node (if any)

	// Job queues
	NodeQueue []*RenderJob // Queue of render jobs to be performed on this node

	// Node status
	Busy bool // True, if the node is already rendering

}

// Data required to manage the nodes
type PackageManager struct {

	// Basic data on this node and its user
	User     UserData
	Node     NodeData
	Renderer RenderData

	// Network data
	HiveCycle    HiveCycle
	NetworkQueue []*RenderJob // Queue of render jobs on the render hive

	// Hedera consensus service topics
	// Hive cycle topics
	HiveCycleSynchronizationTopic *hedera.HederaTopic
	HiveCycleApplicationTopic     *hedera.HederaTopic
	HiveCycleValidationTopic      *hedera.HederaTopic

	// Render job topics
	JobQueueTopic *hedera.HederaTopic
	JobTopics     []*hedera.HederaTopic

	// Command line interface
	Command      *cobra.Command
	CommandFlags struct {
		FlagPlaceholder bool
	}
}

// NODE MANAGER
// #############################################################################
// create the node manager variable
var Manager = PackageManager{}

// Initialize everything required for the node management
func (nm *PackageManager) Init() error {
	var err error

	// log information
	logger.Manager.Package["node"].Info().Msg("Initializing the node manager ...")

	// Read the node configuration
	err = nm.LoadConfiguration()
	if err != nil {
		return err
	}

	// Initialize the render offer
	nm.InitRenderOffer()

	// initialized the render requests
	nm.Renderer.Requests = make(map[int]*RenderRequest)

	// Add a Blender version to the node's render offer
	nm.Renderer.Offer.AddBlenderVersion("3.2.1", "/Applications/Blender 3.00.app/Contents/MacOS/blender", &[]string{"CYCLES", "EEVEE"}, &[]string{"CPU"}, 4)

	// // start a benchmark with this version
	// err = nm.Renderer.Offer.Blender["3.2.1"].BenchmarkTool.Run(nm.Renderer.Offer, "3.2.1", "CPU")
	// if err  != nil {
	//     // log error event
	//     logger.Manager.Package["node"].Error().Msg(err.Error())
	// }

	return err

}

// Deinitialize the node manager
func (nm *PackageManager) DeInit() error {
	var err error

	// log event
	logger.Manager.Package["node"].Debug().Msg("Deinitializing the node manager ...")

	return err

}

// Write the details of the node to the configuration file
func (nm *PackageManager) WriteNodeData(id int, name string, client_node bool, render_node bool, accountid string, publicKey string) error {
	var err error
	var node NodeDataJSON

	// log event
	logger.Manager.Package["node"].Debug().Msg("Save node data in the configuration file ...")

	// prepare the data for the JSON file
	node.ID = id
	node.Name = name
	node.ClientNode = client_node
	node.RenderNode = render_node
	node.HederaAccount.AccountID = accountid
	node.HederaAccount.PublicKey = publicKey

	// store the operator data in a file, which can be loaded the next time
	data, err := json.MarshalIndent(node, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal operator: %v", err)
	}
	err = os.WriteFile(filepath.Join(RENDERHIVE_APP_DIRECTORY_CONFIG, "node.json"), data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	// read the user data into the node manager
	err = nm.ReadNodeData(false)
	if err != nil {
		return fmt.Errorf("failed to read node date: %v", err)
	}

	return err

}

// Read the details of the node from the configuration file
func (nm *PackageManager) ReadNodeData(allowSkipping bool) error {
	var err error
	var node NodeDataJSON

	// log event
	logger.Manager.Package["node"].Debug().Msg(" [#] Reading the node data from the configuration file ...")

	// Open the configuration file
	file, err := os.Open(filepath.Join(RENDERHIVE_APP_DIRECTORY_CONFIG, "node.json"))
	if err != nil {

		// do not generate an error if skipping is enabled
		if allowSkipping {
			logger.Manager.Package["node"].Error().Msg(err.Error())
			logger.Manager.Package["node"].Error().Msg("Continue without node configuration.")
			return nil
		}
		return err
	}
	defer file.Close()

	// Read the file content
	fileData, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	// get the file content into a structure
	err = json.Unmarshal(fileData, &node)
	if err != nil {
		return err
	}

	// read the node data into the node manager
	nm.Node.ID = node.ID
	nm.Node.Name = node.Name
	nm.Node.ClientNode = node.ClientNode
	nm.Node.RenderNode = node.RenderNode
	nm.Node.HederaAccount.AccountID = node.HederaAccount.AccountID
	nm.Node.HederaAccount.PublicKey = node.HederaAccount.PublicKey

	// log event
	logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] [*] Node ID: %v", nm.Node.ID))
	logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] [*] Name: %v", nm.Node.Name))
	logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] [*] Client node: %v", nm.Node.ClientNode))
	logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] [*] Render node: %v", nm.Node.RenderNode))
	logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] [*] Public Key: %v", nm.Node.HederaAccount.PublicKey))

	return err

}

// Hash the node from the configuration file
func (nm *PackageManager) HashNodeData() ([]byte, error) {
	var err error

	// Open the node configuration file
	file, err := os.Open(filepath.Join(RENDERHIVE_APP_DIRECTORY_CONFIG, "node.json"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a new SHA-256 hasher
	hasher := sha256.New()

	// Copy the file content to the hasher and calculate the fingerprint
	_, err = io.Copy(hasher, file)
	if err != nil {
		return nil, err
	}

	// Get the hash value
	hash := hasher.Sum(nil)
	return hash, nil
}

// Load the node configuration from the configuration file
func (nm *PackageManager) LoadConfiguration() error {
	var err error

	// log event
	logger.Manager.Package["node"].Debug().Msg(" [#] Loading the node configuration ...")

	// Read the user data
	err = nm.ReadUserData()
	if err != nil {
		return err
	}

	// Read the node data
	err = nm.ReadNodeData(true)
	if err != nil {
		return err
	}

	return err

}

// Register this node at the smart contract
func (nm *PackageManager) RegisterNode() error {
	var err error

	// log event
	logger.Manager.Package["node"].Debug().Msg("Registering this node with the smart contract ...")

	return err

}

// COMMAND LINE INTERFACE – NODE MANAGER
// #############################################################################
// Create the command for the command line interface
func (nm *PackageManager) CreateCommand() *cobra.Command {

	// create the package command
	nm.Command = &cobra.Command{
		Use:   "node",
		Short: "Commands for managing the Renderhive node",
		Long:  "This command and its sub-commands enable the management of this Renderhive node",
		Run: func(cmd *cobra.Command, args []string) {

			return

		},
	}

	// add the subcommands
	nm.Command.AddCommand(nm.CreateCommandInfo())
	nm.Command.AddCommand(nm.CreateCommandBlender())
	nm.Command.AddCommand(nm.CreateCommandRequest())

	return nm.Command

}

// Create the CLI command to print node information
func (nm *PackageManager) CreateCommandInfo() *cobra.Command {

	// flags for the info command
	var this bool
	var user bool
	var offer bool
	var hive_cycle bool
	var hive_queue bool

	// create a 'info' command for the node
	command := &cobra.Command{
		Use:   "info",
		Short: "Print information about this node",
		Long:  "This command provides information about the node including those information retrieved or derived from external network data.",
		Run: func(cmd *cobra.Command, args []string) {

			// print the node data of this node
			if this {
				fmt.Println("")
				fmt.Println("Available information about this node:")
				fmt.Printf(" [#] Node ID: %v\n", nm.Node.ID)
				fmt.Printf(" [#] Operating as client node: %v\n", nm.Node.ClientNode)
				fmt.Printf(" [#] Operating as render node: %v\n", nm.Node.RenderNode)
				fmt.Printf(" [#] Node Account ID (Hedera): %v\n", nm.Node.HederaAccount.AccountID)
				fmt.Println("")
			}

			// print the user data
			if user {
				fmt.Println("")
				fmt.Println("This node is registered on the following user:")
				fmt.Printf(" [#] User ID: %v\n", nm.User.ID)
				fmt.Printf(" [#] Username: %v\n", nm.User.Username)
				fmt.Printf(" [#] User Account ID (Hedera): %v\n", nm.User.UserAccount.AccountID.String())
				fmt.Println("")
			}

			// print the render offer
			if offer {
				// if the node has a render offer, print it
				if nm.Renderer.Offer != nil {

					fmt.Println("")
					fmt.Println("This node offers the following render services:")
					fmt.Printf(" [#] Render offer document (CID): %v\n", nm.Renderer.Offer.DocumentCID)
					fmt.Printf(" [#] Supported Blender versions:\n")
					for _, blender := range nm.Renderer.Offer.Blender {
						fmt.Printf("     - Blender v%v (Engines: %v | Devices: %v) \n", blender.BuildVersion, strings.Join(blender.Engines, ", "), strings.Join(blender.Devices, ", "))
					}
					fmt.Println("")

				} else {

					fmt.Println("")
					fmt.Println("This node is not offering a render service.")
					fmt.Println("")

				}
			}

			// print the hive cycle
			if hive_cycle {
				fmt.Println("")
				fmt.Printf("The current hive cycle of the render hive is %v.\n", nm.HiveCycle.Current)
				fmt.Printf(" [#] Started at consensus time: %v\n", nm.HiveCycle.Clock.NetworkStartTime)
				fmt.Printf(" [#] Started at local time: %v\n", nm.HiveCycle.Clock.LocalStartTime)
				fmt.Println("")
			}

			// print the render job queue of the hive
			if hive_queue {

				if len(nm.NetworkQueue) > 0 {
					fmt.Println("")
					fmt.Printf("There are %v render requests in the render hive queue:\n", len(nm.NetworkQueue))

					// go through the list and print each queue
					for i, job := range nm.NetworkQueue {
						fmt.Printf(" [#] [%v] Render job #%v (User: %v; Node: %v): %v\n", job.Request.SubmittedTimestamp, i, job.UserID, job.NodeID, job.Request.DocumentCID)
					}

					fmt.Println("")

				} else {

					fmt.Println("")
					fmt.Println("There are no render requests in the render hive queue.")
					fmt.Println("")

				}
			}

			return

		},
	}

	// add command flags
	command.Flags().BoolVarP(&this, "this", "t", false, "Print the available information about this node")
	command.Flags().BoolVarP(&user, "user", "u", false, "Print the node owner's user data")
	command.Flags().BoolVarP(&offer, "offer", "o", false, "Print the render offer of this node")
	command.Flags().BoolVarP(&hive_cycle, "hive-cycle", "c", false, "Print the current hive cycle this node calculated")
	command.Flags().BoolVarP(&hive_queue, "hive-queue", "q", false, "Print the render job queue of the render hive")

	return command

}

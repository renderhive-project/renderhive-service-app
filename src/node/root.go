/*
 * ************************** BEGIN LICENSE BLOCK ******************************
 *
 * Copyright © 2023 Christian Stolze
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
  "fmt"
  "strings"
  // "os"
  // "time"

  // external
  // hederasdk "github.com/hashgraph/hedera-sdk-go/v2"
  "github.com/spf13/cobra"

  // internal
  // "renderhive/globals"
  "renderhive/logger"
  "renderhive/hedera"

)

// User data of the node's owner
// TODO: add further user data
type UserData struct {

  ID int                              // Renderhive User ID given by the Renderhive Smart Contract
  Username string                     // a user name
  UserAccount hedera.HederaAccount     // Hedera account ID of the user's main account
  NodeAccounts []hedera.HederaAccount  // Hedera account IDs of the user's node accounts

}

// Node data of the node running this service app instance
type NodeData struct {

  ID int                      // Renderhive Node ID given by the Renderhive Smart Contract
  ClientNode bool             // True, if the node acts as a client node
  RenderNode bool             // True, if the node acts as a render node

  UserData *UserData
  NodeAccount *hedera.HederaAccount

}

// Render data of the node running this service app instance
type RenderData struct {

  // Render requests and offers
  Offer *RenderOffer                      // the render offer provided by this node (if any)
  Requests map[int]RenderRequest          // the render jobs requested by this node (if any)

  // Job queues
  NodeQueue *[]RenderJob                  // the queue of render jobs to be performed on this node
  NetworkQueue *[]RenderJob               // the queue of render jobs on the renderhive network

  // Benchmark
  RenderPower []BlenderBenchmarkResult   // Blender benchmark points per Blender version

}

// Data required to manage the nodes
type PackageManager struct {

  // Basic data on this node and its user
  User UserData
  Node NodeData
  Renderer RenderData

  // Hivc cycle management
  HiveCycle HiveCycle

  // Command line interface
  Command *cobra.Command
  CommandFlags struct {

    FlagPlaceholder bool

  }

}


// NODE MANAGER
// #############################################################################
// create the node manager variable
var Manager = PackageManager{}

// Initialize everything required for the node management
func (nm *PackageManager) Init() (error) {
    var err error

    // log information
    logger.Manager.Package["node"].Info().Msg("Initializing the node manager ...")

    // Initialize the render offer
    nm.InitRenderOffer()
    
    // initialized the render requests
    nm.Renderer.Requests = map[int]RenderRequest{}

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
func (nm *PackageManager) DeInit() (error) {
    var err error

    // log event
    logger.Manager.Package["node"].Debug().Msg("Deinitializing the node manager ...")

    return err

}

// COMMAND LINE INTERFACE – NODE MANAGER
// #############################################################################
// Create the command for the command line interface
func (nm *PackageManager) CreateCommand() (*cobra.Command) {

    // create the package command
    nm.Command = &cobra.Command{
    	Use:   "node",
    	Short: "Commands for managing the Renderhive node",
    	Long: "This command and its sub-commands enable the management of this Renderhive node",
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
func (nm *PackageManager) CreateCommandInfo() (*cobra.Command) {

    // flags for the info command
    var hivecycle bool
    var offer bool
    var this bool
    var user bool

    // create a 'info' command for the node
    command := &cobra.Command{
    	Use:   "info",
    	Short: "Print information about this node",
    	Long: "This command provides information about the node including those information retrieved or derived from external network data.",
      Run: func(cmd *cobra.Command, args []string) {

        // print the hive cycle
        if hivecycle {
            fmt.Println("")
            fmt.Printf("The current hive cycle of the renderhive is %v.\n", nm.HiveCycle.Current)
            fmt.Println("")
        }

        // print the node data of this node
        if this {
            fmt.Println("")
            fmt.Println("Available information about this node:")
            fmt.Printf(" [#] Node ID: %v\n", nm.Node.ID)
            fmt.Printf(" [#] Operating as client node: %v\n", nm.Node.ClientNode)
            fmt.Printf(" [#] Operating as render node: %v\n", nm.Node.RenderNode)
            if nm.Node.NodeAccount != nil {
                fmt.Printf(" [#] Node Account ID (Hedera): %v\n", nm.Node.NodeAccount.AccountID.String())
            }
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

        // print the user data
        if user {
            fmt.Println("")
            fmt.Println("This node is registered on the following user:")
            fmt.Printf(" [#] User ID: %v\n", nm.User.ID)
            fmt.Printf(" [#] Username: %v\n", nm.User.Username)
            fmt.Printf(" [#] User Account ID (Hedera): %v\n", nm.User.UserAccount.AccountID.String())
            fmt.Println("")
        }

        return

    	},
    }

    // add command flags
    command.Flags().BoolVarP(&hivecycle, "hivecycle", "c", false, "Print the current hive cycle this node calculated")
    command.Flags().BoolVarP(&offer, "offer", "o", false, "Print the render offer of this node")
    command.Flags().BoolVarP(&this, "this", "t", false, "Print the available information about this node")
    command.Flags().BoolVarP(&user, "user", "u", false, "Print the node owner's user data")

    return command

}

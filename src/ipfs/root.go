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

package ipfs

/*

The ipfs package handles the local IPFS node that is run as part of the Renderhive
Service app. IPFS is used for exchange of Blender files, render results, and
other types of data required to submit and process render jobs.

*/

/*

GO-IPFS EXAMPLES:

- Spawn a local node
  https://github.com/ipfs/kubo/tree/c9cc09f6f7ebe95da69be6fa92c88e4cb245d90b/docs/examples/go-ipfs-as-a-library

*/

import (

  // standard
  // "fmt"
  // "os"
  // "time"

  // external
  // hederasdk "github.com/hashgraph/hedera-sdk-go/v2"
  "github.com/spf13/cobra"

  // internal
  "renderhive/logger"
  // . "renderhive/globals"
  // "renderhive/hedera"
)

// structure for the IPFS manager
type PackageManager struct {

  // Placeholder
  Placeholder string

  // Command line interface
  Command *cobra.Command
  CommandFlags struct {

    FlagPlaceholder bool

  }

}


// IPFS MANAGER
// #############################################################################
// create the ipfs manager variable
var Manager = PackageManager{}

// Initialize everything required for the IPFS management
func (ipfsm *PackageManager) Init() (error) {
    var err error

    // log information
    logger.Manager.Package["ipfs"].Info().Msg("Initializing the IPFS manager ...")

    return err

}

// Deinitialize the ipfs manager
func (ipfsm *PackageManager) DeInit() (error) {
    var err error

    // log event
    logger.Manager.Package["ipfs"].Debug().Msg("Deinitializing the IPFS manager ...")

    return err

}

// IPFS MANAGER COMMAND LINE INTERFACE
// #############################################################################
// Create the command for the command line interface
func (ipfsm *PackageManager) CreateCommand() (*cobra.Command) {

    // create the package command
    ipfsm.Command = &cobra.Command{
    	Use:   "ipfs",
    	Short: "Commands for the interaction with the IPFS and Filecoin services",
    	Long: "This command and its sub-commands enable the interaction with the IPFS and Filecoin services required by the Renderhive network",
      Run: func(cmd *cobra.Command, args []string) {

        return

    	},
    }

    return ipfsm.Command

}

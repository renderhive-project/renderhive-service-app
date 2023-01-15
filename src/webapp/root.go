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

package webapp

/*

The webapp package provides the backend and front end for the user UI, which will
be served locally as a webapp.

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
  // "renderhive/globals"
  // "renderhive/hedera"
)

// structure for the web app manager
type PackageManager struct {

  // Placeholder
  Placeholder string

  // Command line interface
  Command *cobra.Command
  CommandFlags struct {

    FlagPlaceholder bool

  }

}


// WEBAPP MANAGER
// #############################################################################
// create the render manager variable
var Manager = PackageManager{}

// Initialize everything required for the web app management
func (webappm *PackageManager) Init() (error) {
    var err error

    // log information
    logger.Manager.Package["webapp"].Info().Msg("Initializing the web app manager ...")

    return err

}

// Deinitialize the web app manager
func (webappm *PackageManager) DeInit() (error) {
    var err error

    // log event
    logger.Manager.Package["webapp"].Debug().Msg("Deinitializing the web app manager ...")

    return err

}

// RENDER MANAGER COMMAND LINE INTERFACE
// #############################################################################
// Create the command for the command line interface
func (webappm *PackageManager) CreateCommand() (*cobra.Command) {

    // create the package command
    webappm.Command = &cobra.Command{
    	Use:   "webapp",
    	Short: "Commands for the web frontend of the Renderhive Service App",
    	Long: "This command and its sub-commands enable the management of the web frontend for the Renderhive Service App.",
      Run: func(cmd *cobra.Command, args []string) {

        return

    	},
    }

    return webappm.Command

}

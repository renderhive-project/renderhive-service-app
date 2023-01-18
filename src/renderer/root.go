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

package renderer

/*

The renderer package handles all the functionality for rendering. It manages
render jobs, the render result, calls Blender, etc.

*/

import (

  // standard
  "fmt"
  "strings"
  "errors"
  // "os"
  // "time"

  // external
  // hederasdk "github.com/hashgraph/hedera-sdk-go/v2"
  "github.com/spf13/cobra"

  // internal
  // . "renderhive/globals"
  "renderhive/logger"
  // "renderhive/hedera"
)



// RENDER JOBS, OFFERS, AND REQUESTS
// #############################################################################
// App data of supported Blender version
type BlenderAppData struct {

  Version string              // Version of this Blender app
  Path string                 // Path to the Blender app

  // Render settings supported by this node's Blender instance
  Engines *[]string                    // Supported render engines
  Devices *[]string                    // Supported devices

}

// Blender file data
type BlenderFileData struct {

  // TODO: Fill with information
  CID string                  // content identifier (CID) of the .blend file on the IPFS

  // Render settings
  Settings RenderSettings     // rendering settings of this render job

}

// Blender render settings
type RenderSettings struct {

  // TODO: Fill with further required information
  // Render settings
  Engine string               // Render engine to be used (i.e., Cycles, EEVEE)
  FeatureSet string           // Blender feature set to be used
  Device string               // CPU, GPU or both?
  ResolutionX int             // x resolution of the render result
  ResolutionY int             // y resolution of the render result
  TileX int                   // x resolution of tiles to be rendered
  TileY int                   // y resolution of tiles to be rendered

  OutputFilepath string       // Output path (includes file naming)

}

// a render job claimed for rendering on the renderhive by this node
type RenderJob struct {

  // TODO: Fill with information
  UserID int                   // ID of the user this render job belongs to

  // File data
  BlenderFile BlenderFileData  // Data of the Blender file to be rendered
  Document string              // content identifier (CID) of the render request document on the IPFS

}

// a render job that is requested by this node for rendering on the renderhive
type RenderRequest struct {

  // TODO: Fill with information
  UserID int                   // ID of the user this request belongs to

  // File data
  BlenderFile BlenderFileData  // data of the Blender file to be rendered
  Document string              // content identifier (CID) of the render request document on the IPFS

}

// a render offer that is provided by this node for rendering on the renderhive
type RenderOffer struct {

  // TODO: Fill with information
  UserID int                  // ID of the user this offer belongs to
  Document string             // content identifier (CID) of the render offer document on the IPFS

  // Render offer
  Blender map[string]BlenderAppData    // supported Blender version and render settings

}



// RENDER MANAGER
// #############################################################################

// structure for the render manager
type PackageManager struct {

  // Command line interface
  Command *cobra.Command
  CommandFlags struct {

    FlagPlaceholder bool

  }

}

// create the render manager variable
var Manager = PackageManager{}

// Initialize everything required for the render management
func (rm *PackageManager) Init() (error) {
    var err error

    // log information
    logger.Manager.Package["renderer"].Info().Msg("Initializing the render manager ...")

    return err

}

// Deinitialize the render manager
func (rm *PackageManager) DeInit() (error) {
    var err error

    // log event
    logger.Manager.Package["renderer"].Debug().Msg("Deinitializing the render manager ...")

    return err

}


// RENDER OFFERS
// #############################################################################
// Add a Blender version to the render offer
func (ro *RenderOffer) AddBlenderVersion(version string, path string, engines *[]string, devices *[]string) (error) {
    var err error

    // log event
    logger.Manager.Package["renderer"].Trace().Msg("Add a Blender version supported by this node:")
    logger.Manager.Package["renderer"].Trace().Msg(fmt.Sprintf(" [#] Version: %v", version))
    logger.Manager.Package["renderer"].Trace().Msg(fmt.Sprintf(" [#] Path: %v", path))
    logger.Manager.Package["renderer"].Trace().Msg(fmt.Sprintf(" [#] Engines: %v", strings.Join(*engines, ",")))
    logger.Manager.Package["renderer"].Trace().Msg(fmt.Sprintf(" [#] Engines: %v", strings.Join(*devices, ",")))

    // append it to the slice of supported Blender versions in the render offer
    ro.Blender[version] = BlenderAppData{
                            Version: version,
                            Path: path,
                            Engines: engines,
                            Devices: devices,
                          }

    // dcheck if the new element exists in the map and return in error if not
    _, ok := ro.Blender[version]
    if ok == false {
        err = errors.New(fmt.Sprintf("Blender version '%v' could not be added.", version))
    }

    return err

}

// Delete a Blender version from the render offer
func (ro *RenderOffer) DeleteBlenderVersion(version string) (error) {
    var err error

    // log event
    logger.Manager.Package["renderer"].Trace().Msg("Delete a Blender version from the node's render offer:")
    logger.Manager.Package["renderer"].Trace().Msg(fmt.Sprintf(" [#] Version: %v", version))

    // delete the element from the map, if it exists
    _, ok := ro.Blender[version]
    if ok {
        delete(ro.Blender, version)
    } else {
        err = errors.New(fmt.Sprintf("Blender version '%v' could not be deleted.", version))
    }

    return err

}

// RENDER MANAGER COMMAND LINE INTERFACE
// #############################################################################
// Create the command for the command line interface
func (rm *PackageManager) CreateCommand() (*cobra.Command) {

    // create the package command
    rm.Command = &cobra.Command{
    	Use:   "renderer",
    	Short: "Commands for rendering and render job management",
    	Long: "This command and its sub-commands enable the management of the render jobs for this Renderhive node",
      Run: func(cmd *cobra.Command, args []string) {

        return

    	},
    }

    return rm.Command

}

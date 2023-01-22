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

import (

  // standard
  "fmt"
  "strings"
  "errors"
  "os"
  "os/exec"
  "bufio"
  "io"
  // "time"

  // external
  // hederasdk "github.com/hashgraph/hedera-sdk-go/v2"
  "github.com/mattn/go-shellwords"
  "github.com/spf13/cobra"

  // internal
  // . "renderhive/globals"
  "renderhive/logger"

)



// RENDER JOBS, OFFERS, AND REQUESTS
// #############################################################################
// App data of supported Blender version
type BlenderAppData struct {

  // App info
  Version string              // Version of this Blender app
  Path string                 // Path to the Blender app
  StdOut io.ReadCloser      // Command-line standard output of the Blender app
  StdErr io.ReadCloser      // Command-line error output of the Blender app

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

// Blender benchmark
type BlenderBenchmark struct {

  Version float64             // Blender benchmark tool version
  Points float64              // Blender benchmark points

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
  UserID int                           // ID of the user this offer belongs to
  Document string                      // content identifier (CID) of the render offer document on the IPFS

  // Render offer
  RenderPower float64                  // render power offered by the node
  Price float64                        // price of rendering
  Blender map[string]BlenderAppData    // supported Blender version and render settings

}

// RENDER OFFERS
// #############################################################################
// Add a Blender version to the render offer
func (ro *RenderOffer) AddBlenderVersion(version string, path string, engines *[]string, devices *[]string) (error) {
    var err error

    // log event
    logger.Manager.Package["node"].Trace().Msg("Add a Blender version supported by this node:")
    logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Version: %v", version))
    logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Path: %v", path))
    logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Engines: %v", strings.Join(*engines, ",")))
    logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Engines: %v", strings.Join(*devices, ",")))

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
    logger.Manager.Package["node"].Trace().Msg("Delete a Blender version from the node's render offer:")
    logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Version: %v", version))

    // delete the element from the map, if it exists
    _, ok := ro.Blender[version]
    if ok {
        delete(ro.Blender, version)
    } else {
        err = errors.New(fmt.Sprintf("Blender version '%v' could not be deleted.", version))
    }

    return err

}


// BLENDER CONTROL
// #############################################################################
// Start Blender with command line flags and render the given blend_file
func (b *BlenderAppData) Execute(args []string) (error) {
    var err error

    // log event
    logger.Manager.Package["node"].Debug().Msg("Starting Blender:")
    logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] Version: %v", b.Version))
    logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] Path: %v", b.Path))

    // Check if path is pointing to an existing file
    if _, err = os.Stat(b.Path); os.IsNotExist(err) {
        return err
    }

    // Execute Blender in background mode
    cmd := exec.Command(b.Path, append([]string{"-b"}, args...)...)
    b.StdOut, _ = cmd.StdoutPipe()
    b.StdErr, _ = cmd.StderrPipe()
    err = cmd.Start()
    if err != nil {
        fmt.Println(err)
        return err
    }

    // Print the process ID of the running Blender instance
    logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] PID: %v", cmd.Process.Pid))

    // check for both Blender output in go routine
    go b.ProcessOutput("StdOut", b.StdOut)
    go b.ProcessOutput("StdErr", b.StdErr)

    return err

}


// Start Blender with command line flags and render the given blend_file
func (b *BlenderAppData) ProcessOutput(name string, output io.ReadCloser) (error) {
    var err error

    // empty line
    fmt.Println("")

    // Create a new scanner to read the command output
    scanner := bufio.NewScanner(output)
    for scanner.Scan() {

      // Print the command line output of Blender
      fmt.Printf("(blender) > %v: %v \n", name, scanner.Text())

    }

    return err

}

// COMMAND LINE INTERFACE - BLENDER AND RENDERING
// #############################################################################
// Create the CLI command to control Blender from the Render Service App
func (nm *PackageManager) CreateCommandBlender() (*cobra.Command) {

    // flags for the 'blender' command
    var list bool

    // create a 'blender' command for the node
    command := &cobra.Command{
    	Use:   "blender",
    	Short: "Manage the node's Blender versions",
    	Long: "This command is for adding/removing Blender versions from the node's render offer. You can also start Blender in the background mode for rendering purposes.",
      Run: func(cmd *cobra.Command, args []string) {

        // if a render offer exists
        if nm.Renderer.Offer != nil {

            // list all Blender versions
            if list {

                fmt.Println("")
                fmt.Println("The node offers the following Blender versions for rendering:")

                // find each Blender version added to the node
                for version, _ := range nm.Renderer.Offer.Blender {

                  fmt.Printf(" [#] Version: %v\n", version)

                }
                fmt.Println("")

            }

        } else {

          fmt.Println("")
          fmt.Println(fmt.Errorf("The node has no render offer."))
          fmt.Println("")
        }

        return

    	},
    }

    // add command flags
    command.Flags().BoolVarP(&list, "list", "l", false, "List all the Blender versions from the node's render offer")

    // add the subcommands
    command.AddCommand(nm.CreateCommandBlender_Add())
    command.AddCommand(nm.CreateCommandBlender_Remove())
    command.AddCommand(nm.CreateCommandBlender_Run())

    return command

}

// Create the CLI command to add a Blender version to the node's render offer
func (nm *PackageManager) CreateCommandBlender_Add() (*cobra.Command) {

    // flags for the 'blender' command
    var version string
    var path string
    var engines []string
    var devices []string

    // create a 'blender add' command for the node
    command := &cobra.Command{
    	Use:   "add",
    	Short: "Add a Blender version to the node's render offer",
    	Long: "This command is for adding a Blender version to the node's render offer.",
      Run: func(cmd *cobra.Command, args []string) {

        // if a render offer exists
        if nm.Renderer.Offer != nil {

            // add a Blender version
            if len(version) != 0 {
                fmt.Println("")

                // Check if path is pointing to an existing file
                if _, err := os.Stat(path); os.IsNotExist(err) {
                    fmt.Println(fmt.Errorf("The given path '%v' is not a valid path.", path))
                    return
                }

                fmt.Printf("Adding the version '%v' with path '%v' to the render offer. \n", version, path)
                fmt.Println("")

                // Add a new Blender version to the node's render offer
                nm.Renderer.Offer.AddBlenderVersion(version, path, &engines, &devices)

            }

        } else {

          fmt.Println("")
          fmt.Println(fmt.Errorf("The node has no render offer."))
          fmt.Println("")

        }

        return

    	},
    }

    // add command flag parameters
    command.Flags().StringVarP(&version, "version", "v", "", "The version of Blender to be used")
    command.Flags().StringVarP(&path, "path", "p", "", "The path to a Blender executable on this computer")
    command.Flags().StringArrayVarP(&engines, "engines", "E", []string{"EEVEE", "CYCLES"}, "The supported engines (only EEVEE and CYCLES)")
    command.Flags().StringArrayVarP(&devices, "devices", "D", []string{"CPU"}, "The supported devices (only CPU and GPU)")

    return command

}

// Create the CLI command to remove a Blender version from the node's render offer
func (nm *PackageManager) CreateCommandBlender_Remove() (*cobra.Command) {

    // flags for the 'blender remove' command
    var version string

    // create a 'blender remove' command for the node
    command := &cobra.Command{
    	Use:   "remove",
    	Short: "Remove a Blender version from the node's render offer",
    	Long: "This command is for removing a Blender version from the node's render offer.",
      Run: func(cmd *cobra.Command, args []string) {

        // if a render offer exists
        if nm.Renderer.Offer != nil {

            // remove the Blender version
            if len(version) != 0 {

                // if the parsed version is supported by the node
                _, ok := nm.Renderer.Offer.Blender[version]
                if ok {
                    fmt.Println("")
                    fmt.Printf("Removing Blender v%v from the render offer of this node. \n", version)
                    fmt.Println("")

                    // Delete the Blender version from the node's render offer
                    nm.Renderer.Offer.DeleteBlenderVersion(version)

                } else {

                  fmt.Println("")
                  fmt.Println(fmt.Errorf("The node does not support Blender v%v.", version))
                  fmt.Println("")

                }

            }

        } else {

          fmt.Println("")
          fmt.Println(fmt.Errorf("The node has no render offer."))
          fmt.Println("")

        }

        return

    	},
    }

    // add command flag parameters
    command.Flags().StringVarP(&version, "version", "v", "", "The version of Blender to be used")

    return command

}

// Create the CLI command to run a Blender version from the node's render offer
func (nm *PackageManager) CreateCommandBlender_Run() (*cobra.Command) {

    // flags for the 'blender remove' command
    var version string
    var param string

    // create a 'blender remove' command for the node
    command := &cobra.Command{
    	Use:   "run",
    	Short: "Run a Blender version from the node's render offer",
    	Long: "This command is for starting a particular Blender version, which is in the node's render offer.",
      Run: func(cmd *cobra.Command, args []string) {

        // if a render offer exists
        if nm.Renderer.Offer != nil {

            // if a version was parsed and
            if len(version) != 0 {

                // if the parsed version is supported by the node
                blender, ok := nm.Renderer.Offer.Blender[version]
                if ok {
                    fmt.Println("")
                    fmt.Printf("Starting Blender v%v. \n", version)
                    fmt.Println("")

                    // parse the command line parameters for Blender
                    args, err := shellwords.Parse(param)
                    // fmt.Println(args)
                    if err != nil {
                        fmt.Println("")
                        fmt.Println(fmt.Errorf("Could not parse the command line arguments for Blender."))
                        fmt.Println("")
                    }
                    // run the this Blender version
                    blender.Execute(args)

                } else {

                  fmt.Println("")
                  fmt.Println(fmt.Errorf("The node does not support Blender v%v.", version))
                  fmt.Println("")

                }
            }

        } else {

          fmt.Println("")
          fmt.Println(fmt.Errorf("The node has no render offer."))
          fmt.Println("")

        }

        return

    	},
    }

    // add command flag parameters
    command.Flags().StringVarP(&version, "version", "v", "", "The version of Blender to be used")
    command.Flags().StringVarP(&param, "param", "p", "", "The command line options for Blender")

    return command

}

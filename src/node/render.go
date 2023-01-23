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
  "regexp"
  // "time"

  // external
  // hederasdk "github.com/hashgraph/hedera-sdk-go/v2"
  "github.com/mattn/go-shellwords"
  "github.com/spf13/cobra"
  // "golang.org/x/exp/slices" <-- would be handy, but requires Go 1.18; TODO: Update possible for Hedera SDK?

  // internal
  . "renderhive/globals"
  "renderhive/logger"

)



// RENDER JOBS, OFFERS, AND REQUESTS
// #############################################################################
// App data of supported Blender version
type BlenderAppData struct {

  // App info
  Path string               // Path to the Blender app

  // Build info
  BuildVersion string       // Build version of this Blender app
  BuildHash string          // Build hast of this Blender app
  BuildDate string          // Build date of this Blender app
  BuildTime string          // Build time of this Blender app

  // Render settings supported by this node's Blender instance
  Engines []string                    // Supported render engines
  Devices []string                    // Supported devices

  // Process status
  Cmd *exec.Cmd             // pointer to the exec.Command type
  Param []string            // Command line options this Blender process was called with
  PID int                   // PID of the process
  Running bool              // Is the process still running
  StdOut io.ReadCloser      // Command-line standard output of the Blender app
  StdErr io.ReadCloser      // Command-line error output of the Blender app

  // Blender render status
  Frame string              // Current frame number
  Memory string             // Current memory usage
  Peak string               // Peak memory usage
  Time string               // Render time
  Note string               // Render status note

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
    logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Internal name: %v", version))
    logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Path: %v", path))
    logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Engines: %v", strings.Join(*engines, ",")))
    logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Devices: %v", strings.Join(*devices, ",")))

    // check if the version WAS already added before
    if _, ok := ro.Blender[version]; ok {
      return errors.New(fmt.Sprintf("Blender version '%v' already exists. Use 'blender edit' to change parameters.", version))
    }

    // check if given engines and devices are valid
    _, err = GetBlenderEngineEnum(*engines)
    if err != nil {
        return errors.New(fmt.Sprintf("At least one of the defined engines '%v' is not valid.", *engines))
    }
    _, err = GetBlenderDeviceEnum(*devices)
    if err != nil {
        return errors.New(fmt.Sprintf("At least one of the defined devices '%v' is not valid.", *devices))
    }

    // create the BlenderAppData instance for the new version
    blender := BlenderAppData{
                  Path: path,
                  Engines: *engines,
                  Devices: *devices,
               }

    // start this Blender version and query its version and build info
    err = blender.Execute([]string{"-v"})
    if err != nil {
        return err
    }

    // wait until the command execution finished
    err = blender.Cmd.Wait()
    if err != nil {
        return err
    }

    // add the new BlenderAppData to the map
    ro.Blender[version] = blender

    // log debug event
    logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Build Version: %v", blender.BuildVersion))
    logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Build Date: %v", blender.BuildDate))
    logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Build Time: %v", blender.BuildTime))
    logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Build Hash: %v", blender.BuildHash))

    return err

}

// Delete a Blender version from the render offer
func (ro *RenderOffer) DeleteBlenderVersion(version string) (error) {
    var err error

    // log event
    logger.Manager.Package["node"].Trace().Msg("Remove a Blender version from the node's render offer:")

    // delete the element from the map, if it exists
    blender, ok := ro.Blender[version]
    if ok {
        logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Version: %v", blender.BuildVersion))
        delete(ro.Blender, version)
    } else {
        err = errors.New(fmt.Sprintf("Blender v'%v' could not be removed from the node's render offer.", blender.BuildVersion))
    }

    return err

}


// BLENDER CONTROL
// #############################################################################
// Start Blender with command line flags and render the given blend_file
func (b *BlenderAppData) Execute(args []string) (error) {
    var err error

    // log event
    logger.Manager.Package["node"].Trace().Msg("Starting Blender:")
    logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Path: %v", b.Path))

    // Check if path is pointing to an existing file
    if _, err = os.Stat(b.Path); os.IsNotExist(err) {
        return err
    }

    // TODO: Validate args and DISALLOW python (for security reasons)
    // if ...

    // Execute Blender in background mode
    b.Cmd = exec.Command(b.Path, append([]string{"-b"}, args...)...)
    b.Param = args
    b.StdOut, _ = b.Cmd.StdoutPipe()
    b.StdErr, _ = b.Cmd.StderrPipe()
    err = b.Cmd.Start()
    if err != nil {
        fmt.Println(err)
        return err
    }

    // store status information
    b.PID = b.Cmd.Process.Pid
    b.Running = true

    // Print the process ID of the running Blender instance
    logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] PID: %v", b.Cmd.Process.Pid))

    // check for both Blender output in go routine
    go b.ProcessOutput("StdOut", b.StdOut)
    go b.ProcessOutput("StdErr", b.StdErr)

    return err

}


// Start Blender with command line flags and render the given blend_file
func (b *BlenderAppData) ProcessOutput(name string, output io.ReadCloser) (error) {
    var err error
    var line string

    // empty line
    // fmt.Println("")

    // Create a new scanner to read the command output
    scanner := bufio.NewScanner(output)
    for scanner.Scan() {

      // read the line
      line = scanner.Text()

      // check for specific parameters
      var version bool
      for _, p := range b.Param {
        if p == "-v" || p == "--version" {
          version = true
          break
        }
      }

      // BLENDER VERSION & BUILD INFO
      // ***********************************************************************
      // extract the Blender version and build info, if the "-v" option was parsed
      str := strings.Split(line, ":")
      if version {

          // Compile regular expression to match the values
          versiontest := regexp.MustCompile("Blender (?:\\d+.\\d+.\\d+.)")
          builddatetest := regexp.MustCompile("build date:")
          buildtimetest := regexp.MustCompile("build time: (?:\\d{2}:\\d{2}:\\d{2})")
          buildhashtest := regexp.MustCompile("build hash:")

          // if the expression matches
          if versiontest.MatchString(line) {

              // extract the values
              re := regexp.MustCompile("(?:\\d+.\\d+.\\d+.)")
              matches := re.FindAllString(line, -1)
              b.BuildVersion = strings.TrimSpace(matches[0])

          }

          // if the expression matches
          if builddatetest.MatchString(line) {

              // extract the values
              b.BuildDate = strings.TrimSpace(str[1])

          }

          // if the expression matches
          if buildtimetest.MatchString(line) {

              // extract the values
              re := regexp.MustCompile("(?:\\d{2}:\\d{2}:\\d{2})")
              matches := re.FindAllString(line, -1)
              b.BuildTime = strings.TrimSpace(matches[0])

          }

          // if the expression matches
          if buildhashtest.MatchString(line) {

              // extract the hash
              b.BuildHash = strings.TrimSpace(str[1])

          }

      }

      // BLENDER QUIT
      // ***********************************************************************
      if strings.EqualFold(line, "Blender quit") {

          // log event message
          logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf("Blender v%v process (pid: %v) finished with 'Blender quit'.", b.BuildVersion, b.PID))

          // store internally that the process is finished
          b.Running = false

          // stop the loop
          break
      }

      // RENDER STATUS
      // ***********************************************************************
      // separate status line into substrings
      str = strings.Split(line, "|")
      if len(str) == 3 {

          // Compile regular expression to match the values
          statustest := regexp.MustCompile("Fra:[0-9]+ Mem:[0-9]*\\.[0-9]+M \\(Peak [0-9]*\\.[0-9]+M\\)")
          timetest := regexp.MustCompile("(?i)Time:\\d\\d:\\d\\d\\.\\d\\d")

          for _, substr := range str {
            // if the expression matches
            if statustest.MatchString(substr) {

                // extract the values
                re := regexp.MustCompile("(?:([0-9]*\\.[0-9]+M)|[0-9]+)")
                matches := re.FindAllString(substr, -1)
                b.Frame = matches[0]
                b.Memory = matches[1]
                b.Peak = matches[2]

            } else if timetest.MatchString(substr) {

                // extract the values
                re := regexp.MustCompile("(?i)\\d\\d:\\d\\d\\.\\d\\d")
                matches := re.FindAllString(substr, -1)
                b.Time = matches[0]
            }
          }

          // extract the status note
          b.Note = str[2]

          // log event message
          logger.Manager.Package["node"].Trace().Msg("The current render status is:")
          logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Current frame: %v", b.Frame))
          logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Memory usage: %v", b.Memory))
          logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Peak memory usage: %v", b.Peak))
          logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Render time: %v", b.Time))
          logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Status note: %v", b.Note))

      }

      // Print the command line output of Blender
      // fmt.Printf("(blender) > %v: %v \n", name, line)

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
                for _, blender := range nm.Renderer.Offer.Blender {
                  fmt.Printf(" [#] Version: %v (Engines: %v | Devices: %v) \n", blender.BuildVersion, strings.Join(blender.Engines, ", "), strings.Join(blender.Devices, ", "))
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

                // Add a new Blender version to the node's render offer
                err := nm.Renderer.Offer.AddBlenderVersion(version, path, &engines, &devices)
                if err != nil {
                    fmt.Println("")
                    fmt.Println(err)
                } else {

                    fmt.Printf("Added the Blender v'%v' with path '%v' to the render offer. \n", nm.Renderer.Offer.Blender[version].BuildVersion, path)

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

    // add command flag parameters
    command.Flags().StringVarP(&version, "version", "v", "", "The version of Blender to be used")
    command.Flags().StringVarP(&path, "path", "p", "", "The path to a Blender executable on this computer")
    command.Flags().StringSliceVarP(&engines, "engines", "E", GetBlenderEngineString([]uint8{BLENDER_RENDER_ENGINE_OPTIONS}), "The supported engines (only EEVEE and CYCLES)")
    command.Flags().StringSliceVarP(&devices, "devices", "D", GetBlenderDeviceString([]uint8{BLENDER_RENDER_DEVICE_OPTIONS}), "The supported devices for rendering (all GPU options may be combined with '+CPU' for hybrid rendering)")

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

                // if the parsed version is supported by this node
                if blender, ok := nm.Renderer.Offer.Blender[version]; ok {
                    fmt.Println("")
                    fmt.Printf("Starting Blender v%v. \n", blender.BuildVersion)
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

            // if no version argument was parsed
            } else {

              fmt.Println("")
              fmt.Println(fmt.Errorf("Cannot run Blender, because no version was passed."))
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

    // add command flag parameters
    command.Flags().StringVarP(&version, "version", "v", "", "The version of Blender to be used")
    command.Flags().StringVarP(&param, "param", "p", "", "The command line options for Blender")

    return command

}

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
  "time"
  "path/filepath"
  "encoding/json"
	"sort"

  // external
  // hederasdk "github.com/hashgraph/hedera-sdk-go/v2"
  "github.com/mattn/go-shellwords"
  "github.com/spf13/cobra"
  // "github.com/cockroachdb/apd"
  // "golang.org/x/exp/slices" <-- would be handy, but requires Go 1.18; TODO: Update possible for Hedera SDK?

  // internal
  . "renderhive/globals"
  . "renderhive/utility"
  "renderhive/logger"

)



// RENDER JOBS, OFFERS, AND REQUESTS
// #############################################################################
// App data of supported Blender version
type BlenderAppData struct {

  // Blender app and build info
  Path string                         // Path to the Blender app
  BuildVersion string                 // Build version of this Blender app
  BuildHash string                    // Build hast of this Blender app
  BuildDate string                    // Build date of this Blender app
  BuildTime string                    // Build time of this Blender app

  // Render settings supported by this node's Blender instance
  Engines []string                    // Supported render engines
  Devices []string                    // Supported devices
  Threads uint8                       // Supported number of threads

  // Process status
  Cmd *exec.Cmd                       // pointer to the exec.Command type
  Param []string                      // Command line options this Blender process was called with
  PID int                             // PID of the process
  Running bool                        // Is the process still running
  StdOut io.ReadCloser                // Command-line standard output of the Blender app
  StdErr io.ReadCloser                // Command-line error output of the Blender app

  // Blender render status
  Frame string                        // Current frame number
  Memory string                       // Current memory usage
  Peak string                         // Peak memory usage
  Time string                         // Render time
  Note string                         // Render status note

  // Blender benchmarks
  BenchmarkTool *BlenderBenchmarkTool // Blender benchmark results

}

// Blender file data
type BlenderFileData struct {

  // TODO: Fill with information
  // General info
  CID string                  // Content identifier (CID) of the .blend file on the IPFS
  Path string                 // Local path to the Blender file

  // Render settings
  Settings RenderSettings     // Render settings of this Blender file

}

// Blender benchmark result
// This struct is a wrapper for the JSON schema returned by the benchmark tool
type BlenderBenchmarkResult struct {

  Timestamp time.Time `json:"timestamp"`
  BlenderVersion struct {
      Version string `json:"version"`
      BuildDate string `json:"build_date"`
      BuildTime string `json:"build_time"`
      BuildCommitDate string `json:"build_commit_date"`
      BuildCommitTime string `json:"build_commit_time"`
      BuildHash string `json:"build_hash"`
      Label string `json:"label"`
      Checksum string `json:"checksum"`
  } `json:"blender_version"`
  BenchmarkLauncher struct {
      Label string `json:"label"`
      Checksum string `json:"checksum"`
  } `json:"benchmark_launcher"`
  BenchmarkScript struct {
      Label string `json:"label"`
      Checksum string `json:"checksum"`
  } `json:"benchmark_script"`
  Scene struct {
      Label string `json:"label"`
      Checksum string `json:"checksum"`
  } `json:"scene"`
  SystemInfo struct {
      Bitness string `json:"bitness"`
      Machine string `json:"machine"`
      System string `json:"system"`
      DistName string `json:"dist_name"`
      DistVersion string `json:"dist_version"`
      Devices []struct {
          Type string `json:"type"`
          Name string `json:"name"`
      } `json:"devices"`
      NumCpuSockets int `json:"num_cpu_sockets"`
      NumCpuCores int `json:"num_cpu_cores"`
      NumCpuThreads int `json:"num_cpu_threads"`
  } `json:"system_info"`
  DeviceInfo struct {
      DeviceType string `json:"device_type"`
      ComputeDevices []struct {
          Type string `json:"type"`
          Name string `json:"name"`
      } `json:"compute_devices"`
      NumCpuThreads int `json:"num_cpu_threads"`
  } `json:"device_info"`
  Stats struct {
      DevicePeakMemory float64 `json:"device_peak_memory"`
      NumberOfSamples int `json:"number_of_samples"`
      TimeForSamples float64 `json:"time_for_samples"`
      SamplesPerMinute float64 `json:"samples_per_minute"`
      TotalRenderTime float64 `json:"total_render_time"`
      RenderTimeNoSync float64 `json:"render_time_no_sync"`
      TimeLimit float64 `json:"time_limit"`
  } `json:"stats"`

  // NOTE: The Blender Benchmark Score on OpenData is given in samples per minute
  //       and is the sum of the SamplesPerMinute over all benchmark scenes.

}

type BlenderBenchmarkTool struct {

  Result []BlenderBenchmarkResult

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
  // General info
  ID int                               // Internal ID for the render request management
  DocumentCID string                   // content identifier (CID) of the render request document on the IPFS
  DocumentPath string                  // local path of the render request document on this node
  CreatedTimestamp time.Time           // The datetime this request was created
  ModifiedTimestamp time.Time          // The datetime this request was last modified
  PublishedTimestamp time.Time         // The datetime this request was published to the render hive
  FinishedTimestamp time.Time          // The datetime this request was completely rendered by the render hive

  // File data
  BlenderFile BlenderFileData          // data of the Blender file to be rendered

  // Render request data
  // TODO: Prices need to be implemented using Decimals instead float ("apd" package or "currency" package?)
  Version string                       // Blender version the job should be rendered on
  Price float64                        // Price maximum in cents (USD) per BBP
  ThisNode bool                        // True, if this node participates in rendering this job

}

// a render offer that is provided by this node for rendering on the renderhive
type RenderOffer struct {

  // TODO: Fill with information
  // General offer information
  UserID int                           // ID of the user this offer belongs to
  DocumentCID string                   // content identifier (CID) of the render offer document on the IPFS
  DocumentPath string                  // local path of the render offer document on this node
  ModifiedTimestamp time.Time          // The datetime this request was last modified
  PublishedTimestamp time.Time         // The datetime this request was published to the render hive
  Paused bool                          // True, if the offer is currently paused and no new render jobs are accepted

  // Render offer data
  // TODO: Prices need to be implemented using Decimals instead float ("apd" package or "currency" package?)
  Blender map[string]BlenderAppData    // supported Blender versions and Blender render options (includes benchmark results, i.e. "offered render power" per version)
  Price float64                        // price threshold in cents (USD) per BBP for rendering
  Tax []struct {                       // Some jurisdictions may require taxation for the services offered on the render hive by a node

    Name string                        // Name of the tax (e.g., Sales Tax, VAT, etc.)
    Description string                 // Description of the text
    Value float64                      // Tax value in %

  }

  // Terms of Service
  // Each node can allow/disallow certain

}



// RENDER OFFERS
// #############################################################################
// Initialize the render offer for this node
func (nm *PackageManager) InitRenderOffer() *RenderOffer {

  // initialize the node's render offer
  nm.Renderer.Offer = &RenderOffer{}
  nm.Renderer.Offer.Blender = map[string]BlenderAppData{}

  return nm.Renderer.Offer

}

// Set the render price limit
func (ro *RenderOffer) SetPrice(price float64, currency string) (error) {
    var err error

    // Set the new price
    ro.Price = price

    return err

}

// Get the render price limit
func (ro *RenderOffer) GetPrice() (float64) {

    return ro.Price

}

// Add a Blender version to the render offer
func (ro *RenderOffer) AddBlenderVersion(version string, path string, engines *[]string, devices *[]string, threads uint8) (error) {
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
                  Threads: threads,
                  BenchmarkTool: &BlenderBenchmarkTool{},
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

// create a new render offer document (locally) and pin it to the local IPFS node
// NOTE: This document will be downloaded by all render nodes
func (ro *RenderOffer) Publish() (error) {
  var err error

  // log event
  logger.Manager.Package["node"].Trace().Msg("Remove a Blender version from the node's render offer:")

  // Create the render offer document JSON file
  // ...

  // Pin the file on the local IPFS
  // ...

  // Update the internal CID
  ro.DocumentCID = ""

  return err

}



// RENDER REQUEST
// #############################################################################
// Create a new render render request from this node
func (nm *PackageManager) AddRenderRequest(request *RenderRequest) (int, error) {
  var err error
  var newID int

  // initialize ID
  newID = 0

  // if no request was passed
  if request == nil {
      return 0, errors.New(fmt.Sprintf("No request was passed."))
  }

  // if the node has any render requests
  if nm.Renderer.Requests != nil {

    	// convert map to slice of keys
    	ids := make([]int, 0, len(nm.Renderer.Requests))
    	for id, _ := range nm.Renderer.Requests {
    		ids = append(ids, id)
    	}

    	// sort the slice of IDs
    	sort.Ints(ids)

      // Get the ID assigned to the last render request and add 1 for the new ID
      if len(ids) > 0 {
          newID = ids[len(ids) - 1] + 1
      }

  } else {

      // initialized the map first
      nm.Renderer.Requests = map[int]RenderRequest{}

  }

  // Update the ID
  request.ID = newID

  // Append the request to the list of requests of this node
  nm.Renderer.Requests[newID] = *request

  return newID, err

}

// Remove a render request from the node
func (nm *PackageManager) RemoveRenderRequest(id int) (error) {
    var err error

    // log event
    logger.Manager.Package["node"].Trace().Msg("Removing a render request from the node:")

    // delete the element from the map, if it exists
    request, ok := nm.Renderer.Requests[id]
    if ok {
        logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] ID: %v", request.ID))
        delete(nm.Renderer.Requests, id)
    } else {
        err = errors.New(fmt.Sprintf("Render request %v could not be removed from the node.", id))
    }

    return err

}


// BLENDER BENCHMARK TOOL CONTROL
// #############################################################################
// Execute the command line interface for the Blender benchmark tool
func (tool *BlenderBenchmarkTool) _execute(path string, args []string) (string, error) {
    var err error

    // Check if 'path' is pointing to an existing file
    if _, err = os.Stat(path); os.IsNotExist(err) {
        return "", err
    }

    // get supported blender versions
    cmd := exec.Command(path, args...)
    output, err := cmd.Output()
    if err != nil {
        fmt.Println("Error:",err)
        return "", err
    }

    return string(output), err

}

// Run the Blender benchmark tool with the specified Blender version and
// rendering device
func (tool *BlenderBenchmarkTool) Run(ro *RenderOffer, benchmark_version string, benchmark_device string, benchmark_scene string) (error) {
    var err error
    var versions []string
    var device_names []string
    var device_types []string
    var scenes []string
    var ok bool

    // if the Blender version is supported by this node
    blender, ok := ro.Blender[benchmark_version]
    if ok {

        // log event
        logger.Manager.Package["node"].Debug().Msg("Benchmarking a supported Blender version:")

        // path to te benchmark tool
        path, _ := filepath.Abs("benchmark/benchmark-launcher-cli")

        // Check if 'path' is pointing to an existing file
        if _, err = os.Stat(path); os.IsNotExist(err) {
            return err
        }

        // get list of Blender versions supported by this tool version
        output, err := tool._execute(path, []string{"blender", "list"})
        if err != nil {
            return errors.New(fmt.Sprintf("Could not retrieve Blender benchmark tool version list. (Error: %v)", err))
        } else {

            // scan lines
            scanner := bufio.NewScanner(strings.NewReader(output))
            for scanner.Scan() {
                // get version
                supported_version := strings.Fields(scanner.Text())
                versions = append(versions, supported_version[0])
            }
        }

        // check if 'benchmark_version' is supported
        ok = InStringSlice(versions, benchmark_version)
        if !ok {
            return errors.New(fmt.Sprintf("Blender v%v is not supported by this Blender benchmark tool.", benchmark_version))
        }

        // download the suitable Blender version
        output, err = tool._execute(path, []string{"blender", "download", benchmark_version})
        if err != nil {
            return errors.New(fmt.Sprintf("Could not download blender version %v. (Error: %v)", benchmark_version, err))
        }

        // log trace event
        logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] Retrieving supported devices for Blender version: %v", benchmark_version))

        // get list of devices
        output, err = tool._execute(path, []string{"devices", "--blender-version", benchmark_version, "list"})
        if err != nil {
            return errors.New(fmt.Sprintf("Could not retrieve Blender benchmark tool device list. (Error: %v)", err))
        } else {

          // scan lines
          scanner := bufio.NewScanner(strings.NewReader(output))
          for scanner.Scan() {

              // get version
              device := strings.Fields(scanner.Text())
              device_names = append(device_names, device[:len(device)-1]...)
              device_types = append(device_types, device[len(device)-1])

              // log trace event
              logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] [*] Supported device: %v (%v)", device[:len(device)-1], device[len(device)-1]))

          }

        }

        // check if 'benchmark_device' is supported
        ok = (InStringSlice(device_names, benchmark_device) || InStringSlice(device_types, benchmark_device))
        if !ok {
            return errors.New(fmt.Sprintf("Device '%v' is not supported by this Blender benchmark tool.", benchmark_device))
        } else if benchmark_device == "" {
            return errors.New(fmt.Sprintf("No device was specified for the benchmark rendering."))
        }

        // log trace event
        logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] Retrieving supported scenes for Blender version: %v", benchmark_version))

        // get list of benchmark scenes
        output, err = tool._execute(path, []string{"scenes", "--blender-version", benchmark_version, "list"})
        if err != nil {
            return errors.New(fmt.Sprintf("Could not retrieve Blender benchmark tool scene list. (Error: %v)", err))
        } else {

          // scan lines
          scanner := bufio.NewScanner(strings.NewReader(output))
          for scanner.Scan() {

            // get version
            scene := strings.Fields(scanner.Text())
            scenes = append(scenes, scene[0])

            // log trace event
            logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] [*] Supported scene: %v", scene[0]))

          }
        }

        // check if 'benchmark_scene' is supported
        ok = InStringSlice(scenes, benchmark_scene)
        if !ok {
            return errors.New(fmt.Sprintf("Scene '%v' is not supported by this Blender benchmark tool.", benchmark_scene))
        } else if benchmark_scene == "" {
            return errors.New(fmt.Sprintf("No scene was specified for the benchmark rendering."))
        } else {

            // download the scene
            output, err = tool._execute(path, []string{"scenes", "download", "--blender-version", benchmark_version, benchmark_scene})
            if err != nil {
                return errors.New(fmt.Sprintf("Could not download Blender benchmark scene '%v'. (Error: %v)", benchmark_scene, err))
            }

            // log trace event
            logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] Downloaded Benchmark scene '%v' and started benchmark rendering ...", benchmark_scene))

            // start the benchmark
            output, err = tool._execute(path, []string{"benchmark", "--blender-version", benchmark_version, "--device-type", "CPU", "--json", benchmark_scene})
            if err != nil {
                return errors.New(fmt.Sprintf("Failed to execute benchmark rendering for scene '%v'. (Error: %v)", benchmark_scene, err))
            } else {

              // parse the benchmark result
              json.Unmarshal([]byte(output), &tool.Result)

              // log trace event
              logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] [*] Benchmark result: %v samples / min", tool.Result[0].Stats.SamplesPerMinute))

            }

        }

        // TODO: store the Benchmark result locally as JSON file and on IPFS
        // ...

    } else {
        err = errors.New(fmt.Sprintf("Blender v'%v' is not in the node's render offer.", blender.BuildVersion))
    }

    return err

}


// BLENDER CONTROL
// #############################################################################
// TODO: Some notes on things to implement
//
//       1) A set of Python scripts could be stored on IPFS (immutability
//          guarenteed by CID), which nodes can execute with the "--python" flag
//          of Blender. These could be things like a script to set the region to
//          render.
//

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

// COMMAND LINE INTERFACE - RENDER REQUESTS & OFFERS
// #############################################################################
// Create the CLI command to manage the render requests of this node
func (nm *PackageManager) CreateCommandRequest() (*cobra.Command) {

    // flags for the 'request' command
    var list bool

    // create a 'blender' command for the node
    command := &cobra.Command{
    	Use:   "request",
    	Short: "Manage the node's render requests",
    	Long: "This command is for adding/removing/editing the render requests of this node.",
      Run: func(cmd *cobra.Command, args []string) {

        // if a render offer exists
        if nm.Renderer.Requests != nil {

            // list all Blender versions
            if list {

                fmt.Println("")
                fmt.Println("The node has the following render requests:")

                // find each Blender version added to the node
                for _, request := range nm.Renderer.Requests {
                  fmt.Printf(" [#] ID: %v for %v \n", request.ID, request.BlenderFile.Path)
                }
                fmt.Println("")

            }

        } else {

          fmt.Println("")
          fmt.Println(fmt.Errorf("The node has no render requests."))
          fmt.Println("")
        }

        return

    	},
    }

    // add command flags
    command.Flags().BoolVarP(&list, "list", "l", false, "List all the render requests of the node")


    // add the subcommands
    command.AddCommand(nm.CreateCommandRequest_Add())
    command.AddCommand(nm.CreateCommandRequest_Remove())
    // command.AddCommand(nm.CreateCommandRequest_Publish())
    // command.AddCommand(nm.CreateCommandRequest_Pause())
    // command.AddCommand(nm.CreateCommandRequest_Revoke())

    return command

}

// Create the CLI command to add a new render request for this node
func (nm *PackageManager) CreateCommandRequest_Add() (*cobra.Command) {

    // flags for the 'request add' command
    var blender_version string
    var blender_file string
    var render_price float64
    var this_node bool

    // create a 'request add' command for the node
    command := &cobra.Command{
    	Use:   "add",
    	Short: "Add a Blender version to the node's render offer",
    	Long: "This command is for adding a Blender version to the node's render offer.",
      Run: func(cmd *cobra.Command, args []string) {

        // if the map was not correctly initialized
        if nm.Renderer.Requests != nil {

            // add a Blender version
            if (blender_version != "" && blender_file != "" && render_price > 0) {
                fmt.Println("")

                // Check if path is pointing to an existing blender file
                if _, err := os.Stat(blender_file); os.IsNotExist(err) {
                    fmt.Println(fmt.Errorf("The given path '%v' is not a valid path.", blender_file))
                    return
                }

                // Create a new render request
                request := RenderRequest{
                              BlenderFile: BlenderFileData{Path: blender_file},
                              Version: blender_version,
                              Price: render_price,
                              ThisNode: this_node,
                           }

                // Add the render request to the node
                id, err := nm.AddRenderRequest(&request)

                if err != nil {
                    fmt.Println("")
                    fmt.Println(err)
                } else {

                    fmt.Println("Added a new render request to the node:")
                    fmt.Printf(" [#] ID: %v\n", id)
                    fmt.Printf(" [#] Blender file: %v\n", blender_file)
                    fmt.Printf(" [#] Requested Blender version: %v\n", blender_version)
                    fmt.Printf(" [#] Maximum price: %v USD / BBP \n", render_price)
                    fmt.Printf(" [#] Node participates: %v \n", this_node)

                }
                fmt.Println("")

            } else {

              fmt.Println("")
              fmt.Println(fmt.Errorf("Failed to create the render request."))
              if blender_version == "" { fmt.Println(fmt.Errorf(" [#] Missing a required parameter: Blender version (--blender-version).")) }
              if blender_file == "" { fmt.Println(fmt.Errorf(" [#] Missing a required parameter: Blender file (--blender-file).")) }
              if render_price == 0 { fmt.Println(fmt.Errorf(" [#] Missing a required parameter: Maximum render price (--render-price).")) }
              fmt.Println("")

            }

        } else {

          fmt.Println("")
          fmt.Println(fmt.Errorf("There was an error in initializing the render requests."))
          fmt.Println("")

        }

        return

    	},
    }

    // add command flag parameters
    command.Flags().StringVarP(&blender_version, "blender-version", "v", "", "The Blender version to be used for rendering")
    command.Flags().StringVarP(&blender_file, "blender-file", "f", "", "The path to the Blender file to be rendered")
    command.Flags().Float64VarP(&render_price, "render-price", "p", 0, "The maximum price the node will pay for rendering")
    command.Flags().BoolVarP(&this_node, "this-node", "t", false, "Set if this node shall participate in rendering its own request")

    return command

}


// Create the CLI command to remove a previously created render request from this node
func (nm *PackageManager) CreateCommandRequest_Remove() (*cobra.Command) {

    // flags for the 'request remove' command
    var id int

    // create a 'request remove' command for the node
    command := &cobra.Command{
    	Use:   "remove",
    	Short: "Remove a render request from this node",
    	Long: "This command is for removing a render request from this node. In case it was submitted to the network, it will be cancelled and revoked.",
      Run: func(cmd *cobra.Command, args []string) {

        // if render requests exists
        if nm.Renderer.Requests != nil {

            // was a valid ID passed?
            if (id != -1) {

                // if the parsed version is supported by the node
                _, ok := nm.Renderer.Requests[id]
                if ok {

                    // Remove the render request
                    err := nm.RemoveRenderRequest(id)
                    if err != nil {

                      fmt.Println("")
                      fmt.Println(fmt.Errorf(err.Error()))
                      fmt.Println("")

                      return

                    }

                    fmt.Println("")
                    fmt.Printf("Removed render request with ID %v from this node. \n", id)
                    fmt.Println("")

                } else {

                  fmt.Println("")
                  fmt.Println(fmt.Errorf("There is no render request with ID %v.", id))
                  fmt.Println("")

                }

            } else {

              fmt.Println("")
              fmt.Println(fmt.Errorf("Failed to remove the render request."))
              if id == -1 { fmt.Println(fmt.Errorf(" [#] Missing a required parameter: Request ID (--request-id).")) }
              fmt.Println("")

            }

        } else {

          fmt.Println("")
          fmt.Println(fmt.Errorf("The node has no render requests."))
          fmt.Println("")

        }

        return

    	},
    }

    // add command flag parameters
    command.Flags().IntVarP(&id, "request-id", "i", -1, "The ID of the render request to remove")

    return command

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
    command.AddCommand(nm.CreateCommandBlender_Benchmark())

    return command

}

// Create the CLI command to add a Blender version to the node's render offer
func (nm *PackageManager) CreateCommandBlender_Add() (*cobra.Command) {

    // flags for the 'blender' command
    var version string
    var path string
    var engines []string
    var devices []string
    var threads uint8

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
                err := nm.Renderer.Offer.AddBlenderVersion(version, path, &engines, &devices, threads)
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
    command.Flags().Uint8VarP(&threads, "threads", "t", 1, "The supported number of threads rendered simultaneously by this Blender version (default: 1)")

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

// Create the CLI command to run a Blender benchmark with the Blender benchmark
// command line interface tool
func (nm *PackageManager) CreateCommandBlender_Benchmark() (*cobra.Command) {

    // flags for the 'blender benchmark' command
    var version string
    var use_tool bool
    var scene string
    var device string

    // create a 'blender remove' command for the node
    command := &cobra.Command{
    	Use:   "benchmark",
    	Short: "Run a Blender benchmark",
    	Long: "This command is for starting a benchmark rendering for a particular Blender version supported by this node.",
      Run: func(cmd *cobra.Command, args []string) {

        // if a render offer exists
        if nm.Renderer.Offer != nil {

            // if a version was parsed and
            if len(version) != 0 {

                // if the parsed version is supported by this node
                if blender, ok := nm.Renderer.Offer.Blender[version]; ok {

                    // if the official Blender benchmark tool shall be used
                    if use_tool {

                        // run the this Blender version
                        err := blender.BenchmarkTool.Run(nm.Renderer.Offer, version, device, scene)
                        if err  != nil {
                            // log error event
                            logger.Manager.Package["node"].Error().Msg(err.Error())
                        }
                    }

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
    command.Flags().BoolVarP(&use_tool, "blender-benchmark", "B", true, "Use the official Blender benchmark tool (default: yes)")
    command.Flags().StringVarP(&scene, "scene", "S", "", "The scene(s) to be used for the benchmark rendering")
    command.Flags().StringVarP(&device, "device", "D", "", "The device(s) to be used for the benchmark rendering")

    return command

}

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

import (

	// standard
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	// external

	hederasdk "github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/ipfs/boxo/files"
	"github.com/mattn/go-shellwords"
	"github.com/spf13/cobra"

	// "github.com/cockroachdb/apd"
	// "golang.org/x/exp/slices" <-- would be handy, but requires Go 1.18; TODO: Update possible for Hedera SDK?

	// internal
	. "renderhive/globals"
	"renderhive/hedera"
	"renderhive/ipfs"
	"renderhive/logger"
	. "renderhive/utility"
)

// RENDER JOBS, OFFERS, AND REQUESTS
// #############################################################################

// App data of supported Blender version
type BlenderAppData struct {

	// Blender app and build info
	Path         string // Path to the Blender app on the local filesystem
	BuildVersion string // Build version of this Blender app
	BuildHash    string // Build hash of this Blender app
	BuildDate    string // Build date of this Blender app
	BuildTime    string // Build time of this Blender app

	// Render settings supported by this node's Blender instance
	Engines []string // Supported render engines
	Devices []string // Supported devices
	Threads uint8    // Supported number of threads

	// Process status
	Cmd     *exec.Cmd     // pointer to the exec.Command type
	Param   []string      // Command line options this Blender process was called with
	PID     int           // PID of the process
	Running bool          // Is the process still running
	StdOut  io.ReadCloser // Command-line standard output of the Blender app
	StdErr  io.ReadCloser // Command-line error output of the Blender app

	// Blender render status
	Frame  string // Current frame number
	Memory string // Current memory usage
	Peak   string // Peak memory usage
	Time   string // Render time
	Note   string // Render status note

	// Blender benchmarks
	BenchmarkTool *BlenderBenchmarkTool // Blender benchmark results

}

// Blender file data
type BlenderFileData struct {

	// TODO: Fill with information
	// General info
	CID  string // Content identifier (CID) of the .blend file on the IPFS
	Path string // Local path to the Blender file

	// Render settings
	Settings RenderSettings // Render settings of this Blender file

}

// Blender benchmark result
// This struct is a wrapper for the JSON schema returned by the benchmark tool
type BlenderBenchmarkResult struct {
	Timestamp      time.Time `json:"timestamp"`
	BlenderVersion struct {
		Version         string `json:"version"`
		BuildDate       string `json:"build_date"`
		BuildTime       string `json:"build_time"`
		BuildCommitDate string `json:"build_commit_date"`
		BuildCommitTime string `json:"build_commit_time"`
		BuildHash       string `json:"build_hash"`
		Label           string `json:"label"`
		Checksum        string `json:"checksum"`
	} `json:"blender_version"`
	BenchmarkLauncher struct {
		Label    string `json:"label"`
		Checksum string `json:"checksum"`
	} `json:"benchmark_launcher"`
	BenchmarkScript struct {
		Label    string `json:"label"`
		Checksum string `json:"checksum"`
	} `json:"benchmark_script"`
	Scene struct {
		Label    string `json:"label"`
		Checksum string `json:"checksum"`
	} `json:"scene"`
	SystemInfo struct {
		Bitness     string `json:"bitness"`
		Machine     string `json:"machine"`
		System      string `json:"system"`
		DistName    string `json:"dist_name"`
		DistVersion string `json:"dist_version"`
		Devices     []struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"devices"`
		NumCpuSockets int `json:"num_cpu_sockets"`
		NumCpuCores   int `json:"num_cpu_cores"`
		NumCpuThreads int `json:"num_cpu_threads"`
	} `json:"system_info"`
	DeviceInfo struct {
		DeviceType     string `json:"device_type"`
		ComputeDevices []struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"compute_devices"`
		NumCpuThreads int `json:"num_cpu_threads"`
	} `json:"device_info"`
	Stats struct {
		DevicePeakMemory float64 `json:"device_peak_memory"`
		NumberOfSamples  int     `json:"number_of_samples"`
		TimeForSamples   float64 `json:"time_for_samples"`
		SamplesPerMinute float64 `json:"samples_per_minute"`
		TotalRenderTime  float64 `json:"total_render_time"`
		RenderTimeNoSync float64 `json:"render_time_no_sync"`
		TimeLimit        float64 `json:"time_limit"`
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
	Engine      string // Render engine to be used (i.e., Cycles, EEVEE)
	FeatureSet  string // Blender feature set to be used
	Device      string // CPU, GPU or both?
	ResolutionX int    // x resolution of the render result
	ResolutionY int    // y resolution of the render result
	TileX       int    // x resolution of tiles to be rendered
	TileY       int    // y resolution of tiles to be rendered

	OutputPath string // Output path (includes file naming)

}

// a render job claimed for rendering on the render hive by this node
type RenderJob struct {

	// TODO: Fill with information
	UserID int // ID of the user this render job belongs to
	NodeID int // ID of the node this render job belongs to

	// Request data
	Request *RenderRequest // Render request
	// Job status
	// ...

}

// a render job that is requested by this node for rendering on the render hive
// NOTE: Fields with the `json:"-"` tag are not included in the JSON representation
type RenderRequest struct {

	// TO BE DEPRECATED
	ID int `json:"-"` // Internal ID for the render request management

	// General info
	DocumentCID        string    `json:"-"` // content identifier (CID) of the render request document on the IPFS
	DocumentPath       string    `json:"-"` // local path of the render request document on this node
	DirectoryCID       string    // content identifier (CID) of the render request directory on IPFS
	CreatedTimestamp   time.Time // The datetime this request was created
	ModifiedTimestamp  time.Time // The datetime this request was last modified
	SubmittedTimestamp time.Time `json:"-"` // The datetime this request was submitted to the network
	ClosedTimestamp    time.Time `json:"-"` // The datetime this request was closed (finished, cancelled, etc)

	// Project files
	Files       map[string]files.Node `json:"-"`
	Directory   files.Directory       `json:"-"`
	BlenderFile BlenderFileData       // data of the Blender file to be rendered

	// Render request data
	// TODO: Prices need to be implemented using Decimals instead float ("apd" package or "currency" package?)
	Version   string  // Blender version the job should be rendered on
	Price     float64 // Price maximum in cents (USD) per BBP
	ThisNode  bool    // True, if this node participates in rendering this job
	Cancelled bool    `json:"-"` // True, if the render request was cancelled

	// Hedera data
	Owner   *hederasdk.AccountID          // Account ID of the operator who created this render request
	Receipt *hederasdk.TransactionReceipt `json:"-"` // Transaction receipt of the render request submission
}

// Representation of the JSON message for the Job Queue Topic
type RenderRequestMessage struct {

	// TODO: Add version info?
	// General info
	DocumentCID    string `json:"document_cid"`     // Render request document CID
	BlenderFileCID string `json:"blender_file_cid"` // Blender file CID

}

// Supported Blender versions
type RenderOfferBlenderVersions struct {
	Version string   // Blender version
	Engines []string // Render engines supported with this offer
	Devices []string // Devices supported with this offer
	Threads uint8    // Threads supported with this offer
}

// a render offer that is provided by this node for rendering on the render hive
type RenderOffer struct {

	// General offer information
	DocumentCID        string    `json:"-"` // content identifier (CID) of the render offer document on the IPFS
	DocumentPath       string    `json:"-"` // local path of the render offer document on this node
	CreatedTimestamp   time.Time // The datetime this offer was created
	ModifiedTimestamp  time.Time // The datetime this offer was last modified
	SubmittedTimestamp time.Time `json:"-"` // The datetime this offer was submitted to the network
	PausedTimestamp    time.Time `json:"-"` // The datetime this offer was paused

	// Render offer data
	// TODO: Prices need to be implemented using Decimals instead float ("apd" package or "currency" package?)
	BlenderVersions []RenderOfferBlenderVersions // Blender versions supported with this offer
	Price           float64                      // Psrice threshold in cents (USD) per BBP for rendering
	// Tax     []struct {                // Some jurisdictions may require taxation for the services offered on the render hive by a node

	// 	Name        string  // Name of the tax (e.g., Sales Tax, VAT, etc.)
	// 	Description string  // Description of the text
	// 	Value       float64 // Tax value in %

	// }
	Paused bool `json:"-"` // True, if the offer is currently paused

	// Terms of Service
	// Each node can allow/disallow certain

	// Blender data
	Blender map[string]BlenderAppData `json:"-"` // supported Blender versions and Blender render options (includes benchmark results, i.e. "offered render power" per version)

	// Hedera data
	Owner   *hederasdk.AccountID          // Account ID of the operator who created this render offer
	Receipt *hederasdk.TransactionReceipt `json:"-"` // Transaction receipt of the last transaction of this render offer
}

// RENDER OFFERS
// #############################################################################
// Initialize the render offers for this node
func (nm *PackageManager) InitRenderOffers() error {
	var err error

	// initialize the node's render offers
	nm.Renderer.Offers = make(map[string]*RenderOffer)

	// load the render offers from the local file system
	err = nm.LoadRenderOffers()
	if err != nil {
		return errors.New(fmt.Sprintf("Could not load render offers: %v", err))
	}

	return err

}

// Load a render offer into memory
func (nm *PackageManager) LoadRenderOffers() error {
	var err error

	// path to the local render offer documents
	offer_document_directory := filepath.Join(GetAppDataPath(), RENDERHIVE_APP_DIRECTORY_LOCAL_OFFERS)

	// if the directory does NOT exist
	if _, err := os.Stat(offer_document_directory); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Render offer directory '%v' does not exist.", offer_document_directory))
	}

	// go through all files in the directory
	err = filepath.Walk(offer_document_directory, func(path string, info os.FileInfo, err error) error {

		// if the file is a regular file
		if info.Mode().IsRegular() {

			// if the file is a render offer document
			if matched, _ := regexp.MatchString(`^offer-.*\.json$`, info.Name()); matched {

				// load the render offer from the file
				err = nm.LoadRenderOfferFromFile(path)

			}

		}

		// log error event
		if err != nil {
			logger.Manager.Package["node"].Error().Msg(fmt.Sprintf("Could not load render offer %v: %v", path, err))
		}

		// reset error
		err = nil

		return nil

	})

	return err

}

// RENDER OFFER OBJECTS
// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

// Load a render offer into memory
func (nm *PackageManager) LoadRenderOfferFromFile(path string) error {
	var err error

	// if the offer does NOT exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Render offer document '%v' does not exist.", path))
	}

	// get the CID of the render offer document
	offer_document_cid, err := ipfs.Manager.GetHashFromPath(path)
	if err != nil {
		return err
	}

	// load the render offer document file
	offer_document_file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer offer_document_file.Close()

	// decode the render offer data from the file
	// NOTE: The *hedera.AccountID is not supported by the JSON decoder.
	//		 Therefore, the Owner field is decoded manually using this workaround.
	var offer struct {
		RenderOffer
		Owner struct {
			Shard   uint64 `json:"Shard"`
			Realm   uint64 `json:"Realm"`
			Account uint64 `json:"Account"`

			// TODO: Currently, the AliasKey and AliasEvmAddress are ignored
			// AliasKey        hederasdk.PublicKey `json:"AliasKey"`
			// AliasEvmAddress []byte              `json:"AliasEvmAddress"`
		} `json:"Owner"`
	}
	decoder := json.NewDecoder(offer_document_file)
	err = decoder.Decode(&offer)
	if err != nil {
		return err
	}

	// create the render offer object
	nm.Renderer.Offers[offer_document_cid] = &RenderOffer{
		DocumentCID:       offer_document_cid,
		DocumentPath:      path,
		CreatedTimestamp:  offer.CreatedTimestamp,
		ModifiedTimestamp: offer.ModifiedTimestamp,
		BlenderVersions:   offer.BlenderVersions,
		Price:             offer.Price,
		Blender:           make(map[string]BlenderAppData),
		Owner: &hederasdk.AccountID{
			Shard:   offer.Owner.Shard,
			Realm:   offer.Owner.Realm,
			Account: offer.Owner.Account,
		},
		Receipt: offer.Receipt,
	}

	// add all Blender versions to the offer
	for _, blender := range offer.BlenderVersions {
		nm.Renderer.Offers[offer_document_cid].AddBlenderVersion(blender.Version, &blender.Engines, &blender.Devices, blender.Threads)
	}

	return nil

}

// Create a new render offer object for this node
func (nm *PackageManager) NewRenderOffer(render_price float64) (*RenderOffer, error) {
	var err error

	// create the render offer object
	offer := &RenderOffer{
		CreatedTimestamp:  time.Now(),
		ModifiedTimestamp: time.Now(),
		BlenderVersions:   []RenderOfferBlenderVersions{},
		Price:             render_price,
		Blender:           make(map[string]BlenderAppData),

		Owner: &Manager.User.UserAccount.AccountID,
	}

	return offer, err

}

// Set the node's active render offer object
func (nm *PackageManager) SetActiveRenderOffer(offer *RenderOffer) error {
	var err error
	var ok bool

	// if the offer does NOT exist
	if nm.Renderer.ActiveOffer, ok = nm.Renderer.Offers[offer.DocumentCID]; !ok {
		return errors.New(fmt.Sprintf("Render offer with CID '%v' does not exist.", offer.DocumentCID))
	}

	return err

}

// Get the node's active render offer object
func (nm *PackageManager) GetActiveRenderOffer() *RenderOffer {

	return nm.Renderer.ActiveOffer

}

// Get the render offer object from the render offer document CID
func (nm *PackageManager) GetRenderOffer(document_cid string) (*RenderOffer, error) {
	var err error

	// Get the Render Offer from the CID of the render offer document
	offer, ok := nm.Renderer.Offers[document_cid]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Render offer with CID '%v' does not exist.", document_cid))
	}

	// create the render offer object
	return offer, err

}

// Create the render offer document file from the offer object
func (offer *RenderOffer) AddDocument() error {
	var err error

	// // check if the render request was already submitted
	// if offer._isSubmitted() {
	// 	return errors.New(fmt.Sprintf("Render request was already submitted and cannot be modified."))
	// }

	// check if the document was already added
	if offer.DocumentCID != "" {
		return errors.New(fmt.Sprintf("Render offer document '%v' already exists.", offer.DocumentCID))
	}

	// Prepare the creation of a local render offer document file
	offer_document_filename := fmt.Sprintf("offer-%v-%v.json", strings.ReplaceAll(offer.Owner.String(), ".", "_"), offer.CreatedTimestamp.Unix())
	offer_document_directory := filepath.Join(GetAppDataPath(), RENDERHIVE_APP_DIRECTORY_LOCAL_OFFERS)
	offer.DocumentPath = filepath.Join(offer_document_directory, offer_document_filename)

	// if the request does NOT already exist
	if _, err := os.Stat(offer.DocumentPath); os.IsNotExist(err) {

		// create the directory
		err = os.MkdirAll(offer_document_directory, 0700)
		if err != nil && !os.IsExist(err) {
			return err
		}

		// create the local render offer document file
		offer_document_file, err := os.Create(offer.DocumentPath)
		if err != nil {
			return err
		}
		defer offer_document_file.Close()

		// write the render offer data into the file in JSON format
		encoder := json.NewEncoder(offer_document_file)
		encoder.SetIndent("", "  ")
		encoder.Encode(offer)

		// add the CID of the render offer document to the offer data
		offer.DocumentCID, err = ipfs.Manager.GetHashFromPath(offer.DocumentPath)
		if err != nil {
			return err
		}

	} else {
		return errors.New(fmt.Sprintf("Render offer document '%v' already exists.", offer.DocumentPath))
	}

	return err

}

// Deploy the render offer to IPFS via the local IPFS node
// NOTE:
// This makes the render offer document available to the IPFS network.
// Anyone, who knows the CID, the render offer document.
// However, the CID is not shared at this point with anyone.
func (offer *RenderOffer) Deploy() (string, error) {
	var err error

	// add the render request document to the file list
	err = offer.AddDocument()
	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not add render request document: %v", err))
	}

	// Upload the render request document file to IPFS
	offer.DocumentCID, err = ipfs.Manager.AddObjectFromPath(offer.DocumentPath, true)
	if err != nil {
		return "", err
	}

	// add the offer to the node's render offers
	Manager.Renderer.Offers[offer.DocumentCID] = offer

	return offer.DocumentCID, err

}

// Set the render price limit
func (ro *RenderOffer) SetPrice(price float64, currency string) error {
	var err error

	// Set the new price
	ro.Price = price

	return err

}

// Get the render price limit
func (ro *RenderOffer) GetPrice() float64 {

	return ro.Price

}

// Add a Blender version to the render offer
func (ro *RenderOffer) AddBlenderVersion(version string, engines *[]string, devices *[]string, threads uint8) error {
	var err error

	// log event
	logger.Manager.Package["node"].Debug().Msg("Start adding a Blender version for this node:")
	logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] Internal name: %v", version))
	logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] Engines: %v", strings.Join(*engines, ",")))
	logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] Devices: %v", strings.Join(*devices, ",")))

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

	// get the blender version from the map of valid Blender versions
	blender_bin, ok := RENDERHIVE_BLENDER_ARCHIVE_FILES[version]
	if !ok {
		return errors.New(fmt.Sprintf("Blender version '%v' is not supported by the Renderhive network.", version))
	}

	// check if the Blender binary is already available on the local file system
	blender_bin_path := filepath.Join(RENDERHIVE_APP_DIRECTORY_BLENDER_BINARIES, version+"-"+blender_bin.Linux.Commit, "blender")
	fmt.Println()
	if _, err := os.Stat(blender_bin_path); os.IsNotExist(err) {

		// log info event
		logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] Fetching Blender binary for version v%v from IPFS: %v (%v)", version, blender_bin.Linux.Filename, blender_bin.Linux.CID))

		// download the Blender binary from IPFS
		_, err := ipfs.Manager.GetObject(blender_bin.Linux.CID, filepath.Join(RENDERHIVE_APP_DIRECTORY_TEMP, blender_bin.Linux.Filename))
		if err != nil {
			return err
		}

		// TODO: unpack the .tar.xz and move the Blender binary to the correct location

	} else {

		// log info event
		logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] Blender binary for version v%v found: %v", version, blender_bin_path))

	}

	// create the BlenderAppData instance for the new version
	blender := BlenderAppData{
		Path:          blender_bin_path,
		Engines:       *engines,
		Devices:       *devices,
		Threads:       threads,
		BenchmarkTool: &BlenderBenchmarkTool{},
	}

	// log info event
	logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] Start Blender v%v and get app data.", version))

	// TODO: Does currently not work in Docker on Apple Silicon, because of Pixar/USD bug,
	//	     which makes it unable to run Blender (/proc/cpuinfo does not deliver the correct CPU info)
	// // start this Blender version and query its version and build info
	// err = blender.Execute([]string{"-v"})
	// if err != nil {
	// 	return err
	// }

	// // wait until the command execution finished
	// err = blender.Cmd.Wait()
	// if err != nil {
	// 	return err
	// }

	// append to the list of supported Blender versions
	ro.BlenderVersions = append(ro.BlenderVersions, RenderOfferBlenderVersions{
		Version: version,
		Engines: *engines,
		Devices: *devices,
		Threads: threads,
	})

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
func (ro *RenderOffer) DeleteBlenderVersion(version string) error {

	// log event
	logger.Manager.Package["node"].Trace().Msg("Remove a Blender version from the node's render offer:")

	// delete the element from the map, if it exists
	blender, ok := ro.Blender[version]
	if ok {

		// delete from the map
		delete(ro.Blender, version)

		// delete from the slice of supported versions
		for index, element := range ro.BlenderVersions {
			if element.Version == version {

				// make a new slice without the element
				ro.BlenderVersions = append(ro.BlenderVersions[:index], ro.BlenderVersions[index+1:]...)
				break

			}
		}

		// log debug event
		logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Version: %v", blender.BuildVersion))

		return nil

	}

	return errors.New(fmt.Sprintf(" [#] Blender v'%v' could not be removed from the node's render offer.", blender.BuildVersion))

}

// Submit the render request to the network
// NOTE:
// This announces the render offer to the renderhive network.
// From that point on, anyone can access the render offer document and expects the
// node to be ready for rendering.
func (offer *RenderOffer) Submit() (*hederasdk.TransactionReceipt, []byte, error) {
	var err error
	var transactionBytes []byte

	// // check if the render offer was already submitted
	// if offer._isSubmitted() {
	// 	return nil, nil, errors.New(fmt.Sprintf("Render offer was already submitted and cannot be modified."))
	// }

	// Submit the message to the render hive network
	// Prepare the HCS message
	jsonMessage, err := Manager.EncodeCommand(
		[]string{},
		SERVICE_NODE,
		METHOD_NODE_SUBMIT_RENDER_OFFER,
		&SubmitRenderOfferArgs{
			RenderOfferCID: offer.DocumentCID,
		},
	)

	// Encode the message as JSON
	if err != nil {
		return nil, nil, err
	} else {

		// send it to the Renderhive Job Queue topic on Hedera
		offer.Receipt, transactionBytes, err = Manager.JobQueueTopic.SubmitMessage(string(jsonMessage), "renderhive-v0.1.0::submit-render-offer", nil, hedera.TransactionOptions.SetExecute(false, Manager.User.UserAccount.AccountID))
		if err != nil {
			logger.Manager.Package["hedera"].Error().Err(err).Msg("")
			return nil, nil, errors.New(fmt.Sprintf("Render offer %v could not be submitted: %v.", nil, err.Error()))
		}

	}

	// TODO: Temporary workaround – Does not account for failed transactions
	// 		 We may safe the transactions ID and query the status later?
	// update the submitted timestamp
	offer._updateSubmittedTimestamp()

	return offer.Receipt, transactionBytes, err

}

// Pause the render offer
func (offer *RenderOffer) Pause() (*hederasdk.TransactionReceipt, []byte, error) {
	var err error
	var transactionBytes []byte
	var receipt *hederasdk.TransactionReceipt

	// Submit the render offer message to the job queue topic
	// Prepare the HCS message
	jsonMessage, err := Manager.EncodeCommand(
		[]string{},
		SERVICE_NODE,
		METHOD_NODE_PAUSE_RENDER_OFFER,
		&PauseRenderOfferArgs{
			RenderOfferCID: offer.DocumentCID,
		},
	)

	// Encode the message as JSON
	if err != nil {
		return nil, nil, err
	} else {

		// send it to the Renderhive Job Queue topic on Hedera
		receipt, transactionBytes, err = Manager.JobQueueTopic.SubmitMessage(string(jsonMessage), "renderhive-v0.1.0::pause-render-offer", nil, hedera.TransactionOptions.SetExecute(false, Manager.User.UserAccount.AccountID))
		if err != nil {
			logger.Manager.Package["hedera"].Error().Err(err).Msg("")
			return nil, nil, errors.New(fmt.Sprintf("Command could not be submitted: %v.", nil, err.Error()))
		}

	}

	// TODO: Temporary workaround – Does not account for failed transactions
	// update the cancelled status and closed timestamp
	offer._updatePausedTimestamp()

	return receipt, transactionBytes, err

}

// helper function to check if the offer was already successfully submitted
func (offer *RenderOffer) _isSubmitted() bool {

	// if the receipt is nil then it was not submitted
	if offer.Receipt == nil || offer.Receipt.Status != hederasdk.StatusSuccess {
		return false
	}

	return offer.Receipt.Status == hederasdk.StatusSuccess

}

// helper function to check if the offer was paused
func (offer *RenderOffer) _isPaused() bool {

	return offer.Paused

}

// helper function to update the modified timestamp of the offer
func (offer *RenderOffer) _updateModifiedTimestamp() {

	offer.ModifiedTimestamp = time.Now()

}

// helper function to update the submitted timestamp of the offer
func (offer *RenderOffer) _updateSubmittedTimestamp() {

	offer.SubmittedTimestamp = time.Now()

}

// helper function to update the activity status and the paused timestamp of the offer
func (offer *RenderOffer) _updatePausedTimestamp() {

	// if the was already submitted and not cancelled
	if offer._isSubmitted() && !offer._isPaused() {

		// update the closed timestamp and cancelled status
		offer.PausedTimestamp = time.Now()
		offer.Paused = true

	}

}

// RENDER REQUESTS
// #############################################################################
// Initialize the render requests for this node
func (nm *PackageManager) InitRenderRequests() error {
	var err error

	// initialize the node's render requests
	nm.Renderer.Requests = make(map[string]*RenderRequest)

	// load the render requests from the local file system
	err = nm.LoadRenderRequests()
	if err != nil {
		return errors.New(fmt.Sprintf("Could not load render requests: %v", err))
	}

	return err

}

// Load a render request into memory
func (nm *PackageManager) LoadRenderRequests() error {
	var err error

	// path to the local render request documents
	request_document_directory := filepath.Join(GetAppDataPath(), RENDERHIVE_APP_DIRECTORY_LOCAL_REQUESTS)

	// if the directory does NOT exist
	if _, err := os.Stat(request_document_directory); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Render request directory '%v' does not exist.", request_document_directory))
	}

	// go through all files in the directory
	err = filepath.Walk(request_document_directory, func(path string, info os.FileInfo, err error) error {

		// if the file is a regular file
		if info.Mode().IsRegular() {

			// if the file is a render request document
			if matched, _ := regexp.MatchString(`^request-.*\.json$`, info.Name()); matched {

				// load the render request from the file
				err = nm.LoadRenderRequestFromFile(path)

			}

		}

		// log error event
		if err != nil {
			logger.Manager.Package["node"].Error().Msg(fmt.Sprintf("Could not load render request %v: %v", path, err))
		}

		// reset error
		err = nil

		return nil

	})

	return err

}

// RENDER REQUEST OBJECTS
// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// Create a new render request object for this node
func (nm *PackageManager) NewRenderRequest(blender_version string, render_price float64) (*RenderRequest, error) {
	var err error

	// create the render request object
	return &RenderRequest{

		Files: make(map[string]files.Node),

		CreatedTimestamp:  time.Now(),
		ModifiedTimestamp: time.Now(),

		BlenderFile: BlenderFileData{},
		Version:     blender_version,
		Price:       render_price,
		ThisNode:    false,

		Owner: &Manager.User.UserAccount.AccountID,
	}, err

}

// Load a render request into memory
func (nm *PackageManager) LoadRenderRequestFromFile(path string) error {
	var err error

	// if the request does NOT exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Render request document '%v' does not exist.", path))
	}

	// get the CID of the render request document
	request_document_cid, err := ipfs.Manager.GetHashFromPath(path)
	if err != nil {
		return err
	}

	// load the render request document file
	request_document_file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer request_document_file.Close()

	// decode the render offer data from the file
	// NOTE: The *hedera.AccountID is not supported by the JSON decoder.
	//		 Therefore, the Owner field is decoded manually using this workaround.
	var request struct {
		RenderRequest
		Owner struct {
			Shard   uint64 `json:"Shard"`
			Realm   uint64 `json:"Realm"`
			Account uint64 `json:"Account"`

			// TODO: Currently, the AliasKey and AliasEvmAddress are ignored
			// AliasKey        hederasdk.PublicKey `json:"AliasKey"`
			// AliasEvmAddress []byte              `json:"AliasEvmAddress"`
		} `json:"Owner"`
	}
	decoder := json.NewDecoder(request_document_file)
	err = decoder.Decode(&request)
	if err != nil {
		return err
	}

	// create the render offer object
	nm.Renderer.Requests[request_document_cid] = &RenderRequest{
		DocumentCID:       request_document_cid,
		DocumentPath:      path,
		DirectoryCID:      request.DirectoryCID,
		CreatedTimestamp:  request.CreatedTimestamp,
		ModifiedTimestamp: request.ModifiedTimestamp,
		BlenderFile:       request.BlenderFile,
		Version:           request.Version,
		Price:             request.Price,
		ThisNode:          request.ThisNode,
		Owner: &hederasdk.AccountID{
			Shard:   request.Owner.Shard,
			Realm:   request.Owner.Realm,
			Account: request.Owner.Account,
		},
		Receipt: request.Receipt,
	}

	return nil
}

// Get the render request object from a CID
func (nm *PackageManager) GetRenderRequest(document_cid string) (*RenderRequest, error) {
	var err error

	// Get the Render Request from the CID of the render request document
	request, ok := nm.Renderer.Requests[document_cid]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Render request with CID '%v' does not exist.", document_cid))
	}

	// create the render request object
	return request, err

}

// Add a local file to the render request
func (request *RenderRequest) AddFile(path string, filename string) error {
	var err error
	var stat os.FileInfo

	// check if the render request was already submitted
	if request._isSubmitted() {
		return errors.New(fmt.Sprintf("Render request was already submitted and cannot be modified."))
	}

	// check if the file exists
	if stat, err = os.Stat(path); os.IsNotExist(err) {
		return err
	}

	// create a new serial file from the file path
	file, err := files.NewSerialFile(path, false, stat)
	if err != nil {
		return err
	}

	// add the file to the list of files
	request.Files[filename] = file

	// update the modified timestamp
	request._updateModifiedTimestamp()

	return err

}

// Add a file from zje file data to the render request
func (request *RenderRequest) AddFileFromBytes(filename string, data []byte) error {
	var err error

	// check if the render request was already submitted
	if request._isSubmitted() {
		return errors.New(fmt.Sprintf("Render request was already submitted and cannot be modified."))
	}

	// Create a File from a byte array of file data
	file := files.NewBytesFile(data)
	if err != nil {
		return err
	}

	// add the file to the list of files
	request.Files[filename] = file

	// update the modified timestamp
	request._updateModifiedTimestamp()

	return err

}

// Remove a file from the render request
func (request *RenderRequest) RemoveFile(filename string) error {
	var err error

	// check if the render request was already submitted
	if request._isSubmitted() {
		return errors.New(fmt.Sprintf("Render request was already submitted and cannot be modified."))
	}

	// delete the element from the map, if it exists
	_, ok := request.Files[filename]
	if ok {
		delete(request.Files, filename)
	} else {
		err = errors.New(fmt.Sprintf("File '%v' could not be removed from the render request.", filename))
	}

	// update the modified timestamp
	request._updateModifiedTimestamp()

	return err
}

// Add the directory mapping from the files to the render request
func (request *RenderRequest) MakeDirectory(overwrite bool) error {
	var err error

	// check if the render request was already submitted
	if request._isSubmitted() {
		return errors.New(fmt.Sprintf("Render request was already submitted and cannot be modified."))
	}

	// if there is no directory yet or it should be overwritten
	if request.Directory == nil || overwrite == true {

		// if there are NO files
		if len(request.Files) == 0 {
			return errors.New(fmt.Sprintf("No files were added to the render request."))
		}

		// create a new directory from the files
		request.Directory = files.NewMapDirectory(request.Files)

	} else {

		err = errors.New(fmt.Sprintf("Directory already exists."))

	}

	// update the modified timestamp
	request._updateModifiedTimestamp()

	return err

}

// Remove the directory mapping from the files to the render request
func (request *RenderRequest) RemoveDirectory() error {
	var err error

	// check if the render request was already submitted
	if request._isSubmitted() {
		return errors.New(fmt.Sprintf("Render request was already submitted and cannot be modified."))
	}

	// TODO: Implement the removal of the directory from the request
	// check if the directory was already uploaded
	if request.DirectoryCID != "" {
		return errors.New(fmt.Sprintf("Directory was already uploaded and cannot be removed."))
	}

	// if there is a directory
	if request.Directory != nil {

		// remove the directory
		request.Directory = nil

	} else {

		err = errors.New(fmt.Sprintf("Directory does not exist."))
	}

	// update the modified timestamp
	request._updateModifiedTimestamp()

	return err

}

// Create the render request document file from the request object and add it to the request directory
func (request *RenderRequest) AddDocument() error {
	var err error

	// check if the render request was already submitted
	if request._isSubmitted() {
		return errors.New(fmt.Sprintf("Render request was already submitted and cannot be modified."))
	}

	// check if the document was already added
	if request.DocumentCID != "" {
		return errors.New(fmt.Sprintf("Render request document '%v' already exists.", request.DocumentCID))
	}

	// Prepare the creation of a local render request document file
	request_document_filename := fmt.Sprintf("request-%v-%v.json", strings.ReplaceAll(request.Owner.String(), ".", "_"), request.CreatedTimestamp.Unix())
	request_document_directory := filepath.Join(GetAppDataPath(), RENDERHIVE_APP_DIRECTORY_LOCAL_REQUESTS)
	request.DocumentPath = filepath.Join(request_document_directory, request_document_filename)

	// if the directory does NOT exist
	if _, err := os.Stat(request.DocumentPath); os.IsNotExist(err) {

		// create the directory
		err = os.MkdirAll(request_document_directory, 0700)
		if err != nil && !os.IsExist(err) {
			return err
		}

		// create the local render request document file
		request_document_file, err := os.Create(request.DocumentPath)
		if err != nil {
			return err
		}
		defer request_document_file.Close()

		// write the render request data into the file in JSON format
		encoder := json.NewEncoder(request_document_file)
		encoder.SetIndent("", "  ")
		encoder.Encode(request)

		// add the CID of the render request document to the request data
		request.DocumentCID, err = ipfs.Manager.GetHashFromPath(request.DocumentPath)
		if err != nil {
			return err
		}

	} else {
		return errors.New(fmt.Sprintf("Render request document '%v' already exists.", request.DocumentPath))
	}

	return err

}

// Deploy the render request directory to IPFS via the local IPFS node
// NOTE:
// This makes the render request document and all files available to the IPFS network.
// Anyone, who knows the CID, can access the files and the render request document.
// However, the CID is not shared at this point with anyone.
func (request *RenderRequest) Deploy() (string, error) {
	var err error

	// check if the render request was already submitted
	if request._isSubmitted() {
		return "", errors.New(fmt.Sprintf("Render request was already submitted and cannot be modified."))
	}

	// make the render request directory
	err = request.MakeDirectory(false)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not create render request directory: %v", err))
	}

	// Upload the request directory to the local IPFS node
	request.DirectoryCID, err = ipfs.Manager.AddObject(request.Directory, true)
	if err != nil {
		return "", err
	}

	// close all files and free the memory
	for _, file := range request.Files {
		file.Close()
	}

	// add the render request document to the file list
	err = request.AddDocument()
	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not add render request document: %v", err))
	}

	// Upload the render request document file to IPFS
	request.DocumentCID, err = ipfs.Manager.AddObjectFromPath(request.DocumentPath, true)
	if err != nil {
		return "", err
	}

	// add the request to the node's render requests
	Manager.Renderer.Requests[request.DocumentCID] = request

	return request.DocumentCID, err

}

// Submit the render request to the network
// NOTE:
// This announces the render request to the renderhive network.
// From that point on, anyone can access the render request document and the Blender files.
// Later we might add some encryption here to protect the data and making it only
// accessible to those render nodes, which shall render the file. However, this would
// require that the user submitting the render request needs to stay online until the
// job distribution is done.
func (request *RenderRequest) Submit() (*hederasdk.TransactionReceipt, []byte, error) {
	var err error
	var transactionBytes []byte

	// check if the render request was already submitted
	if request._isSubmitted() {
		return nil, nil, errors.New(fmt.Sprintf("Render request was already submitted and cannot be modified."))
	}

	// Submit the message to the render hive network
	// Prepare the HCS message
	jsonMessage, err := Manager.EncodeCommand(
		[]string{},
		SERVICE_NODE,
		METHOD_NODE_SUBMIT_RENDER_REQUEST,
		&SubmitRenderRequestArgs{
			RenderRequestCID: request.DocumentCID,
			BlenderFileCID:   request.BlenderFile.CID,
		},
	)

	// Encode the message as JSON
	if err != nil {
		return nil, nil, err
	} else {

		// send it to the Renderhive Job Queue topic on Hedera
		request.Receipt, transactionBytes, err = Manager.JobQueueTopic.SubmitMessage(string(jsonMessage), "renderhive-v0.1.0::submit-render-request", nil, hedera.TransactionOptions.SetExecute(false, Manager.User.UserAccount.AccountID))
		if err != nil {
			logger.Manager.Package["hedera"].Error().Err(err).Msg("")
			return nil, nil, errors.New(fmt.Sprintf("Render request %v could not be submitted: %v.", nil, err.Error()))
		}

	}

	// TODO: Temporary workaround – Does not account for failed transactions
	// update the submitted timestamp
	request._updateSubmittedTimestamp()

	return request.Receipt, transactionBytes, err

}

// Cancel the render request
func (request *RenderRequest) Cancel() (*hederasdk.TransactionReceipt, []byte, error) {
	var err error
	var transactionBytes []byte
	var receipt *hederasdk.TransactionReceipt

	// Submit the render request message to the job queue topic
	// Prepare the HCS message
	jsonMessage, err := Manager.EncodeCommand(
		[]string{},
		SERVICE_NODE,
		METHOD_NODE_CANCEL_RENDER_REQUEST,
		&CancelRenderRequestArgs{
			RenderRequestCID: request.DocumentCID,
		},
	)

	// Encode the message as JSON
	if err != nil {
		return nil, nil, err
	} else {

		// send it to the Renderhive Job Queue topic on Hedera
		receipt, transactionBytes, err = Manager.JobQueueTopic.SubmitMessage(string(jsonMessage), "renderhive-v0.1.0::cancel-render-request", nil, hedera.TransactionOptions.SetExecute(false, Manager.User.UserAccount.AccountID))
		if err != nil {
			logger.Manager.Package["hedera"].Error().Err(err).Msg("")
			return nil, nil, errors.New(fmt.Sprintf("Command could not be submitted: %v.", nil, err.Error()))
		}

	}

	// if the transaction was successful
	if receipt.Status == hederasdk.StatusSuccess {

		// update the cancelled status and closed timestamp
		request._updateCancelledTimestamp()

	}

	return receipt, transactionBytes, err

}

// helper function to check if the request was already successfully submitted
func (request *RenderRequest) _isSubmitted() bool {

	// if the receipt is nil then it was not submitted
	if request.Receipt == nil || request.Receipt.Status != hederasdk.StatusSuccess {
		return false
	}

	return request.Receipt.Status == hederasdk.StatusSuccess

}

// helper function to check if the request was cancelled
func (request *RenderRequest) _isCancelled() bool {

	return request.Cancelled

}

// helper function to update the modified timestamp of the request
func (request *RenderRequest) _updateModifiedTimestamp() {

	request.ModifiedTimestamp = time.Now()

}

// helper function to update the submitted timestamp of the request
func (request *RenderRequest) _updateSubmittedTimestamp() {

	request.SubmittedTimestamp = time.Now()

}

// helper function to update the cancelled status and the closed timestamp of the request
func (request *RenderRequest) _updateCancelledTimestamp() {

	// if the was already submitted and not cancelled
	if request._isSubmitted() && !request._isCancelled() {

		// update the closed timestamp and cancelled status
		request.ClosedTimestamp = time.Now()
		request.Cancelled = true

	}

}

// RENDER REQUEST ON THE NODE
// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

// Add a new render request for this node
func (nm *PackageManager) AddRenderRequest(request *RenderRequest, overwrite bool) (int, error) {
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
		ids := make([]string, 0, len(nm.Renderer.Requests))
		for id := range nm.Renderer.Requests {
			ids = append(ids, id)
		}

		// sort the slice of IDs
		sort.Strings(ids)

		// Get the ID assigned to the last render request and add 1 for the new ID
		if len(ids) > 0 {
			newID = len(ids) + 1
		}

	} else {

		// initialized the map first
		nm.Renderer.Requests = make(map[string]*RenderRequest)

	}

	// Update the ID
	request.ID = newID

	// Append the request to the list of requests of this node
	nm.Renderer.Requests[strconv.Itoa(newID)] = request

	// Add the CID of the Blender file to the render request data
	request.BlenderFile.CID, err = ipfs.Manager.GetHashFromPath(request.BlenderFile.Path)
	if err != nil {
		return 0, err
	}

	// Prepare the creation of a local render request document file
	request_document_filename := fmt.Sprintf("request-%v.json", request.BlenderFile.CID)
	request_document_directory := filepath.Join(GetAppDataPath(), RENDERHIVE_APP_DIRECTORY_LOCAL_REQUESTS)
	request.DocumentPath = filepath.Join(request_document_directory, request_document_filename)
	if _, err := os.Stat(request.DocumentPath); os.IsNotExist(err) || overwrite {

		// create the directory
		err = os.MkdirAll(request_document_directory, 0700)
		if err != nil && !os.IsExist(err) {
			return 0, err
		}

		// create the local render request document file
		request_document_file, err := os.Create(request.DocumentPath)
		if err != nil {
			return 0, err
		}
		defer request_document_file.Close()

		// write the render request data into the file in JSON format
		encoder := json.NewEncoder(request_document_file)
		encoder.Encode(request)

	} else {
		return 0, errors.New(fmt.Sprintf("Render request document '%v' already exists.", request.DocumentPath))
	}

	return newID, err

}

// Remove a render request from the node
func (nm *PackageManager) RemoveRenderRequest(id int) error {
	var err error

	// log event
	logger.Manager.Package["node"].Trace().Msg("Removing a render request from the node:")

	// delete the element from the map, if it exists
	request, ok := nm.Renderer.Requests[strconv.Itoa(id)]
	if ok {
		logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] ID: %v", request.ID))
		delete(nm.Renderer.Requests, strconv.Itoa(id))
	} else {
		err = errors.New(fmt.Sprintf("Render request %v could not be removed from the node.", id))
	}

	return err

}

// Submit a render request from this node to the render hive network
func (nm *PackageManager) SubmitRenderRequest(id int) error {
	var err error
	var request *RenderRequest

	// log trace event
	logger.Manager.Package["node"].Trace().Msg("Submitting a render request for this node to the render hive:")

	// if the render request exists
	request, ok := nm.Renderer.Requests[strconv.Itoa(id)]
	if ok {

		// log trace event
		logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] ID: %v", request.ID))

		// Put the Blender file on the local IPFS node
		request.BlenderFile.CID, err = ipfs.Manager.AddObjectFromPath(request.BlenderFile.Path, true)
		if err != nil {
			return err
		}

		// log trace event
		logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Blender File (CID): %v", request.BlenderFile.CID))

		// TODO: Call the smart contract and add the transaction hash to the
		//       render request document

		// Put the render request document on the local IPFS node
		request.DocumentCID, err = ipfs.Manager.AddObjectFromPath(request.DocumentPath, true)
		if err != nil {
			return err
		}

		// log trace event
		logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Render Request Document (CID): %v", request.DocumentCID))

		// Submit the render request message to the job queue topic
		// Prepare the HCS message
		message := RenderRequestMessage{
			DocumentCID:    request.DocumentCID,
			BlenderFileCID: request.BlenderFile.CID,
		}

		// Encode the message as JSON
		jsonMessage, err := json.Marshal(message)
		if err != nil {
			return err
		} else {

			// send it to the Renderhive Job Queue topic on Hedera
			request.Receipt, _, err = nm.JobQueueTopic.SubmitMessage(string(jsonMessage), "renderhive-v0.1.0::submit-render-request", nil)
			if err != nil {
				logger.Manager.Package["hedera"].Error().Err(err).Msg("")
				return errors.New(fmt.Sprintf("Render request %v could not be submitted: %v.", id, err.Error()))
			}
			if request.Receipt != nil {
				logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] [*] Receipt: %s (Status: %s)", request.Receipt.TransactionID.String(), request.Receipt.Status))
				if !strings.EqualFold(request.Receipt.Status.String(), "SUCCESS") {
					err = errors.New(fmt.Sprintf("Render request %v could not be submitted to Hedera: Receipt status '%v'.", id, request.Receipt.Status.String()))
					return err
				}
			}

		}

	} else {
		err = errors.New(fmt.Sprintf("Render request could not be submitted: Request ID %v does not exist.", id))
		return err
	}

	return err

}

// RENDER QUEUE
// #############################################################################
// Message callback to receive the job queue data from the render hive
func (nm *PackageManager) JobQueueMessageCallback() func(message hederasdk.TopicMessage) {

	return func(message hederasdk.TopicMessage) {
		var err error

		// decode the received command
		command, err := nm.DecodeCommand(message.Contents)
		if err != nil {
			logger.Manager.Package["hedera"].Error().Msg(fmt.Sprintf("Failed to process received command: %s", string(message.Contents)))
			return
		}

		// decode rpc call from base64 to JSON
		jsonMessage := make([]byte, base64.StdEncoding.DecodedLen(len(command.Message)))
		n, err := base64.StdEncoding.Decode(jsonMessage, command.Message)
		if err != nil {
			logger.Manager.Package["hedera"].Error().Msg(fmt.Sprintf("Failed to decode base64 encoded JSON-RPC message: %s", string(command.Message)))
			return
		}

		// reduce to the number of bytes actually written
		jsonMessage = jsonMessage[:n]

		// unmarshal the JSON message into a JsonRpcMessage
		var rpcMessage JsonRpcMessage
		err = json.Unmarshal(jsonMessage, &rpcMessage)
		if err != nil {
			logger.Manager.Package["hedera"].Error().Msg(fmt.Sprintf("Failed to retrieve JSON-RPC message (%s): %v", string(jsonMessage), err))
			return
		}

		// Convert Params to JSON
		params, err := json.Marshal(rpcMessage.Params)
		if err != nil {
			logger.Manager.Package["hedera"].Error().Msg(fmt.Sprintf("Failed to retrieve JSON-RPC parameters (%s): %v", rpcMessage.Params, err))
			return
		}

		// get the service and method types
		service, method, err := nm.GetServiceAndMethodInt(rpcMessage.Method)
		if service == SERVICE_UNKNOWN && method == METHOD_UNKNOWN {
			logger.Manager.Package["hedera"].Error().Msg(fmt.Sprintf("Unknown JSON-RPC method (%s): %v", rpcMessage.Method, err))
			return
		}

		// TODO: Verify that the message is valid.
		// ...

		// Process the message according to the service and method types
		if service == SERVICE_NODE && method == METHOD_NODE_SUBMIT_RENDER_REQUEST {
			// Unmarshal Params into SubmitRenderRequestArgs
			var request SubmitRenderRequestArgs
			err = json.Unmarshal(params, &request)
			if err != nil {
				logger.Manager.Package["hedera"].Error().Msg(fmt.Sprintf("Message received but not processed: %s", string(message.Contents)))
				return
			}

			// Pin the render request document and blender file to the local IPFS node
			// TODO: Add a proper file management. Downloading each file, probably is
			//       too resource intensive at larger network scales.
			go ipfs.Manager.PinObject(request.RenderRequestCID)
			go ipfs.Manager.PinObject(request.BlenderFileCID)

			// create the RenderJob element for the internal job management
			job := &RenderJob{
				Request: &RenderRequest{
					DocumentCID:        request.RenderRequestCID,
					SubmittedTimestamp: message.ConsensusTimestamp,
				},
			}

			// add the request to the slice of render jobs for the internal job management
			nm.NetworkQueue = append(nm.NetworkQueue, job)

			// log trace event
			logger.Manager.Package["node"].Debug().Msg("Received a new render request:")
			logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] Render request document: %v", job.Request.DocumentCID))
			logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] Submitted: %v", job.Request.SubmittedTimestamp))

		} else if service == SERVICE_NODE && method == METHOD_NODE_CANCEL_RENDER_REQUEST {

			// TODO: Implement the cancellation of a render request

		} else if service == SERVICE_NODE && method == METHOD_NODE_SUBMIT_RENDER_OFFER {

			// Unmarshal Params into SubmitRenderOfferArgs
			var offer SubmitRenderOfferArgs
			err = json.Unmarshal(params, &offer)
			if err != nil {
				logger.Manager.Package["hedera"].Error().Msg(fmt.Sprintf("Message received but not processed: %s", string(message.Contents)))
				return
			}

			// Pin the render offer document to the local IPFS node
			go ipfs.Manager.PinObject(offer.RenderOfferCID)

			// create the RenderOffer element for the internal job management
			ro := &RenderOffer{
				DocumentCID:        offer.RenderOfferCID,
				SubmittedTimestamp: message.ConsensusTimestamp,
			}

			// log trace event
			logger.Manager.Package["node"].Debug().Msg("Received a new render offer:")
			logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] Render offer document: %v", ro.DocumentCID))
			logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] Submitted: %v", ro.SubmittedTimestamp))

		}

	}

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
		fmt.Println("Error:", err)
		return "", err
	}

	return string(output), err

}

// Run the Blender benchmark tool with the specified Blender version and
// rendering device
func (tool *BlenderBenchmarkTool) Run(ro *RenderOffer, benchmark_version string, benchmark_device string, benchmark_scene string) error {
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

		// Prepare the creation of a local benchmark result file
		benchmark_result_filename := fmt.Sprintf("benchmark-result-%v.json", benchmark_version)
		benchmark_result_directory := filepath.Join(GetAppDataPath(), RENDERHIVE_APP_DIRECTORY_BLENDER_BENCHMARKS)
		benchmark_result_path := filepath.Join(benchmark_result_directory, benchmark_result_filename)

		// create the directory
		err = os.MkdirAll(benchmark_result_directory, 0700)
		if err != nil && !os.IsExist(err) {
			return err
		}

		// create the benchmark result file
		benchmar_result_file, err := os.Create(benchmark_result_path)
		if err != nil {
			return err
		}
		defer benchmar_result_file.Close()

		// write the render request data into the file in JSON format
		encoder := json.NewEncoder(benchmar_result_file)
		encoder.Encode(tool.Result)

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
func (b *BlenderAppData) Execute(args []string) error {
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
func (b *BlenderAppData) ProcessOutput(name string, output io.ReadCloser) error {
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
func (nm *PackageManager) CreateCommandRequest() *cobra.Command {

	// flags for the 'request' command
	var list bool

	// create a 'blender' command for the node
	command := &cobra.Command{
		Use:   "request",
		Short: "Manage the node's render requests",
		Long:  "This command is for adding/removing/editing the render requests of this node.",
		Run: func(cmd *cobra.Command, args []string) {

			// if a render offer exists
			if nm.Renderer.Requests != nil && len(nm.Renderer.Requests) > 0 {

				// list all Blender versions
				if list {

					fmt.Println("")
					fmt.Println("The node has the following render requests:")

					// find each Blender version added to the node
					for _, request := range nm.Renderer.Requests {
						fmt.Printf(" [#] ID: %v for Blender file '%v' (Submitted: %v) \n", request.ID, request.BlenderFile.Path, (request.Receipt != nil))
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
	command.AddCommand(nm.CreateCommandRequest_Submit())
	// command.AddCommand(nm.CreateCommandRequest_Pause())
	// command.AddCommand(nm.CreateCommandRequest_Revoke())

	return command

}

// Create the CLI command to add a new render request for this node
func (nm *PackageManager) CreateCommandRequest_Add() *cobra.Command {

	// flags for the 'request add' command
	var blender_version string
	var blender_file string
	var render_price float64
	var this_node bool

	// create a 'request add' command for the node
	command := &cobra.Command{
		Use:   "add",
		Short: "Add a Blender version to the node's render offer",
		Long:  "This command is for adding a Blender version to the node's render offer.",
		Run: func(cmd *cobra.Command, args []string) {

			// if the map was not correctly initialized
			if nm.Renderer.Requests != nil {

				// add a Blender version
				if blender_version != "" && blender_file != "" && render_price > 0 {
					fmt.Println("")

					// Check if path is pointing to an existing blender file
					fileInfo, err := os.Stat(blender_file)
					if !os.IsNotExist(err) {

						if !fileInfo.Mode().IsRegular() || !strings.HasSuffix(fileInfo.Name(), ".blend") {
							fmt.Println(fmt.Errorf("The given path '%v' is not pointing to a regular '.blend' file. \n", blender_file))
							return
						}

					} else {

						fmt.Println(fmt.Errorf(err.Error()))
						fmt.Println("")
						return

					}

					// Create a new render request
					request := &RenderRequest{

						CreatedTimestamp:  time.Now(),
						ModifiedTimestamp: time.Now(),

						BlenderFile: BlenderFileData{Path: blender_file},
						Version:     blender_version,
						Price:       render_price,
						ThisNode:    this_node,
					}

					// Add the render request to the node
					id, err := nm.AddRenderRequest(request, true)

					if err != nil {
						fmt.Println(err)
					} else {

						fmt.Println("Added a new render request to the node:")
						fmt.Printf(" [#] ID: %v\n", id)
						fmt.Printf(" [#] Blender file: %v\n", blender_file)
						fmt.Printf(" [#] Blender file CID: %v\n", request.BlenderFile.CID)
						fmt.Printf(" [#] Requested Blender version: %v\n", blender_version)
						fmt.Printf(" [#] Maximum price: %v USD / BBP \n", render_price)
						fmt.Printf(" [#] Node participates: %v \n", this_node)

					}
					fmt.Println("")

				} else {

					fmt.Println("")
					fmt.Println(fmt.Errorf("Failed to create the render request."))
					if blender_version == "" {
						fmt.Println(fmt.Errorf(" [#] Missing a required parameter: Blender version (--blender-version)."))
					}
					if blender_file == "" {
						fmt.Println(fmt.Errorf(" [#] Missing a required parameter: Blender file (--blender-file)."))
					}
					if render_price == 0 {
						fmt.Println(fmt.Errorf(" [#] Missing a required parameter: Maximum render price (--render-price)."))
					}
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
func (nm *PackageManager) CreateCommandRequest_Remove() *cobra.Command {

	// flags for the 'request remove' command
	var id int

	// create a 'request remove' command for the node
	command := &cobra.Command{
		Use:   "remove",
		Short: "Remove a render request from this node",
		Long:  "This command is for removing a render request from this node. In case it was submitted to the network, it will be cancelled and revoked.",
		Run: func(cmd *cobra.Command, args []string) {

			// if render requests exists
			if nm.Renderer.Requests != nil {

				// was a valid ID passed?
				if id != -1 {

					// if the parsed version is supported by the node
					_, ok := nm.Renderer.Requests[strconv.Itoa(id)]
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
					if id == -1 {
						fmt.Println(fmt.Errorf(" [#] Missing a required parameter: Request ID (--request-id)."))
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

	// add command flag parameters
	command.Flags().IntVarP(&id, "request-id", "i", -1, "The ID of the render request to remove")

	return command

}

// Create the CLI command to remove a previously created render request from this node
func (nm *PackageManager) CreateCommandRequest_Submit() *cobra.Command {

	// flags for the 'submit remove' command
	var id int

	// create a 'request submit' command for the node
	command := &cobra.Command{
		Use:   "submit",
		Short: "Submit a render request from this node to the render hve",
		Long:  "This command is for submitting a render request from this node to the render hive network for rendering. Several transactions will be performed to inform the smart contract and to announce the new render request on the HCS topics.",
		Run: func(cmd *cobra.Command, args []string) {

			// if render requests exists
			if nm.Renderer.Requests != nil {

				// was a valid ID passed?
				if id != -1 {

					// if the parsed version is supported by the node
					request, ok := nm.Renderer.Requests[strconv.Itoa(id)]
					if ok {

						// Submit the render request
						err := nm.SubmitRenderRequest(id)
						if err != nil {

							fmt.Println("")
							fmt.Println(fmt.Errorf(err.Error()))
							fmt.Println("")

							return

						}

						fmt.Println("")
						fmt.Printf("Submitted render request with ID %v to the render hive. \n", id)
						fmt.Printf(" [#] Blender file (CID): %v. \n", request.BlenderFile.CID)
						fmt.Printf(" [#] Render request document (CID): %v. \n", request.DocumentCID)
						fmt.Println("")

					} else {

						fmt.Println("")
						fmt.Println(fmt.Errorf("There is no render request with ID %v.", id))
						fmt.Println("")

					}

				} else {

					fmt.Println("")
					fmt.Println(fmt.Errorf("Failed to submit the render request."))
					if id == -1 {
						fmt.Println(fmt.Errorf(" [#] Missing a required parameter: Request ID (--request-id)."))
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

	// add command flag parameters
	command.Flags().IntVarP(&id, "request-id", "i", -1, "The ID of the render request to remove")

	return command

}

// COMMAND LINE INTERFACE - BLENDER AND RENDERING
// #############################################################################
// Create the CLI command to control Blender from the Render Service App
func (nm *PackageManager) CreateCommandBlender() *cobra.Command {

	// flags for the 'blender' command
	var list bool

	// create a 'blender' command for the node
	command := &cobra.Command{
		Use:   "blender",
		Short: "Manage the node's Blender versions",
		Long:  "This command is for adding/removing Blender versions from the node's render offer. You can also start Blender in the background mode for rendering purposes.",
		Run: func(cmd *cobra.Command, args []string) {

			// if a render offer exists
			if nm.Renderer.ActiveOffer != nil {

				// list all Blender versions
				if list {

					fmt.Println("")
					fmt.Println("The node offers the following Blender versions for rendering:")

					// find each Blender version added to the node
					for _, blender := range nm.Renderer.ActiveOffer.Blender {
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
func (nm *PackageManager) CreateCommandBlender_Add() *cobra.Command {

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
		Long:  "This command is for adding a Blender version to the node's render offer.",
		Run: func(cmd *cobra.Command, args []string) {

			// if a render offer exists
			if nm.Renderer.ActiveOffer != nil {

				// add a Blender version
				if len(version) != 0 {
					fmt.Println("")

					// Check if path is pointing to an existing file
					if _, err := os.Stat(path); os.IsNotExist(err) {
						fmt.Println(fmt.Errorf("The given path '%v' is not a valid path.", path))
						return
					}

					// Add a new Blender version to the node's render offer
					err := nm.Renderer.ActiveOffer.AddBlenderVersion(version, &engines, &devices, threads)
					if err != nil {
						fmt.Println("")
						fmt.Println(err)
					} else {

						fmt.Printf("Added the Blender v'%v' with path '%v' to the render offer. \n", nm.Renderer.ActiveOffer.Blender[version].BuildVersion, path)

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
func (nm *PackageManager) CreateCommandBlender_Remove() *cobra.Command {

	// flags for the 'blender remove' command
	var version string

	// create a 'blender remove' command for the node
	command := &cobra.Command{
		Use:   "remove",
		Short: "Remove a Blender version from the node's render offer",
		Long:  "This command is for removing a Blender version from the node's render offer.",
		Run: func(cmd *cobra.Command, args []string) {

			// if a render offer exists
			if nm.Renderer.ActiveOffer != nil {

				// remove the Blender version
				if len(version) != 0 {

					// if the parsed version is supported by the node
					_, ok := nm.Renderer.ActiveOffer.Blender[version]
					if ok {
						fmt.Println("")
						fmt.Printf("Removing Blender v%v from the render offer of this node. \n", version)
						fmt.Println("")

						// Delete the Blender version from the node's render offer
						nm.Renderer.ActiveOffer.DeleteBlenderVersion(version)

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
func (nm *PackageManager) CreateCommandBlender_Run() *cobra.Command {

	// flags for the 'blender remove' command
	var version string
	var param string

	// create a 'blender remove' command for the node
	command := &cobra.Command{
		Use:   "run",
		Short: "Run a Blender version from the node's render offer",
		Long:  "This command is for starting a particular Blender version, which is in the node's render offer.",
		Run: func(cmd *cobra.Command, args []string) {

			// if a render offer exists
			if nm.Renderer.ActiveOffer != nil {

				// if a version was parsed and
				if len(version) != 0 {

					// if the parsed version is supported by this node
					if blender, ok := nm.Renderer.ActiveOffer.Blender[version]; ok {
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
func (nm *PackageManager) CreateCommandBlender_Benchmark() *cobra.Command {

	// flags for the 'blender benchmark' command
	var version string
	var use_tool bool
	var scene string
	var device string

	// create a 'blender remove' command for the node
	command := &cobra.Command{
		Use:   "benchmark",
		Short: "Run a Blender benchmark",
		Long:  "This command is for starting a benchmark rendering for a particular Blender version supported by this node.",
		Run: func(cmd *cobra.Command, args []string) {

			// if a render offer exists
			if nm.Renderer.ActiveOffer != nil {

				// if a version was parsed and
				if len(version) != 0 {

					// if the parsed version is supported by this node
					if blender, ok := nm.Renderer.ActiveOffer.Blender[version]; ok {

						// if the official Blender benchmark tool shall be used
						if use_tool {

							// run the this Blender version
							err := blender.BenchmarkTool.Run(nm.Renderer.ActiveOffer, version, device, scene)
							if err != nil {
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

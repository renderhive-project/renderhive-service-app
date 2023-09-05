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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	// external
	gocid "github.com/ipfs/go-cid"
	"github.com/ipfs/go-libipfs/files"
	icore "github.com/ipfs/interface-go-ipfs-core"
	ioptions "github.com/ipfs/interface-go-ipfs-core/options"
	icorepath "github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/ipfs/kubo/commands"
	"github.com/ipfs/kubo/config"
	"github.com/ipfs/kubo/core"
	"github.com/ipfs/kubo/core/coreapi"
	"github.com/ipfs/kubo/core/corehttp"
	"github.com/ipfs/kubo/core/node/libp2p"
	"github.com/ipfs/kubo/plugin/loader"
	"github.com/ipfs/kubo/repo"
	"github.com/ipfs/kubo/repo/fsrepo"
	process "github.com/jbenet/goprocess"
	peer "github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"

	// internal
	. "renderhive/globals"
	"renderhive/logger"
	. "renderhive/utility"
)

// structure for the IPFS manager
type PackageManager struct {

	// Local IPFS node
	IpfsContext       context.Context
	IpfsContextCancel func()
	IpfsRepoPath      string
	IpfsRepo          repo.Repo
	IpfsNode          *core.IpfsNode
	IpfsAPI           icore.CoreAPI

	// w3up service
	W3Agent w3cliAgent

	// Command line interface
	Command      *cobra.Command
	CommandFlags struct {
		FlagPlaceholder bool
	}
}

// IPFS MANAGER
// #############################################################################
// create the ipfs manager variable
var Manager = PackageManager{}

// Initialize everything required for the IPFS management
func (ipfsm *PackageManager) Init() error {
	var err error

	// log information
	logger.Manager.Package["ipfs"].Info().Msg("Initializing the IPFS manager ...")

	// Create the local IPFS node
	_, err = ipfsm.StartLocalNode()
	if err != nil {
		logger.Manager.Package["ipfs"].Error().Msg(err.Error())
	}

	// Initialize w3 CLI command
	ipfsm.W3Agent.Path = "w3"

	// Get the DID of this agent
	_, err = ipfsm.W3Agent.Whoami()
	if err != nil {
		logger.Manager.Package["ipfs"].Error().Msg(err.Error())
		return err
	}

	// log did key of the w3up agent
	logger.Manager.Package["ipfs"].Info().Msg(fmt.Sprintf(" [#] w3up identity: %v", ipfsm.W3Agent.DIDkey))

	// Get list of spaces this agent has access to
	err = ipfsm.W3Agent.SpaceList()
	if err != nil {
		logger.Manager.Package["ipfs"].Error().Msg(err.Error())
		return err
	}

	// log number of known w3up spaces
	logger.Manager.Package["ipfs"].Info().Msg(fmt.Sprintf(" [#] w3up spaces: %v", len(ipfsm.W3Agent.Spaces)))

	// if there are any spaces for this agent
	if len(ipfsm.W3Agent.Spaces) > 0 {

		// Get list of uploads in the active space
		err = ipfsm.W3Agent.UploadList()
		if err != nil {
			logger.Manager.Package["ipfs"].Error().Msg(err.Error())
			return err
		}

		// log currently used w3up spaces
		logger.Manager.Package["ipfs"].Info().Msg(fmt.Sprintf(" [#] w3up current space: %v (%v)", ipfsm.W3Agent.ActiveSpace, ipfsm.W3Agent.Spaces[ipfsm.W3Agent.ActiveSpace].Name))

	}

	return err

}

// Deinitialize the ipfs manager
func (ipfsm *PackageManager) DeInit() error {
	var err error

	// log event
	logger.Manager.Package["ipfs"].Debug().Msg("Deinitializing the IPFS manager ...")

	// stop the local IPFS node
	err = ipfsm.IpfsNode.Close()
	if err == nil {

		// log debug event
		logger.Manager.Package["ipfs"].Info().Msg(" [#] Closed the local IPFS node")

	}
	ipfsm.IpfsContextCancel()

	return err

}

// Start the local IPFS node
func (ipfsm *PackageManager) StartLocalNode() (*core.IpfsNode, error) {
	var err error

	// IPFS Repository
	// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	// create a local repository path, if it does not exist
	ipfsm.IpfsRepoPath = filepath.Join(GetAppDataPath(), RENDERHIVE_APP_DIRECTORY_IPFS_REPO)
	if _, err := os.Stat(ipfsm.IpfsRepoPath); os.IsNotExist(err) {

		err = os.MkdirAll(ipfsm.IpfsRepoPath, 0700)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Could not create IPFS repository path '%v'.", ipfsm.IpfsRepoPath))
		}

	}

	// IPFS Plugins
	// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	// Load any external plugins if available on externalPluginsPath
	plugins, err := loader.NewPluginLoader(filepath.Join(ipfsm.IpfsRepoPath, "plugins"))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error loading plugins: %s", err))
	}

	// Load preloaded and external plugins
	if err := plugins.Initialize(); err != nil {
		return nil, errors.New(fmt.Sprintf("Error initializing plugins: %s", err))
	}

	if err := plugins.Inject(); err != nil {
		return nil, errors.New(fmt.Sprintf("Error initializing plugins: %s", err))
	}

	// IPFS Repo
	// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	// Try to open the repo
	ipfsm.IpfsRepo, err = fsrepo.Open(ipfsm.IpfsRepoPath)
	if err != nil {

		// Create a config with default options and a 2048 bit key
		cfg, err := config.Init(io.Discard, 2048)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Could not init IPFS repo configuration: %v", err.Error()))
		}

		// Enable experimental features
		// https://github.com/ipfs/kubo/blob/master/docs/experimental-features.md#ipfs-filestore
		cfg.Experimental.FilestoreEnabled = false
		// https://github.com/ipfs/kubo/blob/master/docs/experimental-features.md#ipfs-urlstore
		cfg.Experimental.UrlstoreEnabled = false
		// https://github.com/ipfs/kubo/blob/master/docs/experimental-features.md#ipfs-p2p
		cfg.Experimental.Libp2pStreamMounting = false
		// https://github.com/ipfs/kubo/blob/master/docs/experimental-features.md#p2p-http-proxy
		cfg.Experimental.P2pHttpProxy = false

		// Create the repo with the defined configuration
		err = fsrepo.Init(ipfsm.IpfsRepoPath, cfg)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to init IPFS repo: %v", err.Error()))
		}

		// Try to open the repo again
		ipfsm.IpfsRepo, err = fsrepo.Open(ipfsm.IpfsRepoPath)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to open IPFS repo: %v", err.Error()))
		}

		// log debug event
		logger.Manager.Package["ipfs"].Info().Msg(fmt.Sprintf(" [#] Created a new local IPFS repo in '%v'", ipfsm.IpfsRepoPath))

	}

	// Public IP for announcing IPFS node
	// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	// TODO: Need to handle the case when the computer disconnects from the internet
	//       and the IP address changes. Probably we will need to restart the node.

	// Get node configuration
	cfg, err := ipfsm.IpfsRepo.Config()

	// get the public IPv4 and add it to the announced addresses
	ipv4, err := GetPublicIPv4()
	if err != nil {
		return nil, err
	}
	if ipv4 != "" {

		// add the multiaddr with the public IP to the configuration
		cfg.Addresses.AppendAnnounce = append(cfg.Addresses.AppendAnnounce, fmt.Sprintf("/ip4/%v/tcp/4001", ipv4))
		cfg.Addresses.AppendAnnounce = append(cfg.Addresses.AppendAnnounce, fmt.Sprintf("/ip4/%v/udp/4001/quic", ipv4))
		cfg.Addresses.AppendAnnounce = append(cfg.Addresses.AppendAnnounce, fmt.Sprintf("/ip4/%v/udp/4001/quic-v1", ipv4))
		cfg.Addresses.AppendAnnounce = append(cfg.Addresses.AppendAnnounce, fmt.Sprintf("/ip4/%v/udp/4001/quic-v1/webtransport", ipv4))

	}
	logger.Manager.Package["ipfs"].Info().Msg(fmt.Sprintf(" [#] Queried IPv4 address: %v", ipv4))

	// get the public IPv6 and add it to the announced addresses
	ipv6, err := GetPublicIPv6()
	if err != nil {
		return nil, err
	}
	if ipv6 != "" {

		// add the multiaddr with the public IP to the configuration
		cfg.Addresses.AppendAnnounce = append(cfg.Addresses.AppendAnnounce, fmt.Sprintf("/ip6/%v/tcp/4001", ipv6))
		cfg.Addresses.AppendAnnounce = append(cfg.Addresses.AppendAnnounce, fmt.Sprintf("/ip6/%v/udp/4001/quic", ipv6))
		cfg.Addresses.AppendAnnounce = append(cfg.Addresses.AppendAnnounce, fmt.Sprintf("/ip6/%v/udp/4001/quic-v1", ipv6))
		cfg.Addresses.AppendAnnounce = append(cfg.Addresses.AppendAnnounce, fmt.Sprintf("/ip6/%v/udp/4001/quic-v1/webtransport", ipv6))

	}
	logger.Manager.Package["ipfs"].Info().Msg(fmt.Sprintf(" [#] Queried IPv6 address: %v", ipv6))

	// Start IPFS Node
	// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	// Create a context with cancel function
	ipfsm.IpfsContext, ipfsm.IpfsContextCancel = context.WithCancel(context.Background())

	// Spwan the local IPFS node
	ipfsm.IpfsNode, err = core.NewNode(ipfsm.IpfsContext, &core.BuildCfg{
		Online:  true,
		Routing: libp2p.DHTOption, // This option sets the node to be a full DHT node (both fetching and storing DHT Records)
		//Routing: libp2p.DHTClientOption, // This option sets the node to be a client DHT node (only fetching records)
		Repo: ipfsm.IpfsRepo,
	})
	if err != nil {
		return nil, err
	}

	// create the coreAPI interface for this node
	ipfsm.IpfsAPI, err = coreapi.NewCoreAPI(ipfsm.IpfsNode)
	if err != nil {
		return nil, err
	}

	// log debug event
	logger.Manager.Package["ipfs"].Info().Msg(fmt.Sprintf(" [#] Initialized local node in '%v'", ipfsm.IpfsRepoPath))
	logger.Manager.Package["ipfs"].Info().Msg(fmt.Sprintf(" [#] PeerID: %v", ipfsm.IpfsNode.Identity.String()))

	// if the node is online
	if ipfsm.IpfsNode.IsOnline {

		// wait until the node is connected to a minimum amount of peers or the timeout
		// passed
		start := time.Now()
		for {

			// get peer connections
			peers, err := ipfsm.GetConnectedPeers()
			if ipfsm.IpfsNode == nil {
				return nil, err
			}

			// check if the node is connected to a minimum amount of peers
			if len(peers) == 0 && time.Now().Sub(start) > 10*time.Second {
				return nil, errors.New(fmt.Sprintf(" [#] Failed to bootstrap (no peers found)"))
			} else if len(peers) >= 4 {
				logger.Manager.Package["ipfs"].Info().Msg(fmt.Sprintf(" [#] IPFS node is now connected to %v peers", len(peers)))

				// Test pinning and writing file to disk
				// working directory
				// wd, err = os.Getwd()
				// if err != nil {
				//    ipfsm.PinObject("bafybeifpaez32hlrz5tmr7scndxtjgw3auuloyuyxblynqmjw5saapewmu")
				//    ipfsm.GetObject("bafybeifpaez32hlrz5tmr7scndxtjgw3auuloyuyxblynqmjw5saapewmu", filepath.Join(wd, "tmp"))
				// }
				break
			}

			// wait for some ms
			time.Sleep(100 * time.Millisecond)

		}
	}

	return ipfsm.IpfsNode, nil

}

// Get the peers this node is connected to
func (ipfsm *PackageManager) GetConnectedPeers() ([]icore.ConnectionInfo, error) {
	var err error

	// if the local node is active
	if ipfsm.IpfsNode == nil {
		return nil, errors.New(fmt.Sprintf("No IPFS node found"))
	}

	// get connected peers
	peers, err := ipfsm.IpfsAPI.Swarm().Peers(ipfsm.IpfsContext)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to read swarm peers. Error: %v", err))
	}

	return peers, err

}

// Connect to a peer with a given multiaddr
func (ipfsm *PackageManager) SwarmConnect(address string) (*peer.AddrInfo, error) {
	var err error
	var peerAddr *peer.AddrInfo

	peerAddr, err = peer.AddrInfoFromString(address)
	if err != nil {
		peerAddr, err = peer.AddrInfoFromString(filepath.Join("/p2p", address))
		if err != nil {
			return nil, err
		}
	}

	err = ipfsm.IpfsAPI.Swarm().Connect(ipfsm.IpfsContext, *peerAddr)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error connecting to peer: %v", err))
	}

	return peerAddr, err
}

// Disconnect from a given multiaddr
func (ipfsm *PackageManager) SwarmDisconnect(address string) error {
	var err error

	peerAddr, err := ma.NewMultiaddr(address)
	if err != nil {
		if err != nil {
			return err
		}
	}

	err = ipfsm.IpfsAPI.Swarm().Disconnect(ipfsm.IpfsContext, peerAddr)
	if err != nil {
		return errors.New(fmt.Sprintf("Error disconnecting from peer: %v", err))
	}

	return err
}

// Calculate only the hash without putting the file on IPFS
func (ipfsm *PackageManager) GetOnlyHash(path string) (string, error) {
	var err error

	// prepare the local file for storage
	stat, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	file, err := files.NewSerialFile(path, false, stat)
	if err != nil {
		return "", err
	}

	cid, err := ipfsm.IpfsAPI.Unixfs().Add(ipfsm.IpfsContext, file, ioptions.Unixfs.HashOnly(true))
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to calculate only hash: %v", err.Error()))
	}

	return cid.String(), nil

}

// Add a file/directory on the local IPFS node
func (ipfsm *PackageManager) AddObject(path string, pin bool) (string, error) {
	var err error

	// prepare the local file for storage
	stat, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	file, err := files.NewSerialFile(path, false, stat)
	if err != nil {
		return "", err
	}

	cid, err := ipfsm.IpfsAPI.Unixfs().Add(ipfsm.IpfsContext, file, ioptions.Unixfs.Pin(pin))
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to put file on the IPFS node: %v", err.Error()))
	}

	return cid.String(), nil

}

// Get a file/directory from IPFS and write it to a local path
func (ipfsm *PackageManager) GetObject(cid_string string, outputPath string) (string, error) {
	var err error

	// get a CID object from the string
	cidPath := icorepath.New(cid_string)

	// try to retrieve the file/directory
	rootNode, err := ipfsm.IpfsAPI.Unixfs().Get(ipfsm.IpfsContext, cidPath)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not get file with CID: %s", err))
	}

	err = files.WriteTo(rootNode, outputPath)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not write out the fetched CID: %s", err))
	}

	return outputPath, err

}

// Pin a file based on the CID on the local IPFS node
func (ipfsm *PackageManager) PinObject(cid_string string) (bool, error) {
	var err error

	// only if a CID was passed
	if cid_string == "" {
		return false, errors.New(fmt.Sprintf("Could not pin IPFS object '%v': Not a valid CID.", cid_string))
	}

	// get a path object from the CID string
	ipfsPath := icorepath.New(cid_string)

	// test, if file is already pinned
	_, pinned, err := ipfsm.IpfsAPI.Pin().IsPinned(ipfsm.IpfsContext, ipfsPath)
	if err != nil {
		logger.Manager.Package["ipfs"].Trace().Msg(fmt.Sprintf("Could not pin IPFS object '%v': %v", cid_string, err.Error()))
		return false, errors.New(fmt.Sprintf("Could not pin '%v': %s", ipfsPath, err))
	}

	// if the file is already pinned, don't try it again
	if !pinned {

		// Check if object is advertised in the DHT (i.e., if at least one provider exists)
		_, err := ipfsm.IpfsAPI.Dht().FindProviders(context.Background(), ipfsPath, ioptions.Dht.NumProviders(1))
		if err != nil {
			logger.Manager.Package["ipfs"].Trace().Msg(fmt.Sprintf("Could not pin IPFS object '%v': %v", cid_string, err.Error()))
			return false, errors.New(fmt.Sprintf("The file '%v' is not advertised in the DHT yet.", cid_string))
		}

		// pin the file
		err = ipfsm.IpfsAPI.Pin().Add(ipfsm.IpfsContext, ipfsPath)
		if err != nil {
			logger.Manager.Package["ipfs"].Trace().Msg(fmt.Sprintf("Could not pin IPFS object '%v': %v", ipfsPath, err.Error()))
			return false, errors.New(fmt.Sprintf("Could not pin '%v': %s", ipfsPath, err))
		}

		// test, if file is now pinned
		_, pinned, err = ipfsm.IpfsAPI.Pin().IsPinned(ipfsm.IpfsContext, ipfsPath)
		if err != nil {
			logger.Manager.Package["ipfs"].Trace().Msg(fmt.Sprintf("Could not pin IPFS object '%v': %v", ipfsPath, err.Error()))
			return false, errors.New(fmt.Sprintf("Could not pin '%v': %s", ipfsPath, err))
		}

		logger.Manager.Package["ipfs"].Trace().Msg(fmt.Sprintf("Successfully pinned IPFS object: %v", ipfsPath))

	} else {

		logger.Manager.Package["ipfs"].Trace().Msg(fmt.Sprintf("IPFS object already pinned: %v", ipfsPath))

	}

	return pinned, nil

}

// Unpin a file based on the CID on the local IPFS node
func (ipfsm *PackageManager) UnPinObject(cid_string string) (bool, error) {
	var err error

	// get a path object from the CID string
	ipfsPath := icorepath.New(cid_string)

	// unpin the file
	err = ipfsm.IpfsAPI.Pin().Rm(ipfsm.IpfsContext, ipfsPath)
	if err != nil {
		return false, errors.New(fmt.Sprintf("Could not unpin '%v': %s", ipfsPath, err))
	}

	// test, if file is pinned
	_, pinned, err := ipfsm.IpfsAPI.Pin().IsPinned(ipfsm.IpfsContext, ipfsPath)
	if err != nil {
		return false, errors.New(fmt.Sprintf("Could not pin '%v': %s", ipfsPath, err))
	}

	logger.Manager.Package["ipfs"].Debug().Msg(fmt.Sprintf(" [#] Pinning status after attempt to unpin: %v", pinned))

	return pinned, nil

}

// Start HTTP server and webUI
func (ipfsm *PackageManager) StartHTTPServer(path string) error {
	var err error

	var opts = []corehttp.ServeOption{
		corehttp.GatewayOption(true, "/ipfs", "/ipns"),
		corehttp.WebUIOption,
		corehttp.CommandsOption(
			commands.Context{
				ConfigRoot: ipfsm.IpfsRepoPath,
				ConstructNode: func() (*core.IpfsNode, error) {
					return ipfsm.IpfsNode, nil
				},
			}),
	}
	proc := process.WithParent(process.Background())
	proc.Go(func(p process.Process) {
		if err := corehttp.ListenAndServe(ipfsm.IpfsNode, "/ip4/127.0.0.1/tcp/5001", opts...); err != nil {
			return
		}
	})

	return err

}

// IPFS MANAGER COMMAND LINE INTERFACE
// #############################################################################
// Create the command for the command line interface
func (ipfsm *PackageManager) CreateCommand() *cobra.Command {

	// create the package command
	ipfsm.Command = &cobra.Command{
		Use:   "ipfs",
		Short: "Commands for the interaction with the IPFS and Filecoin services",
		Long:  "This command and its sub-commands enable the interaction with the IPFS and Filecoin services required by the Renderhive network",
		Run: func(cmd *cobra.Command, args []string) {

			return

		},
	}

	// add the subcommands (IPFS)
	ipfsm.Command.AddCommand(ipfsm.CreateCommandInfo())
	ipfsm.Command.AddCommand(ipfsm.CreateCommandSwarm())
	ipfsm.Command.AddCommand(ipfsm.CreateCommandAdd())
	ipfsm.Command.AddCommand(ipfsm.CreateCommandGet())
	ipfsm.Command.AddCommand(ipfsm.CreateCommandPin())

	// add the subcommands (Filecoin / w3up service)
	ipfsm.Command.AddCommand(ipfsm.CreateCommandW3())

	return ipfsm.Command

}

// Create the CLI command to get information about the IPFS node
func (ipfsm *PackageManager) CreateCommandInfo() *cobra.Command {

	// flags for the 'info' command
	// var path string

	// create a 'info' command for the node
	command := &cobra.Command{
		Use:   "info",
		Short: "Print information about the IPFS repo",
		Long:  "This command prints the configuration of the IPFS repo.",
		Run: func(cmd *cobra.Command, args []string) {

			// check if the repo is initialized
			if ipfsm.IpfsRepo != nil {

				cfg, err := ipfsm.IpfsRepo.Config()
				if err != nil {

					fmt.Println("")
					fmt.Println(fmt.Errorf("Could retrieve configuration of the repo."))
					fmt.Println("")

				}

				// convert to JSON string
				jsonString, err := json.MarshalIndent(cfg, "", "\t")
				if err != nil {
					fmt.Println(err)
					return
				}

				// print the configuration
				fmt.Println("")
				fmt.Println(string(jsonString))
				fmt.Println("")

			} else {

				fmt.Println("")
				fmt.Println(fmt.Errorf("Could not find repo."))
				fmt.Println("")

			}

			return

		},
	}

	return command

}

// Create the CLI command to manage the IPFS swarm
func (ipfsm *PackageManager) CreateCommandSwarm() *cobra.Command {

	// flags for the 'swarm' command
	// var path string

	// create a 'swarm' command for the node
	command := &cobra.Command{
		Use:   "swarm",
		Short: "Manage connections to the p2p network",
		Long:  "A tool to manipulate the network swarm. The swarm is the component that opens, listens for, and maintains connections to other ipfs peers in the internet.",
		Run: func(cmd *cobra.Command, args []string) {

			// check if the repo is initialized
			if ipfsm.IpfsRepo == nil {

				fmt.Println("")
				fmt.Println(fmt.Errorf("Could not find repo."))
				fmt.Println("")

			}

			return

		},
	}

	// add the subcommands
	command.AddCommand(ipfsm.CreateCommandSwarm_Connect())
	command.AddCommand(ipfsm.CreateCommandSwarm_Disconnect())
	command.AddCommand(ipfsm.CreateCommandSwarm_Peers())

	return command

}

// Create the CLI command to connect to a specific IPFS node
func (ipfsm *PackageManager) CreateCommandSwarm_Connect() *cobra.Command {

	// flags for the 'swarm connect' command
	// var path string

	// create a 'swarm connect' command for the node
	command := &cobra.Command{
		Use:   "connect <address>",
		Short: "Open connection to a given address",
		Long:  "This command opens a new direct connection to a peer address, where the address is given in multiaddr format.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			// check if the repo is initialized
			if ipfsm.IpfsRepo != nil {

				// connect to the peer
				addrInfo, err := ipfsm.SwarmConnect(args[0])
				if err != nil {

					fmt.Println("")
					fmt.Println(fmt.Errorf("Could not connect to the peer: %v", err.Error()))
					fmt.Println("")

					return

				}

				fmt.Println("")
				fmt.Println(fmt.Sprintf("Connected to peer:"))
				fmt.Println(fmt.Sprintf(" [#] ID: %v", addrInfo.ID))
				fmt.Println(fmt.Sprintf(" [#] Multiaddr: %v", addrInfo.Addrs))
				fmt.Println("")

			} else {

				fmt.Println("")
				fmt.Println(fmt.Errorf("Could not find repo."))
				fmt.Println("")

			}

			return

		},
	}

	return command

}

// Create the CLI command to disconnect from a specific IPFS node
func (ipfsm *PackageManager) CreateCommandSwarm_Disconnect() *cobra.Command {

	// flags for the 'swarm disconnect' command
	// var path string

	// create a 'swarm disconnect' command for the node
	command := &cobra.Command{
		Use:   "disconnect <address>",
		Short: "Close connection from a given address",
		Long:  "This command closes a direct connection to a peer address, where the address is given in multiaddr format.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			// check if the repo is initialized
			if ipfsm.IpfsRepo != nil {

				// connect to the peer
				err := ipfsm.SwarmDisconnect(args[0])
				if err != nil {

					fmt.Println("")
					fmt.Println(fmt.Errorf("Could not disconnect from the peer: %v", err.Error()))
					fmt.Println("")

					return

				}

				fmt.Println("")
				fmt.Println(fmt.Sprintf("Disconnected from peer: %v", args[0]))
				fmt.Println("")

			} else {

				fmt.Println("")
				fmt.Println(fmt.Errorf("Could not find repo."))
				fmt.Println("")

			}

			return

		},
	}

	return command

}

// Create the CLI command to list the connected peers of this IPFS node
func (ipfsm *PackageManager) CreateCommandSwarm_Peers() *cobra.Command {

	// flags for the 'swarm peers' command
	// var path string

	// create a 'swarm connect' command for the node
	command := &cobra.Command{
		Use:   "peers",
		Short: "List peers with open connections",
		Long:  "This command lists the set of peers this IPFS node is connected to.",
		Run: func(cmd *cobra.Command, args []string) {

			// check if the repo is initialized
			if ipfsm.IpfsRepo != nil {

				// get peer connections and print all
				peers, err := ipfsm.GetConnectedPeers()
				if ipfsm.IpfsNode == nil {

					fmt.Println("")
					fmt.Println(fmt.Errorf(err.Error()))
					fmt.Println("")

					return

				}
				fmt.Println("")
				for i, peer := range peers {
					fmt.Println(fmt.Sprintf("Peer %v:", i))
					fmt.Println(fmt.Sprintf(" [#] ID: %v", peer.ID()))
					fmt.Println(fmt.Sprintf(" [#] Connected via address: %v", peer.Address()))
					fmt.Println(fmt.Sprintf(" [#] Connection direction: %v", peer.Direction()))
				}
				fmt.Println("")

			} else {

				fmt.Println("")
				fmt.Println(fmt.Errorf("Could not find repo."))
				fmt.Println("")

			}

			return

		},
	}

	return command

}

// Create the CLI command to add a file from IPFS
func (ipfsm *PackageManager) CreateCommandAdd() *cobra.Command {

	// flags for the 'add' command
	var pin bool

	// create a 'add' command for the node
	command := &cobra.Command{
		Use:   "add <path>",
		Short: "Add a local file/directory to the IPFS node",
		Long:  "This command adds a local file/directory to the local IPFS node.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			// add the file/directory from IPFS and write it to the given path
			cid, err := ipfsm.AddObject(args[0], pin)
			if err != nil {

				fmt.Println("")
				fmt.Println(fmt.Errorf("Could not add '%v' to IPFS node: %v", args[0], err.Error()))
				fmt.Println("")

				return

			}

			fmt.Println("")
			fmt.Println("Added file/directory to IPFS:")
			fmt.Printf(" [#] Path: %v\n", args[0])
			fmt.Printf(" [#] CID: %v\n", cid)
			fmt.Println("")

			return

		},
	}

	// add command flags
	command.Flags().BoolVarP(&pin, "pin", "p", true, "Pin locally to protect added file/directoy from garbage collection. Default: true.")

	return command

}

// Create the CLI command to get a file from IPFS
func (ipfsm *PackageManager) CreateCommandGet() *cobra.Command {

	// flags for the 'get' command
	var path string

	// create a 'get' command for the node
	command := &cobra.Command{
		Use:   "get <ipfs-path>",
		Short: "Retrieve a file/directory from IPFS",
		Long:  "This command makes a GET request on the IPFS network to retrieve a file or directory.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			// get the file/directory from IPFS and write it to the given path
			cid, err := gocid.Parse(args[0])
			if err != nil {

				fmt.Println("")
				fmt.Println(fmt.Errorf("'%v' is not a valid CID.", args[0]))
				fmt.Println("")

			} else {

				// retrieve the file
				newpath, err := ipfsm.GetObject(cid.String(), path)
				if err != nil {

					fmt.Println("")
					fmt.Println(fmt.Errorf("Could not get file/directory with CID '%v': %v", cid.String(), err.Error()))
					fmt.Println("")

					return
				}

				fmt.Println("")
				fmt.Println("Retrieved file/directory from IPFS:")
				fmt.Printf(" [#] Path: %v\n", newpath)
				fmt.Println("")

			}

			return

		},
	}

	// add command flags
	command.Flags().StringVarP(&path, "path", "p", "", "Store the file/directory in the given folder")

	return command

}

// Create the CLI command to pin a file on IPFS
func (ipfsm *PackageManager) CreateCommandPin() *cobra.Command {

	// flags for the 'pin' command
	// var path string

	// create a 'pin' command for the node
	command := &cobra.Command{
		Use:   "pin <ipfs-path>",
		Short: "Pin (and unpin) objects to local IPFS node storage.",
		Long:  "Stores an IPFS object(s) from a given path locally to disk.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			// pin the object on the local IPFS node
			cid, err := gocid.Parse(args[0])
			if err != nil {

				fmt.Println("")
				fmt.Println(fmt.Errorf("'%v' is not a valid CID.", args[0]))
				fmt.Println("")

			} else {

				// pin the object
				_, err := ipfsm.PinObject(cid.String())
				if err != nil {

					fmt.Println("")
					fmt.Println(fmt.Errorf("Could not pin file/directory with CID '%v': %v", cid.String(), err.Error()))
					fmt.Println("")

					return
				}

				fmt.Println("")
				fmt.Println("Pinned file/directory on local IPFS node:")
				fmt.Printf(" [#] CID: %v\n", cid.String())
				fmt.Println("")

			}

			return

		},
	}

	// add command flags
	// command.Flags().StringVarP(&var, "", "", "", "DESCRIPTION")

	return command

}

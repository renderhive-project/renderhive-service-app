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
  "fmt"
  "os"
  "io"
  "errors"
  "path/filepath"
  // "time"

  // external
	"github.com/ipfs/kubo/plugin/loader"
  "github.com/ipfs/kubo/config"
  "github.com/ipfs/kubo/repo"
  "github.com/ipfs/kubo/repo/fsrepo"
  "github.com/ipfs/kubo/core"
	"github.com/ipfs/kubo/core/node/libp2p"
  "github.com/spf13/cobra"

  // internal
  . "renderhive/globals"
  . "renderhive/utility"
  "renderhive/logger"
)

// structure for the IPFS manager
type PackageManager struct {

  // Local IPFS node
  IpfsContext context.Context
  IpfsContextCancel func()
  IpfsRepoPath string
  IpfsRepo repo.Repo
  IpfsNode *core.IpfsNode

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

    // Create the local IPFS node
    _, err = ipfsm.CreateLocalNode()
    if err != nil {
        logger.Manager.Package["ipfs"].Error().Msg(err.Error())
    }

    return err

}

// Deinitialize the ipfs manager
func (ipfsm *PackageManager) DeInit() (error) {
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

// Initilize the local IPFS node
func (ipfsm *PackageManager) CreateLocalNode() (*core.IpfsNode, error) {
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
        return nil, errors.New(fmt.Sprintf("Could not init IPFS repo configuration:", err.Error()))
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

  // IPFS Node
  // +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
  // Create a context with cancel function
  ipfsm.IpfsContext, ipfsm.IpfsContextCancel = context.WithCancel(context.Background())

  // Spwan the local IPFS node
  ipfsm.IpfsNode, err = core.NewNode(ipfsm.IpfsContext, &core.BuildCfg{
  	Online: false,
		Routing: libp2p.DHTOption, // This option sets the node to be a full DHT node (both fetching and storing DHT Records)
		//Routing: libp2p.DHTClientOption, // This option sets the node to be a client DHT node (only fetching records)
		Repo: ipfsm.IpfsRepo,
  })
  if err != nil {
  	return nil, err
  }

  // log debug event
  logger.Manager.Package["ipfs"].Info().Msg(fmt.Sprintf(" [#] Initialized local node in '%v'", ipfsm.IpfsRepoPath))

  return ipfsm.IpfsNode, nil

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

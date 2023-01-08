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

  // internal
  "renderhive/logger"
  // "renderhive/constants"
  // "renderhive/hedera"
)

// structure for the IPFS manager
type IPFSManager struct {

  Placeholder string

}


// IPFS MANAGER
// #############################################################################
// Initialize everything required for the IPFS management
func InitIPFSManager() (*IPFSManager, error) {
    var err error

    // log information
    logger.RenderhiveLogger.Package["ipfs"].Info().Msg("Initializing the IPFS manager ...")

    // create a new IPFS manager
    ipfsm := &IPFSManager{}

    return ipfsm, err

}

// Deinitialize the ipfs manager
func (ipfsm *IPFSManager) DeInitIPFSManager() (error) {
    var err error

    // log information
    logger.RenderhiveLogger.Package["ipfs"].Info().Msg("Deinitializing the IPFS manager ...")

    return err

}

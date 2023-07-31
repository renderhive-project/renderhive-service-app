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

The w3up service (future service of web3.storage) is used as the filecoin
pinning service by the Renderhive. It will make sure that certain files
remain available to the network in case the file owner goes offline. This,
for example, includes .blend files of render jobs that were submitted by a
node that subsequently goes online.

*/

import (

	// standard

	// external

	// internal
	. "renderhive/globals"
	"renderhive/logger"
)

// Hedera account / wallet data
type W3upAgent struct {

	// Hedera account ID
	PLACEHOLDER int
}

// W3UP FUNCTIONS
// #############################################################################
// Initialize everything required for the IPFS management
func (ipfsm *PackageManager) Register() error {
	var err error

	// log information
	logger.Manager.Package["ipfs"].Info().Msg("Initializing the IPFS manager ...")

	// Create the local IPFS node
	_, err = ipfsm.StartLocalNode()
	if err != nil {
		logger.Manager.Package["ipfs"].Error().Msg(err.Error())
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

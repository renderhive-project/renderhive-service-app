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

package jsonrpc

/*

  The (smart) contract service enables the service app to interact with Renderhive's smart contracts via the JSON-RPC.

*/

import (

	// standard
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	// external
	//  hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

	// internal
	. "renderhive/globals"

	"renderhive/ipfs"
	"renderhive/logger"
	"renderhive/node"
)

// SERVICE INITIALIZATION
// #############################################################################

// export the NodeService for net/rpc
type NodeService struct{}

// RENDERHIVE NODE SERVICE – RENDER OFFERS
// #############################################################################

// Method: CreateRenderOffer
// 			- create a new render offer on the local node
// #############################################################################

// Method
func (ops *NodeService) CreateRenderOffer(r *http.Request, args *CreateRenderOfferArgs, reply *CreateRenderOfferReply) error {
	var err error
	var offerCID string

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// TODO: Implement further checks and security measures

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf("Creating a new render offer"))

	// create the render offer
	offer, err := node.Manager.NewRenderOffer(args.Price)
	if err != nil {
		return fmt.Errorf("Could not create new render offer: %v", err)
	}

	// iterate over the Blender versions and add them to the offer
	for _, blender := range args.BlenderVersions {

		// TODO: Find the blender version in the local file system
		path := "/Applications/Blender 3.00.app/Contents/MacOS/blender"

		// add the blender version to the offer
		err = offer.AddBlenderVersion(blender.Version, path, &blender.Engines, &blender.Devices, blender.Threads)
		if err != nil {
			return fmt.Errorf("Could not add blender version to render offer: %v", err)
		}

		// log info
		logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Blender version '%v' added to render offer", blender.Version))

	}

	// deploy the render offer to the local IPFS
	offerCID, err = offer.Deploy()
	if err != nil {
		return fmt.Errorf("Could not deploy the render offer: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Render offer document uploaded with CID: %v", offer.DocumentCID))

	// set a reply message
	reply.Message = "New render offer was created locally: http://localhost:5001/ipfs/" + offerCID + "!"

	// create reply for the RPC client
	return nil

}

// Method: SubmitRenderOffer
// 			- submit a locally created render offer to the network
// #############################################################################

// Method
func (ops *NodeService) SubmitRenderOffer(r *http.Request, args *SubmitRenderOfferArgs, reply *SubmitRenderOfferReply) error {
	var err error
	var transactionBytes []byte

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// TODO: Implement further checks and security measures

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf("Submitting a render offer to the renderhive network"))

	// deploy the render offer to the local IPFS
	offer, err := node.Manager.GetRenderOffer(args.RenderOfferCID)
	if err != nil {
		return fmt.Errorf("Failed to submit render offer: %v", err)
	}

	// submit the render offer to the network
	_, transactionBytes, err = offer.Submit()
	if err != nil {
		return fmt.Errorf("Failed to submit render offer: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Sending transaction bytes to frontend for execution with operator wallet"))

	// set a reply message
	// networkName, _ := hedera.Manager.NetworkClient.GetLedgerID().ToNetworkName()
	reply.Message = "" //"Render offer was successfully submitted: http://hashscan.io/" + networkName.String() + "/transaction/" + receipt.TransactionID.String() + "!"
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// Method: PauseRenderOffer
// 			- pause a locally created render offer to the network
// #############################################################################

// Method
func (ops *NodeService) PauseRenderOffer(r *http.Request, args *PauseRenderOfferArgs, reply *PauseRenderOfferReply) error {
	var err error
	var transactionBytes []byte

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// TODO: Implement further checks and security measures

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf("Submitting a render offer to the renderhive network"))

	// deploy the render offer to the local IPFS
	offer, err := node.Manager.GetRenderOffer(args.RenderOfferCID)
	if err != nil {
		return fmt.Errorf("Failed to submit render offer: %v", err)
	}

	// pause the render offer
	_, transactionBytes, err = offer.Pause()
	if err != nil {
		return fmt.Errorf("Failed to submit render offer: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Sending transaction bytes to frontend for execution with operator wallet"))

	// set a reply message
	// networkName, _ := hedera.Manager.NetworkClient.GetLedgerID().ToNetworkName()
	reply.Message = "" //"Render offer was successfully submitted: http://hashscan.io/" + networkName.String() + "/transaction/" + receipt.TransactionID.String() + "!"
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// RENDERHIVE NODE SERVICE – RENDER REQUESTS
// #############################################################################

// Method: CreateRenderRequest
// 			- create a new render request on the local node
// #############################################################################

// Method
func (ops *NodeService) CreateRenderRequest(r *http.Request, args *CreateRenderRequestArgs, reply *CreateRenderRequestReply) error {
	var err error
	var requestCID string

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// TODO: Implement further checks and security measures

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf("Creating a new render request"))

	// create the render request
	request, err := node.Manager.NewRenderRequest(args.Blender.Version, args.Price)
	if err != nil {
		return fmt.Errorf("Could not create new render request: %v", err)
	}

	// Iterate over the file data and add each file to the request
	for _, file := range args.Files {

		// Decode the Base64 file data
		fileData, err := base64.StdEncoding.DecodeString(file.FileData)
		if err != nil {
			return fmt.Errorf("Error decoding file data: %v", err)
		}

		// add the file to the request
		err = request.AddFileFromBytes(file.FileName, fileData)
		if err != nil {
			return fmt.Errorf("Could not add file to request: %v", err)
		}

		// if the filename contains the .blend file suffix
		if strings.ToLower(filepath.Ext(file.FileName)) == ".blend" {

			// check if there was no blender file added yet
			if request.BlenderFile.CID != "" {
				return fmt.Errorf("Only one .blend file is allowed per render request")
			}

			// get the CID of the .blend file
			request.BlenderFile.CID, err = ipfs.Manager.GetHashFromObject(request.Files[file.FileName])
			if err != nil {
				return fmt.Errorf("Could not get CID of .blend file: %v", err)
			}

		}

		// log info
		logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] File '%v' (%v bytes) added to directory", file.FileName, len(fileData)))

	}

	// check if there was a blender file added
	if request.BlenderFile.CID == "" {
		return fmt.Errorf("No .blend file was added to the render request")
	}

	// deploy the render request to the local IPFS
	requestCID, err = request.Deploy()
	if err != nil {
		return fmt.Errorf("Could not deploy render request: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Render request document uploaded with CID: %v", request.DocumentCID))
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Directory uploaded with CID: %v", request.DirectoryCID))

	// set a reply message
	reply.Message = "New render request was created locally: http://localhost:5001/ipfs/" + requestCID + "!"

	// create reply for the RPC client
	return nil

}

// Method: SubmitRenderRequest
// 			- submit a locally created render request to the network
// #############################################################################

// Method
func (ops *NodeService) SubmitRenderRequest(r *http.Request, args *SubmitRenderRequestArgs, reply *SubmitRenderRequestReply) error {
	var err error
	var transactionBytes []byte

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// TODO: Implement further checks and security measures

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf("Submitting a render request to the renderhive network"))

	// deploy the render request to the local IPFS
	request, err := node.Manager.GetRenderRequest(args.RenderRequestCID)
	if err != nil {
		return fmt.Errorf("Failed to submit render request: %v", err)
	}

	// submit the render request to the network
	_, transactionBytes, err = request.Submit()
	if err != nil {
		return fmt.Errorf("Failed to submit render request: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Sending transaction bytes to frontend for execution with operator wallet"))

	// set a reply message
	// networkName, _ := hedera.Manager.NetworkClient.GetLedgerID().ToNetworkName()
	reply.Message = "" //"Render request was successfully submitted: http://hashscan.io/" + networkName.String() + "/transaction/" + receipt.TransactionID.String() + "!"
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// Method: CancelRenderRequest
// 			- cancel a render request
// #############################################################################

// Method
func (ops *NodeService) CancelRenderRequest(r *http.Request, args *CancelRenderRequestArgs, reply *CancelRenderRequestReply) error {
	var err error
	var transactionBytes []byte

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// TODO: Implement further checks and security measures

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf("Submitting a render request to the renderhive network"))

	// deploy the render request to the local IPFS
	request, err := node.Manager.GetRenderRequest(args.RenderRequestCID)
	if err != nil {
		return fmt.Errorf("Failed to submit render request: %v", err)
	}

	// submit the render request to the network
	_, transactionBytes, err = request.Cancel()
	if err != nil {
		return fmt.Errorf("Failed to submit render request: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Sending transaction bytes to frontend for execution with operator wallet"))

	// set a reply message
	// networkName, _ := hedera.Manager.NetworkClient.GetLedgerID().ToNetworkName()
	reply.Message = "" //"Render request was successfully submitted: http://hashscan.io/" + networkName.String() + "/transaction/" + receipt.TransactionID.String() + "!"
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

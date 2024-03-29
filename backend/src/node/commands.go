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

/*

This file defines the Renderhive command protocol, which is used by nodes to communicate via the
Hedera Consensus Service (HCS) as a transport layer. Each command is received by the node's backend, which then filters,
checks, verifies and executes the command. For command execution, the node calls its own internal JSON-RPC with
the json rpc message included in the command.

Using the HCS as a transport layer for the JSON-RPC messages has the advantage, that the messages are immutable and
fairly ordered, preventing command collisions in the distributed network, which might otherwise occur due to latencies.
Furthermore, it provides an auditable trace of all commands and is inherently spam resistant, since each command causes
a network fee.

*/

import (

	// standard
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	// external
	// "github.com/cockroachdb/apd"
	// "golang.org/x/exp/slices" <-- would be handy, but requires Go 1.18; TODO: Update possible for Hedera SDK?
	// internal
	// . "renderhive/globals"
	// . "renderhive/utility"
)

// RENDERHIVE MESSAGING PROTOCOL
// #############################################################################

// define the version of the renderhive command protocol
const RENDERHIVE_COMMAND_PROTOCOL_VERSION string = "1.0"

// enum for service names
const (
	SERVICE_UNKNOWN  int = iota
	SERVICE_PING     int = 1 + iota
	SERVICE_CONTRACT int = 1001 + iota
	SERVICE_NODE     int = 2001 + iota
)

// enum for method names
const (

	// UNKNOWN
	METHOD_UNKNOWN int = iota

	// PING SERVICE
	METHOD_PING_SAYHELLO int = 1 + iota

	// CONTRACT SERVICE
	METHOD_CONTRACT_DEPLOY int = 1001 + iota
	METHOD_CONTRACT_GET_CURRENT_HIVECYCLE
	METHOD_CONTRACT_REGISTER_OPERATOR
	METHOD_CONTRACT_UNREGISTER_OPERATOR
	METHOD_CONTRACT_DEPOSIT_OPERATOR_FUNDS
	METHOD_CONTRACT_WITHDRAW_OPERATOR_FUNDS
	METHOD_CONTRACT_GET_OPERATOR_BALANCE
	METHOD_CONTRACT_GET_RESERVED_OPERATOR_FUNDS
	METHOD_CONTRACT_IS_OPERATOR
	METHOD_CONTRACT_GET_OPERATOR_LAST_ACTIVITY
	METHOD_CONTRACT_ADD_NODE
	METHOD_CONTRACT_REMOVE_NODE
	METHOD_CONTRACT_IS_NODE
	METHOD_CONTRACT_DEPOSIT_NODE_STAKE
	METHOD_CONTRACT_WITHDRAW_NODE_STAKE
	METHOD_CONTRACT_GET_NODE_STAKE
	METHOD_CONTRACT_ADD_RENDER_JOB
	METHOD_CONTRACT_CLAIM_RENDER_JOB

	// NODE SERVICE
	METHOD_NODE_CREATE_RENDER_REQUEST int = 2001 + iota
	METHOD_NODE_SUBMIT_RENDER_REQUEST
	METHOD_NODE_CANCEL_RENDER_REQUEST
	METHOD_NODE_CREATE_RENDER_OFFER
	METHOD_NODE_SUBMIT_RENDER_OFFER
	METHOD_NODE_PAUSE_RENDER_OFFER
)

// define the default message structure for the renderhive JSON-RPC
type JsonRpcMessage struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Id      int         `json:"id"`
}

// TODO: Should service and method be a string, or rather enums to save space and make renaming easier?
// define the default message structure for the renderhive JSON-RPC
type RenderhiveCommand struct {
	Version  string   `json:"ver"` // version of the renderhive command protocol
	Audience []string `json:"aud"` // list of node account addresses that need to receive the message (empty field means network wide broadcast)
	Message  []byte   `json:"rpc"` // the actual JSON-RPC message encoded in base64

	// // TODO: Add the following fields, which might be useful for verifying the network state in the smart contract later
	// //       The basic idea would be, that whenever a node submits a command, it verifies the state of the network until the previous cycle.
	// //       The merkle root of all valid transactions in the previous cycle is then included in the message.
	// // 		The merkle tree could be created in the following way: For each hive cycle, create a merkle tree.
	// // 		The merkle roots of all cycles are then used to create a merkle tree, which represents the network state.
	// // 		Hive cycles could also be treated as blocks, using block headers and hashes to essential create a blockchain on top of HCS.
	// Hivecycle int			// hive cycle during which this message was created
	// Root      string		// merkle root of valid transactions in the hive cycle ()
}

// get service name from int
func (nm *PackageManager) GetServiceName(service int) string {
	switch service {
	case SERVICE_PING:
		return "PingService"
	case SERVICE_CONTRACT:
		return "ContractService"
	case SERVICE_NODE:
		return "NodeService"
	default:
		return "Unknown"
	}
}

// get method name from int
func (nm *PackageManager) GetMethodName(method int) string {
	switch method {
	case METHOD_CONTRACT_DEPLOY:
		return "Deploy"
	case METHOD_CONTRACT_GET_CURRENT_HIVECYCLE:
		return "GetCurrentHiveCycle"

	// Node Service
	case METHOD_NODE_CREATE_RENDER_REQUEST:
		return "CreateRenderRequest"
	case METHOD_NODE_SUBMIT_RENDER_REQUEST:
		return "SubmitRenderRequest"
	case METHOD_NODE_CANCEL_RENDER_REQUEST:
		return "CancelRenderRequest"
	case METHOD_NODE_CREATE_RENDER_OFFER:
		return "CreateRenderOffer"
	case METHOD_NODE_SUBMIT_RENDER_OFFER:
		return "SubmitRenderOffer"
	case METHOD_NODE_PAUSE_RENDER_OFFER:
		return "PauseRenderOffer"
	default:
		return "Unknown"
	}
}

// get service and method int from a string in the form "SERVICE.METHOD"
func (nm *PackageManager) GetServiceAndMethodInt(methodString string) (int, int, error) {

	// split the string
	parts := strings.Split(methodString, ".")
	if len(parts) != 2 {
		return -1, -1, errors.New(fmt.Sprintf("Invalid service method string: %s", methodString))
	}

	// get the service and method
	service := SERVICE_UNKNOWN
	method := METHOD_UNKNOWN
	switch parts[0] {
	case "PingService":
		service = SERVICE_PING
	case "ContractService":
		service = SERVICE_CONTRACT
	case "NodeService":
		service = SERVICE_NODE
	}

	switch parts[1] {
	case "SayHello":
		method = METHOD_PING_SAYHELLO
	case "Deploy":
		method = METHOD_CONTRACT_DEPLOY
	case "GetCurrentHiveCycle":
		method = METHOD_CONTRACT_GET_CURRENT_HIVECYCLE
	case "CreateRenderRequest":
		method = METHOD_NODE_CREATE_RENDER_REQUEST
	case "SubmitRenderRequest":
		method = METHOD_NODE_SUBMIT_RENDER_REQUEST
	case "CancelRenderRequest":
		method = METHOD_NODE_CANCEL_RENDER_REQUEST
	case "CreateRenderOffer":
		method = METHOD_NODE_CREATE_RENDER_OFFER
	case "SubmitRenderOffer":
		method = METHOD_NODE_SUBMIT_RENDER_OFFER
	case "PauseRenderOffer":
		method = METHOD_NODE_PAUSE_RENDER_OFFER
	}

	return service, method, nil
}

// create a standard message for the renderhive JSON-RPC for submission to the Hedera Consensus Service
func (nm *PackageManager) EncodeCommand(audience []string, service int, method int, args interface{}) (string, error) {
	var err error

	// prepare the JSON-RPC message to encode
	rpc := JsonRpcMessage{
		Jsonrpc: "2.0",
		Method:  nm.GetServiceName(service) + "." + nm.GetMethodName(method),
		Params:  args,
		Id:      -1,
	}

	// convert the message to JSON
	jsonMessage, err := json.Marshal(rpc)
	if err != nil {
		return "", err
	}

	// encode the JSON-RPC message in base64
	encodedMessage := make([]byte, base64.StdEncoding.EncodedLen(len(jsonMessage)))
	base64.StdEncoding.Encode(encodedMessage, jsonMessage)

	// create the command
	message := &RenderhiveCommand{
		RENDERHIVE_COMMAND_PROTOCOL_VERSION,
		audience,
		encodedMessage,
	}

	// convert the message to JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return "", err
	}

	return string(messageJSON), nil

}

// decode a standard message for the renderhive JSON-RPC received on the Hedera Consensus Service
func (nm *PackageManager) DecodeCommand(jsonMessage interface{}) (*RenderhiveCommand, error) {
	var err error

	// if jsonMessage is a string, convert it to a byte array
	var message []byte
	switch jsonMessage.(type) {
	case string:
		message = []byte(jsonMessage.(string))
	case []byte:
		message = jsonMessage.([]byte)
	default:
		return nil, err
	}

	// create the message
	var command RenderhiveCommand
	err = json.Unmarshal(message, &command)
	if err != nil {
		return nil, err
	}

	return &command, nil

}

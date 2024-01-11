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

package jsonrpc

/*

 The (smart) contract service enables the service app to interact with Renderhive's smart contracts via the JSON-RPC.

*/

import (

	// standard

	"fmt"
	"net/http"

	// external
	hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

	// internal
	// . "renderhive/globals"
	"renderhive/hedera"
	"renderhive/logger"
)

// SERVICE INITIALIZATION
// #############################################################################

// export the ContractService for net/rpc
type ContractService struct{}

// HEDERA SMART CONTRACT FUNCTIONS
// #############################################################################

// Method: Deploy
// 			- deploy a given smart contract on Hedera
// #############################################################################

// Arguments and reply
type DeployArgs struct {
	ContractFilepath string
	Gas              int64
}
type DeployReply struct {
	Message string
}

// Deploys a new smart contract on Hedera
func (ops *ContractService) Deploy(r *http.Request, args *DeployArgs, reply *DeployReply) error {
	var err error

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// TODO: Implement further checks and security measures

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf("Starting to deploy smart contract (Gas: %v)", args.Gas))

	// prepare a new contract object
	contract := hedera.HederaSmartContract{}

	// deploy the new contract
	response, receipt, err := contract.NewFromBin(args.ContractFilepath, nil, args.Gas)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] New Contract ID: %v", receipt.ContractID.String()))

	// set a reply message
	reply.Message = "New smart contract was deployed as " + receipt.ContractID.String() + " with transaction: " + response.TransactionID.String() + "!"

	// create reply for the RPC client
	return nil

}

// RENDERHIVE SMART CONTRACT FUNCTIONS
// #############################################################################

// Method: registerOperator
// 			- register a new operator in the Renderhive Smart Contract
// #############################################################################

// Arguments and reply
type RegisterOperatorArgs struct {
	// PublicKey       *ecdsa.PublicKey // the public key of the operator
	ContractID      string // the ID of the smart contract
	OperatorTopicID string // the TopicID of the operator's HCS topic

	Gas uint64 // the gas limit for the transaction
}
type RegisterOperatorReply struct {
	Message string
}

// Register a new operator in the Renderhive Smart Contract
func (ops *ContractService) RegisterOperator(r *http.Request, args *RegisterOperatorArgs, reply *RegisterOperatorReply) error {
	var err error

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// TODO: Implement further checks and security measures

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf("Calling a smart contract function (Gas: %v)", args.Gas))

	// prepare the contract object
	contractID, err := hederasdk.ContractIDFromString(args.ContractID)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	contract := hedera.HederaSmartContract{ID: contractID}

	// prepare the parameters for the function call
	params := hederasdk.NewContractFunctionParameters().AddString(args.OperatorTopicID)
	fmt.Println("Params:", contract.ID.String())
	fmt.Println("Params:", params)

	// call the function
	response, _, err := contract.CallFunction("registerOperator", params, args.Gas)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Contract function called with transaction: %v", response.TransactionID.String()))

	// set a reply message
	reply.Message = "registerOperator function was called with transaction: " + response.TransactionID.String() + "!"

	// create reply for the RPC client
	return nil

}

// Method: registerOperator
// 			- register a new operator in the Renderhive Smart Contract
// #############################################################################

// Arguments and reply
type IsOperatorArgs struct {
	ContractID string // the ID of the smart contract
	AccountID  string // the ID of the account to query for
	Gas        uint64 // the gas limit for the transaction
}
type IsOperatorReply struct {
	Message string
	Value   bool
}

// Register a new operator in the Renderhive Smart Contract
func (ops *ContractService) IsOperator(r *http.Request, args *IsOperatorArgs, reply *IsOperatorReply) error {
	var err error

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// TODO: Implement further checks and security measures

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf("Calling a smart contract function (Gas: %v)", args.Gas))

	// prepare the contract object
	contractID, err := hederasdk.ContractIDFromString(args.ContractID)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	contract := hedera.HederaSmartContract{ID: contractID}

	// prepare the AccountID
	accountID, err := hederasdk.AccountIDFromString(args.AccountID)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// prepare the parameters for the function call
	params, err := hederasdk.NewContractFunctionParameters().AddAddress(accountID.ToSolidityAddress())
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// call the function
	response, receipt, err := contract.CallFunction("isOperator", params, args.Gas)
	fmt.Println("Response:", response)
	fmt.Println("Receipt:", receipt)
	fmt.Println("Error:", err)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Contract function called with transaction: %v", response.TransactionID.String()))

	// get the result of the function call

	record, err := response.GetRecord(hedera.Manager.NetworkClient)
	if err != nil {
		return fmt.Errorf("Error getting contract response record: %v", err)
	}

	functionResult, err := record.GetContractExecuteResult()
	if err != nil {
		return fmt.Errorf("Error getting contract execute result: %v", err)
	}

	// set a reply message
	reply.Message = "IsOperator function was called with transaction: " + response.TransactionID.String() + "!\n\n" + fmt.Sprintf("Result: %v", functionResult.GetBool(0))
	reply.Value = functionResult.GetBool(0)

	// create reply for the RPC client
	return nil

}

// INTERNAL HELPER FUNCTIONS
// #############################################################################

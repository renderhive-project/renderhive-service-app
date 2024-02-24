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
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"

	// external

	hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

	// internal
	. "renderhive/globals"
	"renderhive/hedera"
	"renderhive/logger"
	"renderhive/node"
)

// SERVICE INITIALIZATION
// #############################################################################

// export the ContractService for net/rpc
type ContractService struct{}

// HEDERA SMART CONTRACT – GENERAL FUNCTIONS
// #############################################################################

// Method: Deploy
// 			- deploy a given smart contract on Hedera
// #############################################################################

// Method
func (ops *ContractService) Deploy(r *http.Request, args *DeployArgs, reply *DeployReply) error {
	var err error
	var transactionBytes []byte

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// TODO: Implement further checks and security measures

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf("Starting to deploy smart contract (Gas: %v)", args.Gas))

	// prepare a new contract object
	contract := hedera.HederaSmartContract{}

	// deploy the new contract
	response, receipt, transactionBytes, err := contract.NewFromBin(args.ContractFilepath, nil, args.Gas)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] New Contract ID: %v", receipt.ContractID.String()))

	// set a reply message
	reply.Message = "New smart contract was deployed as " + receipt.ContractID.String() + " with transaction: " + response.TransactionID.String() + "!"
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// Method: GetCurrentHiveCycle
// 			- get the current hive cycle from the contract
// #############################################################################

// Method
func (ops *ContractService) GetCurrentHiveCycle(r *http.Request, args *GetCurrentHiveCycleArgs, reply *GetCurrentHiveCycleReply) error {
	var err error
	var transactionBytes []byte

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

	// call the function
	response, _, transactionBytes, err := contract.CallFunction("getCurrentHiveCycle", nil, args.Gas)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// get the result of the function call
	record, err := response.GetRecord(hedera.Manager.NetworkClient)
	if err != nil {
		return fmt.Errorf("Error getting contract response record: %v", err)
	}
	functionResult, err := record.GetContractExecuteResult()
	if err != nil {
		return fmt.Errorf("Error getting contract execute result: %v", err)
	}

	// set a reply value
	reply.Value = new(big.Int).SetBytes(functionResult.GetInt256(0))

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Contract function called with transaction: %v", response.TransactionID.String()))

	// set a reply message
	reply.Message = "getCurrentHiveCycle function was called with transaction: " + response.TransactionID.String() + "\n\n" + fmt.Sprintf("Result: %v", reply.Value)
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// RENDERHIVE SMART CONTRACT – OPERATOR MANAGEMENT
// #############################################################################

// Method: registerOperator
// 			- register a new operator in the Renderhive Smart Contract
// #############################################################################

// Method
func (ops *ContractService) RegisterOperator(r *http.Request, args *RegisterOperatorArgs, reply *RegisterOperatorReply) error {
	var err error
	var transactionBytes []byte

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
	_, _, transactionBytes, err = contract.CallFunction("registerOperator", params, args.Gas, hedera.TransactionOptions.SetExecute(false, node.Manager.User.UserAccount.AccountID))
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Sending transaction bytes to frontend for execution with operator wallet"))

	// set a reply message
	reply.Message = "" //"registerOperator function was called with transaction: " + response.TransactionID.String()
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// Method: unregisterOperator
// 			- unregister an operator from the Renderhive Smart Contract
// #############################################################################

// Method
func (ops *ContractService) UnregisterOperator(r *http.Request, args *UnregisterOperatorArgs, reply *UnregisterOperatorReply) error {
	var err error
	var transactionBytes []byte

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

	// call the function
	_, _, transactionBytes, err = contract.CallFunction("unregisterOperator", nil, args.Gas, hedera.TransactionOptions.SetExecute(false, node.Manager.User.UserAccount.AccountID))
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Sending transaction bytes to frontend for execution with operator wallet"))

	// set a reply message
	reply.Message = "" //"unregisterOperator function was called with transaction: " + response.TransactionID.String()
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// Method: depositOperatorFunds
// 			- deposit HBAR for an operator in the Renderhive Smart Contract
// #############################################################################

// Method
func (ops *ContractService) DepositOperatorFunds(r *http.Request, args *DepositOperatorFundsArgs, reply *DepositOperatorFundsReply) error {
	var err error
	var transactionBytes []byte

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

	// call the payable function
	response, receipt, transactionBytes, err := contract.CallPayableFunction("depositOperatorFunds", args.Amount, nil, args.Gas, hedera.TransactionOptions.SetExecute(false, node.Manager.User.UserAccount.AccountID))
	fmt.Println("Response:", response)
	fmt.Println("Receipt:", receipt)
	if err != nil {

		// // if a result is returned
		// if response != nil {
		// 	// get the error message, if any
		// 	record, err := response.GetRecordQuery().Execute(hedera.Manager.NetworkClient)
		// 	if err != nil {
		// 		return fmt.Errorf("Error getting contract response record: %v", err)
		// 	}

		// 	functionResult, err := record.GetContractExecuteResult()
		// 	if err != nil {
		// 		return fmt.Errorf("Error getting contract execute result: %v", err)
		// 	}

		// 	fmt.Println("Error (%v): %v", err, functionResult.ErrorMessage)
		// 	return fmt.Errorf("Error (%v): %v", err, functionResult.ErrorMessage)
		// }

		// fmt.Println("Error (%v): %v", err, "No details available")
		return fmt.Errorf("Error (%v): %v", err, "No details available")
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Sending transaction bytes to frontend for execution with operator wallet"))

	// set a reply message
	reply.Message = "" //"DepositOperatorFunds function was called with transaction: " + response.TransactionID.String()
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// Method: withdrawOperatorFunds
// 			- withdraw HBAR from the Renderhive Smart Contract
// #############################################################################

// Method
func (ops *ContractService) WithdrawOperatorFunds(r *http.Request, args *WithdrawOperatorFundsArgs, reply *WithdrawOperatorFundsReply) error {
	var err error
	var transactionBytes []byte

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

	// convert the string amount of HBAR to a Hbar object
	amount, err := hederasdk.HbarFromString(args.Amount)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// convert HBAR to TINYBAR as big.Int
	amountBigInt := new(big.Int).SetInt64(amount.AsTinybar())

	// prepare the parameters for the function call
	params := hederasdk.NewContractFunctionParameters().AddUint256BigInt(amountBigInt)
	fmt.Println("Params:", contract.ID.String())
	fmt.Println("Params:", params)

	// call the payable function
	_, _, transactionBytes, err = contract.CallFunction("withdrawOperatorFunds", params, args.Gas, hedera.TransactionOptions.SetExecute(false, node.Manager.User.UserAccount.AccountID))
	// fmt.Println("Response:", response)
	// fmt.Println("Receipt:", receipt)
	// fmt.Println("Error:", err)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Sending transaction bytes to frontend for execution with operator wallet"))

	// get the result of the function call

	// record, err := response.GetRecord(hedera.Manager.NetworkClient)
	// if err != nil {
	// 	return fmt.Errorf("Error getting contract response record: %v", err)
	// }

	// functionResult, err := record.GetContractExecuteResult()
	// if err != nil {
	// 	return fmt.Errorf("Error getting contract execute result: %v", err)
	// }

	// set a reply message
	reply.Message = "" //"WithdrawOperatorFunds function was called with transaction: " + response.TransactionID.String()
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// Method: getOperatorBalance
// 			- check the balance of an registered operator in the Renderhive Smart Contract
// #############################################################################

// Method
func (ops *ContractService) GetOperatorBalance(r *http.Request, args *GetOperatorBalanceArgs, reply *GetOperatorBalanceReply) error {
	var err error
	var transactionBytes []byte

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

	// prepare the passed AccountID
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
	response, _, transactionBytes, err := contract.CallFunction("getOperatorBalance", params, args.Gas)
	// fmt.Println("Response:", response)
	// fmt.Println("Receipt:", receipt)
	// fmt.Println("Error:", err)
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

	// set a reply value
	reply.Value = new(big.Int).SetBytes(functionResult.GetInt256(0))

	// convert tℏ to ℏ
	amount, err := hederasdk.HbarFromString(reply.Value.String() + " tℏ")
	if err != nil {
		return fmt.Errorf("Error getting contract response record: %v", err)
	}

	// set a reply message
	reply.Message = "getOperatorBalance function was called with transaction: " + response.TransactionID.String() + "\n\n" + fmt.Sprintf("Result: %v", amount)
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// Method: getReservedOperatorFunds
// 			- check the reserved funds of an registered operator in the Renderhive Smart Contract
// #############################################################################

// Method
func (ops *ContractService) GetReservedOperatorFunds(r *http.Request, args *GetReservedOperatorFundsArgs, reply *GetReservedOperatorFundsReply) error {
	var err error
	var transactionBytes []byte

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

	// prepare the passed AccountID
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
	response, _, transactionBytes, err := contract.CallFunction("getReservedOperatorFunds", params, args.Gas)
	// fmt.Println("Response:", response)
	// fmt.Println("Receipt:", receipt)
	// fmt.Println("Error:", err)
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

	// set a reply value
	reply.Value = new(big.Int).SetBytes(functionResult.GetInt256(0))

	// convert tℏ to ℏ
	amount, err := hederasdk.HbarFromString(reply.Value.String() + " tℏ")

	if err != nil {
		return fmt.Errorf("Error getting contract response record: %v", err)
	}

	// set a reply messages
	reply.Message = "getReservedOperatorFunds function was called with transaction: " + response.TransactionID.String() + "\n\n" + fmt.Sprintf("Result: %v", amount)
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// Method: isOperator
// 			- check if the given operator is registered in the Renderhive Smart Contract
// #############################################################################

// Method
func (ops *ContractService) IsOperator(r *http.Request, args *IsOperatorArgs, reply *IsOperatorReply) error {
	var err error
	var transactionBytes []byte

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

	// prepare the passed AccountID
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
	response, _, transactionBytes, err := contract.CallFunction("isOperator", params, args.Gas)
	// fmt.Println("Response:", response)
	// fmt.Println("Receipt:", receipt)
	// fmt.Println("Error:", err)
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
	reply.Value = functionResult.GetBool(0)
	reply.Message = "IsOperator function was called with transaction: " + response.TransactionID.String() + "\n\n" + fmt.Sprintf("Result: %v", reply.Value)
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// Method: getOperatorLastActivity
// 			- check if the given operator is registered in the Renderhive Smart Contract
// #############################################################################

// Method
func (ops *ContractService) GetOperatorLastActivity(r *http.Request, args *GetOperatorLastActivityArgs, reply *GetOperatorLastActivityReply) error {
	var err error
	var transactionBytes []byte

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

	// prepare the passed AccountID
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
	response, _, transactionBytes, err := contract.CallFunction("getOperatorLastActivity", params, args.Gas)
	// fmt.Println("Response:", response)
	// fmt.Println("Receipt:", receipt)
	// fmt.Println("Error:", err)
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

	// set a reply value
	reply.Value = new(big.Int).SetBytes(functionResult.GetInt256(0))

	// set a reply message
	reply.Message = "GetOperatorLastActivity function was called with transaction: " + response.TransactionID.String() + "\n\n" + fmt.Sprintf("Result: %v", reply.Value)
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// RENDERHIVE SMART CONTRACT – NODE MANAGEMENT
// #############################################################################

// Method: addNode
// 			- add a new node for an operator in the Renderhive Smart Contract
// #############################################################################

// Method
func (ops *ContractService) AddNode(r *http.Request, args *AddNodeArgs, reply *AddNodeReply) error {
	var err error
	var transactionBytes []byte

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

	// prepare the passed NodeAccountID
	accountID, err := hederasdk.AccountIDFromString(args.NodeAccountID)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// prepare the parameters for the function call
	params, err := hederasdk.NewContractFunctionParameters().AddAddress(accountID.ToSolidityAddress())
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// add the topic ID to the parameters
	params = params.AddString(args.TopicID)

	// call the function
	_, _, transactionBytes, err = contract.CallPayableFunction("addNode", args.NodeStake, params, args.Gas, hedera.TransactionOptions.SetExecute(false, node.Manager.User.UserAccount.AccountID))
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Sending transaction bytes to frontend for execution with operator wallet"))

	// // get the event log
	// events, err := contract.GetEventLog(response, "AddedNode")
	// if err != nil {
	// 	return fmt.Errorf("Error: %v", err)
	// }

	// // convert event values to usable types
	// fmt.Println("Events:", events)
	// callingAddress, _ := hederasdk.AccountIDFromSolidityAddress(events[0][0].(common.Address).Hex()[2:])
	// nodeAddress, _ := hederasdk.AccountIDFromSolidityAddress(events[0][1].(common.Address).Hex()[2:])
	// nodeTopic := events[0][2].(string)
	// RegistrationTime := time.Unix(events[0][3].(*big.Int).Int64(), 0)

	// // log info
	// logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Contract Event Log: 'Added Node: %v, %v, %v, %v'", callingAddress.String(), nodeAddress.String(), nodeTopic, RegistrationTime.String()))

	// set a reply message
	reply.Message = "" //"addNode function was called with transaction: " + response.TransactionID.String()
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// Method: removeNode
// 			- remove a node of an operator from the Renderhive Smart Contract
// #############################################################################

// Method
func (ops *ContractService) RemoveNode(r *http.Request, args *RemoveNodeArgs, reply *RemoveNodeReply) error {
	var err error
	var transactionBytes []byte

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

	// prepare the passed NodeAccountID
	accountID, err := hederasdk.AccountIDFromString(args.NodeAccountID)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// prepare the parameters for the function call
	params, err := hederasdk.NewContractFunctionParameters().AddAddress(accountID.ToSolidityAddress())
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// call the function
	_, _, transactionBytes, err = contract.CallFunction("removeNode", params, args.Gas, hedera.TransactionOptions.SetExecute(false, node.Manager.User.UserAccount.AccountID))
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Sending transaction bytes to frontend for execution with operator wallet"))

	// // get the event log
	// events, err := contract.GetEventLog(response, "RemovedNode")
	// if err != nil {
	// 	return fmt.Errorf("Error: %v", err)
	// }

	// // convert event values to usable types
	// fmt.Println("Events:", events)
	// callingAddress, _ := hederasdk.AccountIDFromSolidityAddress(events[0][0].(common.Address).Hex()[2:])
	// nodeAddress, _ := hederasdk.AccountIDFromSolidityAddress(events[0][1].(common.Address).Hex()[2:])
	// nodeTopic := events[0][2].(string)
	// DeletionTime := time.Unix(events[0][3].(*big.Int).Int64(), 0)

	// // log info
	// logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Contract Event Log: 'Removed Node: %v, %v, %v, %v'", callingAddress.String(), nodeAddress.String(), nodeTopic, DeletionTime.String()))

	// set a reply message
	reply.Message = "" //"removeNode function was called with transaction: " + response.TransactionID.String()
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// Method: isNode
// 			- check if the given node is registered in the Renderhive Smart Contract
// #############################################################################

// Method
func (ops *ContractService) IsNode(r *http.Request, args *IsNodeArgs, reply *IsNodeReply) error {
	var err error
	var transactionBytes []byte

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

	// prepare the passed AccountIDs
	operatorAccountID, err := hederasdk.AccountIDFromString(args.OperatorAccountID)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	nodeAccountID, err := hederasdk.AccountIDFromString(args.NodeAccountID)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// prepare the parameters for the function call
	params := hederasdk.NewContractFunctionParameters()

	// add operator account ID to the parameters
	params, err = params.AddAddress(operatorAccountID.ToSolidityAddress())
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// add node account ID to the parameters
	params, err = params.AddAddress(nodeAccountID.ToSolidityAddress())
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// call the function
	response, _, transactionBytes, err := contract.CallFunction("isNode", params, args.Gas)
	// fmt.Println("Response:", response)
	// fmt.Println("Receipt:", receipt)
	// fmt.Println("Error:", err)
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
	reply.Value = functionResult.GetBool(0)
	reply.Message = "IsNode function was called with transaction: " + response.TransactionID.String() + "\n\n" + fmt.Sprintf("Result: %v", reply.Value)
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// Method: depositNodeStake
// 			- deposit HBAR for an operator in the Renderhive Smart Contract
// #############################################################################

// Method
func (ops *ContractService) DepositNodeStake(r *http.Request, args *DepositNodeStakeArgs, reply *DepositNodeStakeReply) error {
	var err error
	var transactionBytes []byte

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

	// prepare the passed NodeAccountID
	nodeAccountID, err := hederasdk.AccountIDFromString(args.NodeAccountID)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// prepare the parameters for the function call
	params := hederasdk.NewContractFunctionParameters()

	// add node account ID to the parameters
	params, err = params.AddAddress(nodeAccountID.ToSolidityAddress())
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// call the payable function
	_, _, transactionBytes, err = contract.CallPayableFunction("depositNodeStake", args.NodeStake, params, args.Gas, hedera.TransactionOptions.SetExecute(false, node.Manager.User.UserAccount.AccountID))
	// fmt.Println("Response:", response)
	// fmt.Println("Receipt:", receipt)
	if err != nil {

		// fmt.Println("Error (%v): %v", err, "No details available")
		return fmt.Errorf("Error (%v): %v", err, "No details available")
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Sending transaction bytes to frontend for execution with operator wallet"))

	// set a reply message
	reply.Message = "" //"depositNodeStake function was called with transaction: " + response.TransactionID.String()
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// Method: withdrawNodeStake
// 			- withdraw the complete node stake from the Renderhive Smart Contract
// #############################################################################

// Method
func (ops *ContractService) WithdrawNodeStake(r *http.Request, args *WithdrawNodeStakeArgs, reply *WithdrawNodeStakeReply) error {
	var err error
	var transactionBytes []byte

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

	// prepare the passed NodeAccountID
	nodeAccountID, err := hederasdk.AccountIDFromString(args.NodeAccountID)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// prepare the parameters for the function call
	params := hederasdk.NewContractFunctionParameters()

	// add node account ID to the parameters
	params, err = params.AddAddress(nodeAccountID.ToSolidityAddress())
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	// call the payable function
	_, _, transactionBytes, err = contract.CallFunction("withdrawNodeStake", params, args.Gas, hedera.TransactionOptions.SetExecute(false, node.Manager.User.UserAccount.AccountID))
	// fmt.Println("Response:", response)
	// fmt.Println("Receipt:", receipt)
	// fmt.Println("Error:", err)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Sending transaction bytes to frontend for execution with operator wallet"))

	// set a reply message
	reply.Message = "" //"withdrawNodeStake function was called with transaction: " + response.TransactionID.String()
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// Method: getNodeStake
// 			- get the node stake of the Renderhive Smart Contract
// #############################################################################

// Method
func (ops *ContractService) GetNodeStake(r *http.Request, args *GetNodeStakeArgs, reply *GetNodeStakeReply) error {
	var err error
	var transactionBytes []byte

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

	// prepare the passed NodeAccountID
	nodeAccountID, err := hederasdk.AccountIDFromString(args.NodeAccountID)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// prepare the parameters for the function call
	params := hederasdk.NewContractFunctionParameters()

	// add node account ID to the parameters
	params, err = params.AddAddress(nodeAccountID.ToSolidityAddress())
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	// call the payable function
	response, _, transactionBytes, err := contract.CallFunction("getNodeStake", params, args.Gas)
	// fmt.Println("Response:", response)
	// fmt.Println("Receipt:", receipt)
	// fmt.Println("Error:", err)
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

	// set a reply value
	reply.Value = new(big.Int).SetBytes(functionResult.GetInt256(0))

	// convert tℏ to ℏ
	amount, err := hederasdk.HbarFromString(reply.Value.String() + " tℏ")
	if err != nil {
		return fmt.Errorf("Error getting contract response record: %v", err)
	}

	// set a reply message
	reply.Message = "getNodeStake function was called with transaction: " + response.TransactionID.String() + "\n\n" + fmt.Sprintf("Result: %v", amount)
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// RENDERHIVE SMART CONTRACT – RENDER JOB MANAGEMENT
// #############################################################################

// Method: addRenderJob
// 			- add a new render job to the Renderhive Smart Contract
// #############################################################################

// Method
func (ops *ContractService) AddRenderJob(r *http.Request, args *AddRenderJobArgs, reply *AddRenderJobReply) error {
	var err error
	var transactionBytes []byte

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
	params := hederasdk.NewContractFunctionParameters().AddString(args.JobCID)
	params = params.AddUint256BigInt(new(big.Int).SetUint64(args.Work))

	// call the function
	_, _, transactionBytes, err = contract.CallPayableFunction("addRenderJob", args.Funding, params, args.Gas, hedera.TransactionOptions.SetExecute(false, node.Manager.User.UserAccount.AccountID))
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Sending transaction bytes to frontend for execution with operator wallet"))

	// set a reply message
	reply.Message = "" //"addRenderJob function was called with transaction: " + response.TransactionID.String()
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// Method: claimRenderJob
// 			- claim a render job in the Renderhive Smart Contract
// #############################################################################

// Method
func (ops *ContractService) ClaimRenderJob(r *http.Request, args *ClaimRenderJobArgs, reply *ClaimRenderJobReply) error {
	var err error
	var transactionBytes []byte

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

	// convert consensus root string from hex encoded string (0x15645...) to [32]bytes
	var consensusRoot [32]byte
	var jobRoot [32]byte

	_consensusRoot, err := hex.DecodeString(args.ConsensusRoot[2:])
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	_jobRoot, err := hex.DecodeString(args.JobRoot[2:])
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	if len(_consensusRoot) != 32 || len(_jobRoot) != 32 {
		return fmt.Errorf("Error: %v", "_consensusRoot and _jobRoot must be 32 bytes long")
	} else {
		// convert to [32]byte
		copy(consensusRoot[:], _consensusRoot)
		copy(jobRoot[:], _jobRoot)

	}

	// prepare the parameters for the function call
	params := hederasdk.NewContractFunctionParameters().AddString(args.JobCID)
	params = params.AddUint256BigInt(new(big.Int).SetUint64(args.HiveCycle))
	params = params.AddUint8(args.NodeCount)
	params = params.AddUint128BigInt(new(big.Int).SetUint64(args.NodeShare))
	params = params.AddBytes32(consensusRoot)
	params = params.AddBytes32(jobRoot)

	// call the function
	response, _, transactionBytes, err := contract.CallFunction("claimRenderJob", params, args.Gas)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Contract function called with transaction: %v", response.TransactionID.String()))

	// // get the event log
	// events, err := contract.GetEventLog(response, "AddedNode")
	// if err != nil {
	// 	return fmt.Errorf("Error: %v", err)
	// }

	// // convert event values to usable types
	// fmt.Println("Events:", events)
	// callingAddress, _ := hederasdk.AccountIDFromSolidityAddress(events[0][0].(common.Address).Hex()[2:])
	// nodeAddress, _ := hederasdk.AccountIDFromSolidityAddress(events[0][1].(common.Address).Hex()[2:])
	// nodeTopic := events[0][2].(string)
	// RegistrationTime := time.Unix(events[0][3].(*big.Int).Int64(), 0)

	// // log info
	// logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Contract Event Log: 'Added Node: %v, %v, %v, %v'", callingAddress.String(), nodeAddress.String(), nodeTopic, RegistrationTime.String()))

	// set a reply message
	reply.Message = "claimRenderJob function was called with transaction: " + response.TransactionID.String()
	reply.TransactionBytes = hex.EncodeToString(transactionBytes)

	// create reply for the RPC client
	return nil

}

// INTERNAL HELPER FUNCTIONS
// #############################################################################

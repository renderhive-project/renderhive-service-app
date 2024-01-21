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

	"fmt"
	"math/big"
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

// HEDERA SMART CONTRACT – GENERAL FUNCTIONS
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

// Method
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

// RENDERHIVE SMART CONTRACT – OPERATOR MANAGEMENT
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

// Method
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
	reply.Message = "registerOperator function was called with transaction: " + response.TransactionID.String()

	// create reply for the RPC client
	return nil

}

// Method: unregisterOperator
// 			- unregister an operator from the Renderhive Smart Contract
// #############################################################################

// Arguments and reply
type UnregisterOperatorArgs struct {
	ContractID string // the ID of the smart contract
	Gas        uint64 // the gas limit for the transaction
}
type UnregisterOperatorReply struct {
	Message string
}

// Method
func (ops *ContractService) UnregisterOperator(r *http.Request, args *UnregisterOperatorArgs, reply *UnregisterOperatorReply) error {
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
	params := hederasdk.NewContractFunctionParameters()
	fmt.Println("Params:", contract.ID.String())
	fmt.Println("Params:", params)

	// call the function
	response, _, err := contract.CallFunction("unregisterOperator", params, args.Gas)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Contract function called with transaction: %v", response.TransactionID.String()))

	// set a reply message
	reply.Message = "unregisterOperator function was called with transaction: " + response.TransactionID.String()

	// create reply for the RPC client
	return nil

}

// Method: depositOperatorFunds
// 			- deposit HBAR for an operator in the Renderhive Smart Contract
// #############################################################################

// Arguments and reply
type DepositOperatorFundsArgs struct {
	ContractID string // the ID of the smart contract
	Amount     string // the amount of HBAR to deposit
	Gas        uint64 // the gas limit for the transaction
}
type DepositOperatorFundsReply struct {
	Message string
}

// Method
func (ops *ContractService) DepositOperatorFunds(r *http.Request, args *DepositOperatorFundsArgs, reply *DepositOperatorFundsReply) error {
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
	params := hederasdk.NewContractFunctionParameters()
	fmt.Println("Params:", contract.ID.String())
	fmt.Println("Params:", params)

	// call the payable function
	response, receipt, err := contract.CallPayableFunction("depositOperatorFunds", args.Amount, params, args.Gas)
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
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Contract function called with transaction: %v", response.TransactionID.String()))

	// set a reply message
	reply.Message = "DepositOperatorFunds function was called with transaction: " + response.TransactionID.String()

	// create reply for the RPC client
	return nil

}

// Method: withdrawOperatorFunds
// 			- withdraw HBAR from the Renderhive Smart Contract
// #############################################################################

// Arguments and reply
type WithdrawOperatorFundsArgs struct {
	ContractID string // the ID of the smart contract
	Amount     string // the amount of HBAR to withdraw
	Gas        uint64 // the gas limit for the transaction
}
type WithdrawOperatorFundsReply struct {
	Message string
}

// Method
func (ops *ContractService) WithdrawOperatorFunds(r *http.Request, args *WithdrawOperatorFundsArgs, reply *WithdrawOperatorFundsReply) error {
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
	response, receipt, err := contract.CallFunction("withdrawOperatorFunds", params, args.Gas)
	fmt.Println("Response:", response)
	fmt.Println("Receipt:", receipt)
	fmt.Println("Error:", err)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Contract function called with transaction: %v", response.TransactionID.String()))

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
	reply.Message = "WithdrawOperatorFunds function was called with transaction: " + response.TransactionID.String()

	// create reply for the RPC client
	return nil

}

// Method: getOperatorBalance
// 			- check the balance of an registered operator in the Renderhive Smart Contract
// #############################################################################

// Arguments and reply
type GetOperatorBalanceArgs struct {
	ContractID string // the ID of the smart contract
	AccountID  string // the ID of the account to query for
	Gas        uint64 // the gas limit for the transaction
}
type GetOperatorBalanceReply struct {
	Message string
	Value   *big.Int
}

// Method
func (ops *ContractService) GetOperatorBalance(r *http.Request, args *GetOperatorBalanceArgs, reply *GetOperatorBalanceReply) error {
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
	response, receipt, err := contract.CallFunction("getOperatorBalance", params, args.Gas)
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
	reply.Value = new(big.Int).SetBytes(functionResult.GetInt256(0))

	amount, err := hederasdk.HbarFromString(reply.Value.String() + " tℏ")

	if err != nil {
		return fmt.Errorf("Error getting contract response record: %v", err)
	}
	reply.Message = "getOperatorBalance function was called with transaction: " + response.TransactionID.String() + "\n\n" + fmt.Sprintf("Result: %v", amount)

	// create reply for the RPC client
	return nil

}

// Method: getReservedOperatorFunds
// 			- check the reserved funds of an registered operator in the Renderhive Smart Contract
// #############################################################################

// Arguments and reply
type GetReservedOperatorFundsArgs struct {
	ContractID string // the ID of the smart contract
	AccountID  string // the ID of the account to query for
	Gas        uint64 // the gas limit for the transaction
}
type GetReservedOperatorFundsReply struct {
	Message string
	Value   *big.Int
}

// Method
func (ops *ContractService) GetReservedOperatorFunds(r *http.Request, args *GetReservedOperatorFundsArgs, reply *GetReservedOperatorFundsReply) error {
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
	response, receipt, err := contract.CallFunction("getReservedOperatorFunds", params, args.Gas)
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
	reply.Value = new(big.Int).SetBytes(functionResult.GetInt256(0))

	amount, err := hederasdk.HbarFromString(reply.Value.String() + " tℏ")

	if err != nil {
		return fmt.Errorf("Error getting contract response record: %v", err)
	}
	reply.Message = "getReservedOperatorFunds function was called with transaction: " + response.TransactionID.String() + "\n\n" + fmt.Sprintf("Result: %v", amount)

	// create reply for the RPC client
	return nil

}

// Method: isOperator
// 			- check if the given operator is registered in the Renderhive Smart Contract
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

// Method
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
	reply.Value = functionResult.GetBool(0)
	reply.Message = "IsOperator function was called with transaction: " + response.TransactionID.String() + "\n\n" + fmt.Sprintf("Result: %v", reply.Value)

	// create reply for the RPC client
	return nil

}

// RENDERHIVE SMART CONTRACT – NODE MANAGEMENT
// #############################################################################

// Method: addNode
// 			- add a new node for an operator in the Renderhive Smart Contract
// #############################################################################

// Arguments and reply
type AddNodeArgs struct {
	// PublicKey       *ecdsa.PublicKey // the public key of the node
	ContractID string // the ID of the smart contract
	AccountID  string // the AccountID of the node to be added
	TopicID    string // the TopicID of the nodes's HCS topic
	NodeStake  string // the amount of HBAR to deposit as node stake

	Gas uint64 // the gas limit for the transaction
}
type AddNodeReply struct {
	Message string
}

// Method
func (ops *ContractService) AddNode(r *http.Request, args *AddNodeArgs, reply *AddNodeReply) error {
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

	// add the topic ID to the parameters
	params = params.AddString(args.TopicID)

	// call the function
	response, _, err := contract.CallPayableFunction("addNode", args.NodeStake, params, args.Gas)
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
	reply.Message = "addNode function was called with transaction: " + response.TransactionID.String()

	// create reply for the RPC client
	return nil

}

// Method: removeNode
// 			- remove a node of an operator from the Renderhive Smart Contract
// #############################################################################

// Arguments and reply
type RemoveNodeArgs struct {
	ContractID string // the ID of the smart contract
	AccountID  string // the AccountID of the node to be deleted

	Gas uint64 // the gas limit for the transaction
}
type RemoveNodeReply struct {
	Message string
}

// Method
func (ops *ContractService) RemoveNode(r *http.Request, args *RemoveNodeArgs, reply *RemoveNodeReply) error {
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
	response, _, err := contract.CallFunction("removeNode", params, args.Gas)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Contract function called with transaction: %v", response.TransactionID.String()))

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
	reply.Message = "removeNode function was called with transaction: " + response.TransactionID.String()

	// create reply for the RPC client
	return nil

}

// Method: isNode
// 			- check if the given node is registered in the Renderhive Smart Contract
// #############################################################################

// Arguments and reply
type IsNodeArgs struct {
	ContractID        string // the ID of the smart contract
	NodeAccountID     string // the account ID of the node to query for
	OperatorAccountID string // the account ID of the operator owning the node

	Gas uint64 // the gas limit for the transaction
}
type IsNodeReply struct {
	Message string
	Value   bool
}

// Method
func (ops *ContractService) IsNode(r *http.Request, args *IsNodeArgs, reply *IsNodeReply) error {
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
	response, receipt, err := contract.CallFunction("isNode", params, args.Gas)
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
	reply.Value = functionResult.GetBool(0)
	reply.Message = "IsNode function was called with transaction: " + response.TransactionID.String() + "\n\n" + fmt.Sprintf("Result: %v", reply.Value)

	// create reply for the RPC client
	return nil

}

// Method: depositNodeStake
// 			- deposit HBAR for an operator in the Renderhive Smart Contract
// #############################################################################

// Arguments and reply
type DepositNodeStakeArgs struct {
	ContractID    string // the ID of the smart contract
	NodeAccountID string // the account ID of the node to deposit for
	NodeStake     string // the amount of HBAR to deposit as node stake

	Gas uint64 // the gas limit for the transaction
}
type DepositNodeStakeReply struct {
	Message string
}

// Method
func (ops *ContractService) DepositNodeStake(r *http.Request, args *DepositNodeStakeArgs, reply *DepositNodeStakeReply) error {
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

	// prepare the passed AccountID
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
	response, receipt, err := contract.CallPayableFunction("depositNodeStake", args.NodeStake, params, args.Gas)
	fmt.Println("Response:", response)
	fmt.Println("Receipt:", receipt)
	if err != nil {

		// fmt.Println("Error (%v): %v", err, "No details available")
		return fmt.Errorf("Error (%v): %v", err, "No details available")
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Contract function called with transaction: %v", response.TransactionID.String()))

	// set a reply message
	reply.Message = "depositNodeStake function was called with transaction: " + response.TransactionID.String()

	// create reply for the RPC client
	return nil

}

// Method: withdrawNodeStake
// 			- withdraw the complete node stake from the Renderhive Smart Contract
// #############################################################################

// Arguments and reply
type WithdrawNodeStakeArgs struct {
	ContractID    string // the ID of the smart contract
	NodeAccountID string // the account ID of the node to withdraw the stake from

	Gas uint64 // the gas limit for the transaction
}
type WithdrawNodeStakeReply struct {
	Message string
}

// Method
func (ops *ContractService) WithdrawNodeStake(r *http.Request, args *WithdrawNodeStakeArgs, reply *WithdrawNodeStakeReply) error {
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

	// prepare the passed AccountID
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
	response, receipt, err := contract.CallFunction("withdrawNodeStake", params, args.Gas)
	fmt.Println("Response:", response)
	fmt.Println("Receipt:", receipt)
	fmt.Println("Error:", err)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// log info
	logger.Manager.Package["jsonrpc"].Info().Msg(fmt.Sprintf(" [#] Contract function called with transaction: %v", response.TransactionID.String()))

	// set a reply message
	reply.Message = "withdrawNodeStake function was called with transaction: " + response.TransactionID.String()

	// create reply for the RPC client
	return nil

}

// INTERNAL HELPER FUNCTIONS
// #############################################################################

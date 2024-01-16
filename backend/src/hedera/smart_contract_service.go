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

package hedera

import (

	// standard
	"encoding/json"
	"fmt"
	"os"

	//"errors"

	// external
	"github.com/ethereum/go-ethereum/accounts/abi"
	hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

	// internal
	"renderhive/logger"
)

// Hedera Smart Contract data
type HederaSmartContractData struct {
	Data struct {
		Bytecode struct {
			Object    string `json:"object"`
			OpCodes   string `json:"opcodes"`
			SourceMap string `json:"sourceMap"`
		}
	}
}

// Hedera smart contract service contract
type HederaSmartContract struct {
	ID   hederasdk.ContractID
	Info hederasdk.ContractInfo
	Data HederaSmartContractData
}

// SMART CONTRACT MANAGEMENT - HELPER FUNCTIONS
// #############################################################################
// decodeEvent decodes event data from a Solidity contract
func decodeEvent(eventName string, log []byte, topics [][]byte) ([]interface{}, error) {
	var contractFilePath string

	contractFilePath = "./RenderhiveContract.abi"

	// Import and parse the compiled contract from the contract file
	jsonData, err := os.ReadFile(contractFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading contract file %q: %s", contractFilePath, err)
	}

	// Parse the ABI from the JSON string
	var parsedAbi abi.ABI
	err = parsedAbi.UnmarshalJSON([]byte(jsonData))
	if err != nil {
		return nil, err
	}

	// Iterate over the events in the ABI and print their names
	for _, event := range parsedAbi.Events {
		if event.Name == eventName {

			// Decode the log data using the ABI
			decodedLog, err := event.Inputs.UnpackValues(log)
			if err != nil {
				return nil, err
			}

			return decodedLog, nil

		}
	}
	return nil, fmt.Errorf("event not found: %s", eventName)

}

// SMART CONTRACT MANAGEMENT
// #############################################################################
// This function reads in a contract JSON file, creates a new contract with the contract.Object field as the bytecode,
// deploys it on the Hedera network, and returns transaction receipt.
func (contract *HederaSmartContract) NewFromJSON(contractFilePath string, adminKey interface{}, gas int64) (*hederasdk.TransactionResponse, *hederasdk.TransactionReceipt, error) {
	var err error

	// Import and parse the compiled contract from the contract file
	jsonData, err := os.ReadFile(contractFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading contract file %q: %s", contractFilePath, err)
	}

	var contractData HederaSmartContractData = HederaSmartContractData{}
	err = json.Unmarshal([]byte(jsonData), &contractData)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshaling the json file %q: %s", contractFilePath, err)
	}

	// read the bytecode
	bytecode := []byte(contractData.Data.Bytecode.Object)

	// create the transaction to deploy the contract bytecode on the Hedera network
	newContractCreateFlowTransaction := hederasdk.NewContractCreateFlow().
		SetBytecode(bytecode).
		SetGas(gas)

	// if a contract memo was passed
	if contract.Info.ContractMemo != "" {
		newContractCreateFlowTransaction = newContractCreateFlowTransaction.SetContractMemo(contract.Info.ContractMemo)
	}

	// if a admin key was passed
	if contract.Info.AdminKey != nil {

		// set it in the transaction
		newContractCreateFlowTransaction = newContractCreateFlowTransaction.SetAdminKey(contract.Info.AdminKey)

		// // Freeze the transaction for signing (this prevents the transaction can be
		// // modified while signing it)
		// newContractCreateFlowTransaction, err := newContractCreateFlowTransaction.FreezeWith(Manager.NetworkClient)
		// if err != nil {
		//     return nil, err
		// }

		// // if the type of the passed key is a PrivateKey
		// thisKey, ok := adminKey.(hederasdk.PrivateKey)
		// if ok == true {
		//
		//     logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing transaction with private key of: %s", thisKey.PublicKey()))
		//
		//     // and sign the transaction with this key
		//     newContractCreateFlowTransaction = newContractCreateFlowTransaction.Sign(thisKey)
		//
		// }
		//
		// // if the type of the passed key is a slice of PrivateKey
		// keyList, ok := adminKey.([]hederasdk.PrivateKey)
		// if ok == true {
		//
		//     // iterate through all keys
		//     for i, thisKey := range keyList {
		//
		//       logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing transaction with private key (%v) of: %s", i, thisKey.PublicKey()))
		//
		//       // and sign the transaction with each key
		//       newContractCreateFlowTransaction = newContractCreateFlowTransaction.Sign(thisKey)
		//
		//     }
		//
		// }
	}

	// sign with client operator private key and submit the query to a Hedera network
	transactionResponse, err := newContractCreateFlowTransaction.Execute(Manager.NetworkClient)
	if err != nil {
		return nil, nil, err
	}

	// get the transaction receipt
	transactionReceipt, err := transactionResponse.GetReceipt(Manager.NetworkClient)
	if err != nil {
		return &transactionResponse, nil, err
	}

	// log the receipt status of the transaction
	logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", transactionReceipt.TransactionID.String(), transactionReceipt.Status))

	// get the contract ID from the transaction receipt
	contract.ID = *transactionReceipt.ContractID

	// Return the contract ID
	return &transactionResponse, &transactionReceipt, err

}

// This function reads in a contracts bytecode file, creates a new contract,
// deploys it on the Hedera network, and returns transaction receipt.
func (contract *HederaSmartContract) NewFromBin(contractFilePath string, adminKey interface{}, gas int64) (*hederasdk.TransactionResponse, *hederasdk.TransactionReceipt, error) {
	var err error

	// Import and parse the compiled contract from the contract file
	bytecode, err := os.ReadFile(contractFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading contract file %q: %s", contractFilePath, err)
	}

	// create the transaction to deploy the contract bytecode on the Hedera network
	newContractCreateFlowTransaction := hederasdk.NewContractCreateFlow().
		SetBytecode(bytecode).
		SetGas(gas).
		SetMaxChunks(30)

	// if a contract memo was passed
	if contract.Info.ContractMemo != "" {
		newContractCreateFlowTransaction = newContractCreateFlowTransaction.SetContractMemo(contract.Info.ContractMemo)
	}

	// if a admin key was passed
	if contract.Info.AdminKey != nil {

		// set it in the transaction
		newContractCreateFlowTransaction = newContractCreateFlowTransaction.SetAdminKey(contract.Info.AdminKey)

		// // Freeze the transaction for signing (this prevents the transaction can be
		// // modified while signing it)
		// newContractCreateFlowTransaction, err := newContractCreateFlowTransaction.FreezeWith(Manager.NetworkClient)
		// if err != nil {
		//     return nil, err
		// }

		// // if the type of the passed key is a PrivateKey
		// thisKey, ok := adminKey.(hederasdk.PrivateKey)
		// if ok == true {
		//
		//     logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing transaction with private key of: %s", thisKey.PublicKey()))
		//
		//     // and sign the transaction with this key
		//     newContractCreateFlowTransaction = newContractCreateFlowTransaction.Sign(thisKey)
		//
		// }
		//
		// // if the type of the passed key is a slice of PrivateKey
		// keyList, ok := adminKey.([]hederasdk.PrivateKey)
		// if ok == true {
		//
		//     // iterate through all keys
		//     for i, thisKey := range keyList {
		//
		//       logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing transaction with private key (%v) of: %s", i, thisKey.PublicKey()))
		//
		//       // and sign the transaction with each key
		//       newContractCreateFlowTransaction = newContractCreateFlowTransaction.Sign(thisKey)
		//
		//     }
		//
		// }
	}

	// sign with client operator private key and submit the query to a Hedera network
	transactionResponse, err := newContractCreateFlowTransaction.Execute(Manager.NetworkClient)
	if err != nil {
		return nil, nil, err
	}

	// get the transaction receipt
	transactionReceipt, err := transactionResponse.GetReceipt(Manager.NetworkClient)
	if err != nil {
		return &transactionResponse, nil, err
	}

	// log the receipt status of the transaction
	logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", transactionReceipt.TransactionID.String(), transactionReceipt.Status))

	// get the contract ID from the transaction receipt
	contract.ID = *transactionReceipt.ContractID

	// Return the contract ID
	return &transactionResponse, &transactionReceipt, err

}

// Delete the contract
func (contract *HederaSmartContract) Delete(adminKey interface{}) (*hederasdk.TransactionResponse, *hederasdk.TransactionReceipt, error) {
	var err error

	// delete the topic
	newContractDeleteTransaction := hederasdk.NewContractDeleteTransaction().
		SetContractID(contract.ID).
		SetTransferAccountID(Manager.Operator.AccountID)

	// if the topic has a AdminKey
	if contract.Info.AdminKey != nil {

		// Freeze the transaction for signing (this prevents the transaction can be
		// modified while signing it)
		newContractDeleteTransaction, err := newContractDeleteTransaction.FreezeWith(Manager.NetworkClient)
		if err != nil {
			return nil, nil, err
		}

		// if the type of the passed key is a PrivateKey
		thisKey, ok := adminKey.(hederasdk.PrivateKey)
		if ok == true {

			logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing transaction with private key of: %s", thisKey.PublicKey()))

			// and sign the transaction with this key
			newContractDeleteTransaction = newContractDeleteTransaction.Sign(thisKey)

		}

		// if the type of the passed key is a slice of PrivateKey
		keyList, ok := adminKey.([]hederasdk.PrivateKey)
		if ok == true {

			// iterate through all keys
			for i, thisKey := range keyList {

				logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing transaction with private key (%v) of: %s", i, thisKey.PublicKey()))

				// and sign the transaction with each key
				newContractDeleteTransaction = newContractDeleteTransaction.Sign(thisKey)

			}

		}

	}

	// sign with client operator private key and submit the query to a Hedera network
	// NOTE: This will only delete the contract, if the operator account's key was set
	//       as admin key
	transactionResponse, err := newContractDeleteTransaction.Execute(Manager.NetworkClient)
	if err != nil {
		return nil, nil, err
	}

	// get the transaction receipt
	transactionReceipt, err := transactionResponse.GetReceipt(Manager.NetworkClient)
	if err != nil {
		return &transactionResponse, nil, err
	}

	// log the receipt status of the transaction
	logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", transactionReceipt.TransactionID.String(), transactionReceipt.Status))

	// free the pointer
	contract = nil

	return &transactionResponse, &transactionReceipt, err

}

// Call a smart contract function
func (contract *HederaSmartContract) CallFunction(name string, parameters *hederasdk.ContractFunctionParameters, gas uint64) (*hederasdk.TransactionResponse, *hederasdk.TransactionReceipt, error) {
	var err error

	// create the cmart contract call
	newContractExecuteTransaction := hederasdk.NewContractExecuteTransaction().
		SetContractID(contract.ID).
		SetGas(gas).
		SetFunction(name, parameters)

	// get the transaction response
	transactionResponse, err := newContractExecuteTransaction.Execute(Manager.NetworkClient)
	if err != nil {
		return &transactionResponse, nil, err
	}

	// get the transaction receipt
	transactionReceipt, err := transactionResponse.GetReceipt(Manager.NetworkClient)
	if err != nil {
		return &transactionResponse, &transactionReceipt, err
	}

	return &transactionResponse, &transactionReceipt, err

}

// Call a smart contract function
func (contract *HederaSmartContract) CallPayableFunction(name string, amount string, parameters *hederasdk.ContractFunctionParameters, gas uint64) (*hederasdk.TransactionResponse, *hederasdk.TransactionReceipt, error) {
	var err error

	// get the amount as HBAR
	_amount, err := hederasdk.HbarFromString(amount)
	if err != nil {
		return nil, nil, err
	}

	// create the cmart contract call
	newContractExecuteTransaction := hederasdk.NewContractExecuteTransaction().
		SetContractID(contract.ID).
		SetGas(gas).
		SetFunction(name, parameters).
		SetPayableAmount(_amount)

	// get the transaction response
	transactionResponse, err := newContractExecuteTransaction.Execute(Manager.NetworkClient)
	if err != nil {
		return &transactionResponse, nil, err
	}

	// get the transaction receipt
	transactionReceipt, err := transactionResponse.GetReceipt(Manager.NetworkClient)
	if err != nil {
		return &transactionResponse, &transactionReceipt, err
	}

	return &transactionResponse, &transactionReceipt, err

}

// Call a smart contract function local (i.e., on a single node)
func (contract *HederaSmartContract) CallFunctionLocal(name string, parameters *hederasdk.ContractFunctionParameters, gas uint64) (*hederasdk.ContractFunctionResult, error) {
	var err error

	// create the local smart contract call
	newContractCallQueryTransaction := hederasdk.NewContractCallQuery().
		SetContractID(contract.ID).
		SetGas(gas).
		SetFunction(name, parameters)

	// get the function result
	functionResult, err := newContractCallQueryTransaction.Execute(Manager.NetworkClient)
	if err != nil {
		return nil, err
	}

	return &functionResult, err

}

// Get the events emitted by the contract after a function call
// TODO: Might be good, if the wallet address would be an indexed event parameter
//
//	This would later allow to scan the event history for the user. Useful?
func (contract *HederaSmartContract) GetEventLog(callFunctionResponse *hederasdk.TransactionResponse, eventName string) ([][]interface{}, error) {
	var err error
	var events [][]interface{}

	// get the transaction record
	transactionRecord, err := callFunctionResponse.GetRecord(Manager.NetworkClient)
	if err != nil {
		return nil, err
	}

	// get the contract function result
	contractFunctionResult, err := transactionRecord.GetContractExecuteResult()
	if err != nil {
		return nil, err
	}

	// Iterate over the logs
	for _, log := range contractFunctionResult.LogInfo {
		// Decode the event data: RegisteredUser
		event, err := decodeEvent(eventName, log.Data, log.Topics)
		if err != nil {
			return nil, err
		}

		// Append the event to the slice
		events = append(events, event)

	}

	return events, err

}

// Query the Hedera network for information on the contract
// NOTE: This should be used spareingly, since it has a network fee
func (contract *HederaSmartContract) QueryInfo(m *PackageManager) (string, error) {
	var err error

	// create the topic info query
	newContractInfoQuery := hederasdk.NewContractInfoQuery().
		SetContractID(contract.ID).
		SetMaxQueryPayment(hederasdk.NewHbar(1))

	// get cost of this query
	cost, err := newContractInfoQuery.GetCost(Manager.NetworkClient)
	if err != nil {
		return "", err
	}

	// Sign with client operator private key and submit the query to a Hedera network
	contract.Info, err = newContractInfoQuery.Execute(Manager.NetworkClient)
	if err != nil {
		return "", err
	}

	return cost.String(), nil
}

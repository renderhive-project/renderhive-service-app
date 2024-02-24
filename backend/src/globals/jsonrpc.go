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

package globals

import (
	"math/big"
)

// GLOBALLY REQUIRED DEFINITIONS FOR THE JSON-RPC
// #############################################################################

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
	Message          string
	TransactionBytes string
}

// Method: GetCurrentHiveCycle
// 			- get the current hive cycle from the contract
// #############################################################################

// Arguments and reply
type GetCurrentHiveCycleArgs struct {
	ContractID string
	Gas        uint64
}
type GetCurrentHiveCycleReply struct {
	Message          string
	Value            *big.Int
	TransactionBytes string
}

// RENDERHIVE SMART CONTRACT – OPERATOR MANAGEMENT
// #############################################################################

// Method: registerOperator
// #############################################################################

// Arguments and reply
type RegisterOperatorArgs struct {
	// PublicKey       *ecdsa.PublicKey // the public key of the operator
	ContractID      string // the ID of the smart contract
	OperatorTopicID string // the TopicID of the operator's HCS topic

	Gas uint64 // the gas limit for the transaction
}
type RegisterOperatorReply struct {
	Message          string
	TransactionBytes string
}

// Method: unregisterOperator
// #############################################################################

// Arguments and reply
type UnregisterOperatorArgs struct {
	ContractID string // the ID of the smart contract
	Gas        uint64 // the gas limit for the transaction
}
type UnregisterOperatorReply struct {
	Message          string
	TransactionBytes string
}

// Method: depositOperatorFunds
// #############################################################################

// Arguments and reply
type DepositOperatorFundsArgs struct {
	ContractID string // the ID of the smart contract
	Amount     string // the amount of HBAR to deposit
	Gas        uint64 // the gas limit for the transaction
}
type DepositOperatorFundsReply struct {
	Message          string
	TransactionBytes string
}

// Method: withdrawOperatorFunds
// #############################################################################

// Arguments and reply
type WithdrawOperatorFundsArgs struct {
	ContractID string // the ID of the smart contract
	Amount     string // the amount of HBAR to withdraw
	Gas        uint64 // the gas limit for the transaction
}
type WithdrawOperatorFundsReply struct {
	Message          string
	TransactionBytes string
}

// Method: getOperatorBalance
// #############################################################################

// Arguments and reply
type GetOperatorBalanceArgs struct {
	ContractID string // the ID of the smart contract
	AccountID  string // the ID of the account to query for
	Gas        uint64 // the gas limit for the transaction
}
type GetOperatorBalanceReply struct {
	Message          string
	Value            *big.Int
	TransactionBytes string
}

// Method: getReservedOperatorFunds
// #############################################################################

// Arguments and reply
type GetReservedOperatorFundsArgs struct {
	ContractID string // the ID of the smart contract
	AccountID  string // the ID of the account to query for
	Gas        uint64 // the gas limit for the transaction
}
type GetReservedOperatorFundsReply struct {
	Message          string
	Value            *big.Int
	TransactionBytes string
}

// Method: isOperator
// #############################################################################

// Arguments and reply
type IsOperatorArgs struct {
	ContractID string // the ID of the smart contract
	AccountID  string // the ID of the account to query for
	Gas        uint64 // the gas limit for the transaction
}
type IsOperatorReply struct {
	Message          string
	Value            bool
	TransactionBytes string
}

// Method: getOperatorLastActivity
// #############################################################################

// Arguments and reply
type GetOperatorLastActivityArgs struct {
	ContractID string // the ID of the smart contract
	AccountID  string // the ID of the account to query for
	Gas        uint64 // the gas limit for the transaction
}
type GetOperatorLastActivityReply struct {
	Message          string
	Value            *big.Int
	TransactionBytes string
}

// RENDERHIVE SMART CONTRACT – NODE MANAGEMENT
// #############################################################################

// Method: addNode
// #############################################################################

// Arguments and reply
type AddNodeArgs struct {
	// PublicKey       *ecdsa.PublicKey // the public key of the node
	ContractID    string // the ID of the smart contract
	NodeAccountID string // the AccountID of the node to be added
	TopicID       string // the TopicID of the nodes's HCS topic
	NodeStake     string // the amount of HBAR to deposit as node stake

	Gas uint64 // the gas limit for the transaction
}
type AddNodeReply struct {
	Message          string
	TransactionBytes string
}

// Method: removeNode
// #############################################################################

// Arguments and reply
type RemoveNodeArgs struct {
	ContractID    string // the ID of the smart contract
	NodeAccountID string // the AccountID of the node to be deleted

	Gas uint64 // the gas limit for the transaction
}
type RemoveNodeReply struct {
	Message          string
	TransactionBytes string
}

// Method: isNode
// #############################################################################

// Arguments and reply
type IsNodeArgs struct {
	ContractID        string // the ID of the smart contract
	NodeAccountID     string // the account ID of the node to query for
	OperatorAccountID string // the account ID of the operator owning the node

	Gas uint64 // the gas limit for the transaction
}
type IsNodeReply struct {
	Message          string
	Value            bool
	TransactionBytes string
}

// Method: depositNodeStake
// #############################################################################

// Arguments and reply
type DepositNodeStakeArgs struct {
	ContractID    string // the ID of the smart contract
	NodeAccountID string // the account ID of the node to deposit for
	NodeStake     string // the amount of HBAR to deposit as node stake

	Gas uint64 // the gas limit for the transaction
}
type DepositNodeStakeReply struct {
	Message          string
	TransactionBytes string
}

// Method: withdrawNodeStake
// #############################################################################

// Arguments and reply
type WithdrawNodeStakeArgs struct {
	ContractID    string // the ID of the smart contract
	NodeAccountID string // the account ID of the node to withdraw the stake from

	Gas uint64 // the gas limit for the transaction
}
type WithdrawNodeStakeReply struct {
	Message          string
	TransactionBytes string
}

// Method: getNodeStake
// #############################################################################

// Arguments and reply
type GetNodeStakeArgs struct {
	ContractID    string // the ID of the smart contract
	NodeAccountID string // the account ID of the node to get the stake of

	Gas uint64 // the gas limit for the transaction
}
type GetNodeStakeReply struct {
	Message          string
	Value            *big.Int
	TransactionBytes string
}

// RENDERHIVE SMART CONTRACT – RENDER JOB MANAGEMENT
// #############################################################################

// Method: addRenderJob
// #############################################################################

// Arguments and reply
type AddRenderJobArgs struct {
	ContractID string // the ID of the smart contract
	JobCID     string // the CID of the render job document
	Work       uint64 // the estimated render work in BBh
	Funding    string // the amount of HBAR to deposit as funding for the render job

	Gas uint64 // the gas limit for the transaction
}
type AddRenderJobReply struct {
	Message          string
	TransactionBytes string
}

// Method: claimRenderJob
// #############################################################################

// Arguments and reply
type ClaimRenderJobArgs struct {
	ContractID    string // the ID of the smart contract
	JobCID        string // the CID of the render job document
	HiveCycle     uint64 // the current hive cycle
	NodeCount     uint8  // the number of nodes to claim the job
	NodeShare     uint64 // the share of work to be rendered by this node (in parts per 10,000 of the total work, i.e. 1% = 100 parts per 10,000)
	ConsensusRoot string // the root of the consensus merkle tree for the hive cycle
	JobRoot       string // the root of the job's merkle tree for the hive cycle

	Gas uint64 // the gas limit for the transaction
}
type ClaimRenderJobReply struct {
	Message          string
	TransactionBytes string
}

// RENDERHIVE OPERATOR SERVICE
// #############################################################################

// Defines the operator information
type OperatorInfo struct {
	ID        int    `json:"userid"`     // a unique user id
	Username  string `json:"username"`   // a unique username
	Email     string `json:"email"`      // email address of the user
	AccountID string `json:"accountid"`  // the 0.0.xxxx formated account id
	PublicKey string `json:"public_key"` // public key of the Hedera account
}

// Defines the node information
type NodeInfo struct {
	ID         int    `json:"nodeid"`      // a unique user id
	Name       string `json:"name"`        // a unique alias for this node
	ClientNode bool   `json:"client_node"` // node is a client node
	RenderNode bool   `json:"render_node"` // node is a render node
	AccountID  string `json:"accountid"`   // the 0.0.xxxx formated Hedera account id
	PublicKey  string `json:"public_key"`  // public key of the Hedera account
}

// Method: SignUp
// #############################################################################

// Arguments and reply
type SignUpArgs struct {
	Step                         string
	Operator                     OperatorInfo
	Node                         NodeInfo
	Passphrase                   string
	AccountCreationTransactionID string
}
type SignUpReply struct {
	Message       string
	NodeAccountID string
	Payload       []byte
}

// Method: GetSignInPayload
// #############################################################################

// Arguments and reply
type GetSignInPayloadArgs struct{}
type GetSignInPayloadReply struct {
	Payload []byte
}

// Method: SignIn
// #############################################################################

// Arguments and reply
type SignInArgs struct {
	Passphrase string
}
type SignInReply struct {
	Message  string
	SignedIn bool
	Token    string
}

// Method: SignOut
// #############################################################################

// Arguments and reply
type SignOutArgs struct{}
type SignOutReply struct {
	Message  string
	SignedIn bool
}

// Method: GetInfo
// #############################################################################

// Arguments and reply
type GetInfoArgs struct{}

type GetInfoReply struct {
	// Operator details
	Username    string
	UserEmail   string
	UserAccount string

	// Node Details
	NodeName    string
	NodeAccount string
}

// Method: GetContractInfo
// #############################################################################

// Arguments and reply
type GetContractInfoArgs struct {
	AccountID string
}

type GetContractInfoReply struct {
	// Operator details
	Username    string
	UserEmail   string
	UserAccount string

	// Node Details
	NodeAlias   string
	NodeAccount string
}

// Method: IsSessionValid
// #############################################################################

// Arguments and reply
type IsSessionValidArgs struct{}

type IsSessionValidReply struct {
	Valid bool
}

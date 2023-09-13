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

package webapp

/*

 The operator service enables the webapp to query and request actions regarding operator accounts.
 Operators are the users, which are registered with a wallet address in the smart contract and that
 can manage one or several nodes. While node wallets are maintained by the node and can automatically
 send transactions, operator wallets ALWAYS require user interactions. Therefore, only the operator wallets
 can store and withdraw money in form of tokens in/from the smart contract. In addition, only operator
 wallets can register new nodes or delete them.

*/

import (

	// standard
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	// external
	hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

	// internal
	. "renderhive/globals"
	"renderhive/hedera"
	"renderhive/node"
)

// SERVICE INITIALIZATION
// #############################################################################

// Defines the operator information
type Operator struct {
	Username  string `json:"username"`  // a unique username
	Email     string `json:"email"`     // email address of the user
	AccountID string `json:"accountid"` // the 0.0.xxxx formated account id with checksum
}

var operators = make(map[string]Operator)

// export the OperatorService for net/rpc
type OperatorService struct{}

// Method: SignUp
// 			- register a new operator with the smart contract service
// #############################################################################

// Arguments and reply
type SignUpArgs struct {
	Operator Operator
}
type SignUpReply struct {
	Message string
}

// Adds a known operator
func (ops *OperatorService) SignUp(r *http.Request, args *SignUpArgs, reply *SignUpReply) error {

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// if the operator already exists
	if _, ok := operators[args.Operator.AccountID]; ok {
		return fmt.Errorf("Operator already exists")
	}

	// query the account information from a mirror node
	accounts, err := hedera.Manager.MirrorNode.GetAccountInfo(args.Operator.AccountID, 1, "")
	if err != nil || (accounts == nil || (accounts != nil && len(*accounts) == 0)) {
		return fmt.Errorf("Failed to obtain account information from mirror node: %v", err)
	}

	// TODO: HERE ALL THE LOGIC FOR THE ACTUAL SIGNUP NEEDS TO BE ADDED
	//		 - Register at the smart contract

	// store the operator data in a file, which can be loaded the next time
	data, err := json.MarshalIndent(args.Operator, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal operator: %v", err)
	}
	err = os.WriteFile(filepath.Join(RENDERHIVE_APP_DIRECTORY_CONFIG, "operator.json"), data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	// add the operator as the node operator
	node.Manager.User.Username = args.Operator.Username
	node.Manager.User.Email = args.Operator.Email
	node.Manager.User.UserAccount.AccountID, err = hederasdk.AccountIDFromString(args.Operator.AccountID)
	if err != nil {
		return fmt.Errorf("Failed to obtain account id from string: %v", err)
	}

	// create reply for the client
	reply.Message = "New operator was signed up: " + args.Operator.Username + "!"
	return nil

}

// Method: GetInfo
//			- obtain information about the node operator
// #############################################################################

// Arguments and reply
type GetInfoArgs struct {
	AccountID string // the 0.0.xxxx-xxx formated account id with checksum
}
type GetInfoReply struct {
	Username string
	Email    string
	Account  string
}

// Get info about the operator via the accountid
func (ops *OperatorService) GetInfo(r *http.Request, args *GetInfoArgs, reply *GetInfoReply) error {

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// get the operator from the map
	operator, ok := operators[args.AccountID]
	if !ok {
		return fmt.Errorf("Operator unknown: %v", args.AccountID)
	}

	// create reply for the client
	reply.Username = operator.Username
	reply.Email = operator.Email
	reply.Account = operator.AccountID

	return nil
}

// INTERNAL HELPER FUNCTIONS
// #############################################################################

// Read operator information known to this machine from a file
func (webappm *PackageManager) FromFile(path string) error {

	// read the operator file stored on this machine

	return nil
}

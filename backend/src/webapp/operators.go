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

	"crypto/ed25519"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	// external
	"github.com/golang-jwt/jwt/v5"
	hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

	// internal
	. "renderhive/globals"
	"renderhive/hedera"
	"renderhive/logger"
	"renderhive/node"
	"renderhive/utility"
)

// SERVICE INITIALIZATION
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

// export the OperatorService for net/rpc
type OperatorService struct{}

// Method: SignUp
// 			- register a new operator with the smart contract service
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

// Adds a known operator
func (ops *OperatorService) SignUp(r *http.Request, args *SignUpArgs, reply *SignUpReply) error {

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// STEP 1: Initialize the sign up procedure
	if args.Step == "init" {

		// TODO: Implement further checks and security measures

		// OPERATOR INFO
		// if the node operator info is not yet available
		if node.Manager.User.UserAccount.PublicKey.String() == "" {

			// query the operator account information from a mirror node
			accounts, err := hedera.Manager.MirrorNode.GetAccountInfo(args.Operator.AccountID, 1, "")
			if err != nil || (accounts == nil || (accounts != nil && len(*accounts) == 0)) {
				return fmt.Errorf("Error: %v", err)
			}

			// update the node data
			node.Manager.User.UserAccount.AccountID, err = hederasdk.AccountIDFromString((*accounts)[0].Account)
			if err != nil {
				return fmt.Errorf("Error: %v", err)
			}
			node.Manager.User.UserAccount.PublicKey, err = hederasdk.PublicKeyFromString((*accounts)[0].Key.Key)
			if err != nil {
				return fmt.Errorf("Error: %v", err)
			}

			// write the user data to a file
			err = node.Manager.WriteUserData(args.Operator.ID, args.Operator.Username, args.Operator.Email, args.Operator.AccountID, (*accounts)[0].Key.Key)
			if err != nil {
				return fmt.Errorf("Error: %v", err)
			}
		}

		// NODE INFO
		if args.Node.Name == "" {
			return fmt.Errorf("Error: %v", "Node name is empty")
		}

		// if the node operator info is not yet available
		if node.Manager.Node.HederaAccount.PublicKey == "" {

			// generate a key pair
			privateKey, err := hederasdk.PrivateKeyGenerateEd25519()
			if err != nil {
				return fmt.Errorf("Error: %v", err)
			}
			publicKey := privateKey.PublicKey()

			// Assuming that the target shard and realm are known.
			// For now they are virtually always 0 and 0.
			aliasAccountId := *publicKey.ToAccountID(0, 0)
			if aliasAccountId.AliasKey == nil {
				return fmt.Errorf("Error: %v", "Could not generate alias key for node account")
			}

			// set Hedera operator for the network client
			hedera.Manager.Operator.PublicKey = publicKey
			hedera.Manager.Operator.PrivateKey = privateKey

			// save the node's private key using the users passphrase
			err = hedera.Manager.Operator.ToFile(filepath.Join(RENDERHIVE_APP_DIRECTORY_CONFIG, strings.ReplaceAll(aliasAccountId.String(), ".", "")+".key"), args.Passphrase)
			if err != nil {
				return fmt.Errorf("Failed to save private key: %v", err)
			}

			// create reply for the RPC client
			reply.Message = "Successfully initialized sign up procedure"
			reply.NodeAccountID = aliasAccountId.String()
			reply.Payload = nil

		}

		return nil

	}

	// STEP 2: Create the node account
	if args.Step == "create" {

		logger.Manager.Package["webapp"].Info().Msg(fmt.Sprintf("Waiting for transaction to land on the mirror node: %v", args.AccountCreationTransactionID))

		// wait some seconds to make sure the mirror node receives the transaction data
		time.Sleep(5 * time.Second)

		// query the operator account information from a mirror node
		transactionInfo, err := hedera.Manager.MirrorNode.GetTransactionInfo(args.AccountCreationTransactionID)
		if err != nil {
			return fmt.Errorf("Error: %v", err)
		}

		// if the returned transaction is the account create transaction
		if transactionInfo.Name != "CRYPTOCREATEACCOUNT" {
			return fmt.Errorf("Error: Unexpected transaction type '%v'", transactionInfo.Name)
		}

		// Rename the original file to create a backup, if already on exists
		keystorepath := filepath.Join(RENDERHIVE_APP_DIRECTORY_CONFIG, strings.ReplaceAll(hedera.Manager.Operator.PublicKey.ToAccountID(0, 0).String(), ".", "")+".key")
		if isFile, _ := utility.IsFile(keystorepath); isFile {
			err = os.Rename(keystorepath, filepath.Join(RENDERHIVE_APP_DIRECTORY_CONFIG, strings.ReplaceAll(transactionInfo.EntityID, ".", "")+".key"))
			if err != nil {
				return err // Handle the error appropriately.
			}
		}

		// rename the keyfile
		err = hedera.Manager.LoadAccount(transactionInfo.EntityID, args.Passphrase, hedera.Manager.Operator.PublicKey.String())
		if err != nil {
			return fmt.Errorf("Error: %v", err)
		}

		// Write the node data
		err = node.Manager.WriteNodeData(-1, args.Node.Name, true, false, hedera.Manager.Operator.AccountID.String(), hedera.Manager.Operator.PublicKey.String())
		if err != nil {
			return fmt.Errorf("Error: %v", err)
		}

		// create reply for the RPC client
		reply.Message = "Successfully created the node account"
		reply.NodeAccountID = hedera.Manager.Operator.AccountID.String()
		reply.Payload = nil
		return nil
	}

	// STEP 3: Register with the smart contract
	if args.Step == "contract" {

		// TODO: Implement registration with the smart contract

		// create reply for the RPC client
		reply.Message = "Successfully created the node account"
		reply.NodeAccountID = hedera.Manager.Operator.AccountID.String()
		reply.Payload = nil
		return nil
	}

	// STEP 4: Sign up with the storage service
	if args.Step == "storage" {

		// TODO: Implement sign up with Filecoin service

		// create reply for the RPC client
		reply.Message = "Successfully signed up with the Filecoin service"
		reply.NodeAccountID = node.Manager.Node.HederaAccount.AccountID
		reply.Payload = nil
		return nil
	}

	// STEP 5: Finalize the signup
	if args.Step == "finalize" {

		// TODO: Implement finalization (if required)

	}
	// create reply for the RPC client
	reply.Message = "New operator was signed up: " + args.Operator.Username + "!"
	return nil

}

// Method: GetSignInPayload
// 			- request the payload for the sign in process
// #############################################################################

// Arguments and reply
type GetSignInPayloadArgs struct{}
type GetSignInPayloadReply struct {
	Payload []byte
}

// Signs in using a known operator account
func (ops *OperatorService) GetSignInPayload(r *http.Request, args *GetSignInPayloadArgs, reply *GetSignInPayloadReply) error {

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// Get the hash value of the node configuration file
	hash, err := node.Manager.HashNodeData()
	if err != nil {
		return err
	}

	// create reply for the RPC client
	reply.Payload = hash

	return nil

}

// Method: SignIn
// 			- sign in with the operator wallet
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

// Signs in using a known operator account
func (ops *OperatorService) SignIn(r *http.Request, args *SignInArgs, reply *SignInReply) error {
	var err error

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// LOAD PRIVATE KEY FROM KEYSTORE
	// *************************************************************************

	// log info
	logger.Manager.Main.Info().Msg("Received sign-in request from frontend.")

	// log info
	logger.Manager.Main.Info().Msg(" [#] Decrypting node private key and signing in ...")

	// LOAD ACCOUNT DETAILS AND DECRYPT THE NODE's PRIVATE KEY
	err = hedera.Manager.LoadAccount(node.Manager.Node.HederaAccount.AccountID, args.Passphrase, node.Manager.Node.HederaAccount.PublicKey)
	if err != nil {
		return fmt.Errorf("Failed to load private key: %v", err)
	}

	// GENERATE A SESSION JWT
	// *************************************************************************
	// TODO: We reuse the node's Hedera keys to sign and verify the session token.
	// 		 Would it be significantly more secure to create a new key pair here?

	// use the node's keys to sign a JWT
	Manager.SessionToken.PrivateKey = ed25519.NewKeyFromSeed(hedera.Manager.Operator.PrivateKey.BytesRaw())
	Manager.SessionToken.PublicKey = ed25519.NewKeyFromSeed(hedera.Manager.Operator.PrivateKey.BytesRaw()).Public().(ed25519.PublicKey)

	// generate the JWT and store it in the package manager
	Manager.SessionToken.SignedString, err = Manager.generateJWT(Manager.SessionToken.PrivateKey)
	if err != nil {
		return err
	}

	// update the status variable to notify the middleware that it should set a HttpOnly cookie
	Manager.SessionToken.Update = true

	// set a name for the HttpOnly cookie
	Manager.SessionCookie.Name = "renderhive-session"

	// READ HCS TOPIC INFORMATION & SUBSCRIBE
	// *************************************************************************
	// hive cycle synchronization topic
	if RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_SYNCHRONIZATION != "" {
		node.Manager.HiveCycleSynchronizationTopic, err = hedera.Manager.TopicInfoFromString(RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_SYNCHRONIZATION)
		if err != nil {
			return err
		}
		err = hedera.Manager.TopicSubscribe(node.Manager.HiveCycleSynchronizationTopic, time.Unix(0, 0), node.Manager.HiveCycle.MessageCallback())
		if err != nil {
			return err
		}
	}

	// hive cycle application topic
	if RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_APPLICATION != "" {
		node.Manager.HiveCycleApplicationTopic, err = hedera.Manager.TopicInfoFromString(RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_APPLICATION)
		if err != nil {
			return err
		}
		err = hedera.Manager.TopicSubscribe(node.Manager.HiveCycleApplicationTopic, time.Unix(0, 0), func(message hederasdk.TopicMessage) {

			logger.Manager.Package["hedera"].Info().Msg(fmt.Sprintf("Message received: %s", string(message.Contents)))

		})
		if err != nil {
			return err
		}
	}

	// hive cycle validation topic
	if RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_VALIDATION != "" {
		node.Manager.HiveCycleValidationTopic, err = hedera.Manager.TopicInfoFromString(RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_VALIDATION)
		if err != nil {
			return err
		}
		err = hedera.Manager.TopicSubscribe(node.Manager.HiveCycleValidationTopic, time.Unix(0, 0), func(message hederasdk.TopicMessage) {

			logger.Manager.Package["hedera"].Info().Msg(fmt.Sprintf("Message received: %s", string(message.Contents)))

		})
		if err != nil {
			return err
		}
	}

	// render job queue
	if RENDERHIVE_TESTNET_RENDER_JOB_QUEUE != "" {
		node.Manager.JobQueueTopic, err = hedera.Manager.TopicInfoFromString(RENDERHIVE_TESTNET_RENDER_JOB_QUEUE)
		if err != nil {
			return err
		}
		err = hedera.Manager.TopicSubscribe(node.Manager.JobQueueTopic, time.Unix(0, 0), node.Manager.JobQueueMessageCallback())
		if err != nil {
			return err
		}
	}

	// set the user session to active
	Manager.SessionActive = true

	// create reply for the RPC client
	reply.Message = "Operator signed in!"
	reply.SignedIn = Manager.SessionActive
	return nil

}

// Method: SignOut
// 			- sign out of the frontend session
// #############################################################################

// Arguments and reply
type SignOutArgs struct{}
type SignOutReply struct {
	Message  string
	SignedIn bool
}

// Signs in using a known operator account
func (ops *OperatorService) SignOut(r *http.Request, args *SignOutArgs, reply *SignOutReply) error {

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// log info
	logger.Manager.Main.Info().Msg("Received sign-out request from frontend.")

	// set the cookie expiry time to now
	Manager.SessionToken.ExpiresAt = time.Now()
	Manager.SessionToken.Update = true

	// set the user session to inactive
	Manager.SessionActive = false

	// create reply for the RPC client
	reply.Message = "Operator signed out!"
	reply.SignedIn = Manager.SessionActive
	return nil

}

// Method: GetInfo
//			- obtain information about the node operator from local files
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

// Get info about the operator via the accountid
func (ops *OperatorService) GetInfo(r *http.Request, args *GetInfoArgs, reply *GetInfoReply) error {

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// node operator details
	reply.Username = node.Manager.User.Username
	reply.UserEmail = node.Manager.User.Email
	reply.UserAccount = node.Manager.User.UserAccount.AccountID.String()

	// node details
	reply.NodeName = node.Manager.Node.Name
	reply.NodeAccount = node.Manager.Node.HederaAccount.AccountID

	return nil
}

// Method: GetContractInfo
//			- obtain information about the node operator from the smart contract
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

// Get info about the operator from the smart contract via the operator accountid
func (ops *OperatorService) GetContractInfo(r *http.Request, args *GetContractInfoArgs, reply *GetContractInfoReply) error {

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// TODO: Query the mirror node to check if the operator account id is known
	//		 to the smart contract

	// create reply with info abou the operator and the node for the client
	reply.Username = node.Manager.User.Username
	reply.UserEmail = node.Manager.User.Email
	reply.UserAccount = node.Manager.User.UserAccount.AccountID.String()
	reply.NodeAlias = node.Manager.Node.Name
	reply.NodeAccount = node.Manager.Node.HederaAccount.AccountID

	return nil
}

// Method: IsSessionValid
//			- checks if the session is valid
//			- this call can only respond, if the JWT is set and valid (otherwise
//			  the request will be blocked by the authenticationMiddleware)
// #############################################################################

// Arguments and reply
type IsSessionValidArgs struct{}

type IsSessionValidReply struct {
	Valid bool
}

// Returns true to the frontend if the session is valid
// NOTE: The function always returns true, but is only evaluated if the request from the frontend
//
//	contains a valid session token.
func (ops *OperatorService) IsSessionValid(r *http.Request, args *IsSessionValidArgs, reply *IsSessionValidReply) error {

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// create reply with info abou the operator and the node for the client
	reply.Valid = true

	return nil
}

// INTERNAL HELPER FUNCTIONS
// #############################################################################

// Read operator information known to this machine from a file
func (webappm *PackageManager) FromFile(path string) error {

	// read the operator file stored on this machine

	return nil
}

// helper function to generate the JSON Web Token for frontend session handling
func (webappm *PackageManager) generateJWT(privateKey ed25519.PrivateKey) (string, error) {

	// set expiry time to 1 hour
	webappm.SessionToken.ExpiresAt = time.Now().Add(time.Hour * 1)

	// define the token claims
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{
		"exp": webappm.SessionToken.ExpiresAt.Unix(),
	})

	// create the JWT
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("Error signing the token: %v", err)
	}

	return tokenString, nil
}

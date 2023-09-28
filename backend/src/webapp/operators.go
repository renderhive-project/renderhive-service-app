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
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	// external
	"github.com/golang-jwt/jwt/v5"
	hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

	// internal
	. "renderhive/globals"
	"renderhive/hedera"
	"renderhive/logger"
	"renderhive/node"
)

// SERVICE INITIALIZATION
// #############################################################################

// Defines the operator information
type Operator struct {
	ID        int    `json:"userid"`    // a unique user id
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
	//		 - Set up w3up access for the operator
	//		 - Create a node account on Hedera

	// CREATE A KEYSTORE FILE
	// // Rename the original file to create a backup
	// err = os.Rename(filepath.Join(RENDERHIVE_APP_DIRECTORY_CONFIG, strings.ReplaceAll(account_id, ".", "")+".key"), filepath.Join(RENDERHIVE_APP_DIRECTORY_CONFIG, strings.ReplaceAll(account_id, ".", "")+".bak"))
	// if err != nil {
	// 	return err // Handle the error appropriately.
	// }
	// // Open the keystore file
	// file, err := os.Create(filepath.Join(RENDERHIVE_APP_DIRECTORY_CONFIG, strings.ReplaceAll(account_id, ".", "")+".key"))
	// if err != nil {
	// 	return err
	// }
	// defer file.Close()

	// err = hm.Operator.PrivateKey.WriteKeystore(file, passphrase)
	// if err != nil {
	// 	fmt.Println("WriteKeystore Error:", err)
	// 	return err
	// }
	// fmt.Println(passphrase)

	// Write the user data into the node configuration
	publicKey := ""
	node.Manager.WriteUserData(args.Operator.ID, args.Operator.Username, args.Operator.Email, args.Operator.AccountID, publicKey)

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

	// Open the node configuration file
	file, err := os.Open(filepath.Join(RENDERHIVE_APP_DIRECTORY_CONFIG, "node.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a new SHA-256 hasher
	hasher := sha256.New()

	// Copy the file content to the hasher and calculate the fingerprint
	_, err = io.Copy(hasher, file)
	if err != nil {
		return err
	}

	// Get the hash value
	hash := hasher.Sum(nil)

	// create reply for the RPC client
	reply.Payload = hash

	return nil

}

// Method: SignIn
// 			- sign in with the operator wallet
// #############################################################################

// Arguments and reply
type SignInArgs struct {
	UserSignature []byte
	SignedPayload []byte
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

	// (1) VERIFY THE SIGNATURE
	verifiedSignature := node.Manager.User.UserAccount.PublicKey.Verify(args.SignedPayload, args.UserSignature)
	if !verifiedSignature {
		return errors.New("Invalid user signature")
	}

	// (2) DERIVE THE PASSPHRASE FROM THE SIGNED PAYLOAD
	hasher := sha256.New()
	hasher.Write(args.UserSignature)
	passphrase := hex.EncodeToString(hasher.Sum(nil))

	// log info
	logger.Manager.Main.Info().Msg("Decrypting node account details and signing in ...")

	// (3) LOAD ACCOUNT DETAILS AND DECRYPT THE NODE's PRIVATE KEY
	err = hedera.Manager.LoadAccount(node.Manager.Node.HederaAccount.AccountID, passphrase, node.Manager.Node.HederaAccount.PublicKey)
	if err != nil {
		return fmt.Errorf("Failed to load account details: %v", err)
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
//			- obtain information about the node operator
// #############################################################################

// Arguments and reply
type GetInfoArgs struct{}

type GetInfoReply struct {
	// Operator details
	Username    string
	UserEmail   string
	UserAccount string

	// Node Details
	NodeAlias   string
	NodeAccount string
}

// Get info about the operator via the accountid
func (ops *OperatorService) GetInfo(r *http.Request, args *GetInfoArgs, reply *GetInfoReply) error {

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// create reply with info abou the operator and the node for the client
	reply.Username = node.Manager.User.Username
	reply.UserEmail = node.Manager.User.Email
	reply.UserAccount = node.Manager.User.UserAccount.AccountID.String()
	reply.NodeAlias = node.Manager.Node.Alias
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

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

package node

import (

	// standard

	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	// external
	hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

	// "github.com/cockroachdb/apd"
	// "golang.org/x/exp/slices" <-- would be handy, but requires Go 1.18; TODO: Update possible for Hedera SDK?

	// internal
	. "renderhive/globals"
	"renderhive/hedera"
	"renderhive/logger"
	// . "renderhive/utility"
)

// USER DATA
// #############################################################################

// User data of the node's owner
// TODO: add further user data
type UserData struct {
	ID           int                    // Renderhive User ID given by the Renderhive Smart Contract
	Username     string                 // User name of the node operator
	Email        string                 // Email address of the node operator
	UserAccount  hedera.HederaAccount   // Hedera account ID of the user's main account
	NodeAccounts []hedera.HederaAccount // Hedera account IDs of the user's node accounts

}

// Define the JSON data structure for the user data
type UserDataJSON struct {
	ID            int    `json:"ID"`
	Username      string `json:"Username"`
	Email         string `json:"Email"`
	HederaAccount struct {
		AccountID string `json:"AccountID"`
		PublicKey string `json:"PublicKey"`
	} `json:"HederaAccount"`
}

// USER DATA HANDLING
// #############################################################################
// Set the details of the user operating this node
func (nm *PackageManager) WriteUserData(id int, username string, email string, accountid string, publicKey string) error {
	var err error
	var operator UserDataJSON

	// log event
	logger.Manager.Package["node"].Debug().Msg("Set new user data in the configuration file ...")

	// prepare the data for the JSON file
	operator.ID = id
	operator.Username = username
	operator.Email = email
	operator.HederaAccount.AccountID = accountid
	operator.HederaAccount.PublicKey = publicKey

	// store the operator data in a file, which can be loaded the next time
	data, err := json.MarshalIndent(operator, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal operator: %v", err)
	}
	err = os.WriteFile(filepath.Join(RENDERHIVE_APP_DIRECTORY_CONFIG, "operator.json"), data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	// read the user data into the node manager
	nm.User.ID = operator.ID
	nm.User.Username = operator.Username
	nm.User.Email = operator.Email

	// read the user account ID into the node manager
	nm.User.UserAccount.AccountID, err = hederasdk.AccountIDFromString(operator.HederaAccount.AccountID)
	if err != nil {
		return err
	}
	// read the user public key into the node manager
	nm.User.UserAccount.PublicKey, err = hederasdk.PublicKeyFromString(publicKey)
	if err != nil {
		return err
	}

	return err

}

// Read the details of the user operating this node from the configuration file
func (nm *PackageManager) ReadUserData() error {
	var err error
	var operator UserDataJSON

	// log event
	logger.Manager.Package["node"].Debug().Msg(" [#] Reading the operator data from the configuration file ...")

	// Open the configuration file
	file, err := os.Open(filepath.Join(RENDERHIVE_APP_DIRECTORY_CONFIG, "operator.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	// Read the file content
	fileData, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	// get the file content into a structure
	err = json.Unmarshal(fileData, &operator)
	if err != nil {
		return err
	}

	// read the user data into the node manager
	nm.User.ID = operator.ID
	nm.User.Username = operator.Username
	nm.User.Email = operator.Email

	// read the user account ID into the node manager
	nm.User.UserAccount.AccountID, err = hederasdk.AccountIDFromString(operator.HederaAccount.AccountID)
	if err != nil {
		return err
	}

	// read the public key into the node manager
	nm.User.UserAccount.PublicKey, err = hederasdk.PublicKeyFromString(operator.HederaAccount.PublicKey)
	if err != nil {
		return err
	}

	// log event
	logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] [*] User ID: %v", nm.User.ID))
	logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] [*] Username: %v", nm.User.Username))
	logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] [*] Email: %v", nm.User.Email))
	logger.Manager.Package["node"].Debug().Msg(fmt.Sprintf(" [#] [*] Hedera Account: %v", nm.User.UserAccount.AccountID.String()))

	return err

}

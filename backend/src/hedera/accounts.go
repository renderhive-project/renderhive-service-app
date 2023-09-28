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

package hedera

import (

	// standard
	"errors"
	"fmt"
	"os"

	// external
	hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

	// internal
	// . "renderhive/globals"
	"renderhive/logger"
)

// Hedera account / wallet data
type HederaAccount struct {

	// Hedera account ID
	AccountID hederasdk.AccountID

	// keys
	PrivateKey hederasdk.PrivateKey
	PublicKey  hederasdk.PublicKey

	// account information
	Info hederasdk.AccountInfo
}

// ACCOUNT MANAGEMENT
// #############################################################################
// Create a new account
func (h *HederaAccount) New(InitialBalance float64) (*hederasdk.TransactionReceipt, error) {
	var err error

	// log information
	logger.Manager.Package["hedera"].Debug().Msg("Create a new Hedera account on testnet:")

	// Generate a new private key for a the new account
	h.PrivateKey, err = hederasdk.PrivateKeyGenerateEd25519()
	if err != nil {
		return nil, err
	}
	logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf("[#] Private key: %v", h.PrivateKey))

	// get the public key
	h.PublicKey = h.PrivateKey.PublicKey()
	logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf("[#] Public key: %v", h.PublicKey))

	// Create new account, assign the public key, and transfer 1000 hBar to it
	newAccountTransaction, err := hederasdk.NewAccountCreateTransaction().
		SetKey(h.PublicKey).
		SetInitialBalance(hederasdk.HbarFrom(InitialBalance, hederasdk.HbarUnits.Tinybar)).
		Execute(Manager.NetworkClient)

	// Request the receipt of the account creation transaction
	transactionReceipt, err := newAccountTransaction.GetReceipt(Manager.NetworkClient)
	if err != nil {
		return nil, err
	}

	// Get the new account ID from the receipt
	h.AccountID = *transactionReceipt.AccountID
	logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf("[#] AccountID: %v", h.AccountID))

	return &transactionReceipt, nil
}

// Load the account information from a keystore file
func (h *HederaAccount) FromFile(filepath string, passphrase string, publickey string) error {
	var err error

	// Open the keystore file
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// load and decrypt the private key
	h.PrivateKey, err = hederasdk.PrivateKeyReadKeystore(file, passphrase)
	if err != nil {
		return err
	}

	// derive the public key
	h.PublicKey = h.PrivateKey.PublicKey()

	// check if the known public key is identical to the derived one
	if h.PublicKey.String() != publickey {
		return errors.New("Private key does not match the known node account.")
	}

	return nil
}

// Update the public key of a Hedera account
func (h *HederaAccount) UpdateKey(newKey *hederasdk.PrivateKey) (string, error) {
	var err error

	// Updating the account with the new key
	newAccountUpdateTransaction, err := hederasdk.NewAccountUpdateTransaction().
		SetAccountID(h.AccountID).
		// The new key
		SetKey(newKey.PublicKey()).
		FreezeWith(Manager.NetworkClient)
	if err != nil {
		return "", err
	}
	println(newKey.PublicKey().String())
	println(newKey.String())

	// Have to sign with both keys, the initial key first
	newAccountUpdateTransaction.Sign(Manager.Operator.PrivateKey)
	newAccountUpdateTransaction.Sign(*newKey)

	// Sign with client operator private key and submit the transaction to the Hedera network
	_, err = newAccountUpdateTransaction.Execute(Manager.NetworkClient)
	if err != nil {
		return "", err
	}

	return "", nil

}

// ACCOUNT QUERIES
// #############################################################################
// Query the Hedera network for information on this account
// NOTE: This should be used spareingly, since it has a network fee
func (h *HederaAccount) QueryInfo(m *PackageManager) (string, error) {
	var err error

	//Create the account info query
	newAccountInfoQuery := hederasdk.NewAccountInfoQuery().
		SetAccountID(h.AccountID)

	// get cost of this query
	cost, err := newAccountInfoQuery.GetCost(Manager.NetworkClient)
	if err != nil {
		return "", err
	}

	//Sign with client operator private key and submit the query to a Hedera network
	h.Info, err = newAccountInfoQuery.Execute(Manager.NetworkClient)
	if err != nil {
		return "", err
	}

	return cost.String(), nil
}

// Query the Hedera network for the account balance
func (h *HederaAccount) QueryBalance(m *PackageManager) (string, error) {
	var err error

	//Create the account info query
	newAccountBalanceQuery := hederasdk.NewAccountBalanceQuery().
		SetAccountID(h.AccountID)

	// get cost of this query
	cost, err := newAccountBalanceQuery.GetCost(Manager.NetworkClient)
	if err != nil {
		return "", err
	}

	//Sign with client operator private key and submit the query to a Hedera network
	accountBalance, err := newAccountBalanceQuery.Execute(Manager.NetworkClient)
	if err != nil {
		return "", err
	}

	// update the internal balance
	h.Info.Balance = accountBalance.Hbars

	return cost.String(), nil
}

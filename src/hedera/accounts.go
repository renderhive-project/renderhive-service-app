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
  "fmt"
  "os"

  // external
  "github.com/joho/godotenv"
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
  PublicKey hederasdk.PublicKey

  // account information
  Info hederasdk.AccountInfo

}



// ACCOUNT MANAGEMENT
// #############################################################################
// Create a new account
func (h *HederaAccount) New(m *PackageManager, InitialBalance float64) (*hederasdk.TransactionReceipt, error) {
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
        Execute(m.NetworkClient)

    // Request the receipt of the account creation transaction
    transactionReceipt, err := newAccountTransaction.GetReceipt(m.NetworkClient)
    if err != nil {
      return nil, err
    }

    // Get the new account ID from the receipt
    h.AccountID = *transactionReceipt.AccountID
    logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf("[#] AccountID: %v", h.AccountID))

    return &transactionReceipt, nil
}


// Load the account information from a file
func (h *HederaAccount) FromFile(filepath string) error {

    // TODO: DECRYPTION AND SAFE STORAGE IN MEMORY
    //       For now, we just use an .env file and load the testnet account data
    //       from there.

    // Loads the .env file and throws an error if it cannot load the variables
    // from that file correctly
    err := godotenv.Load(filepath)
    if err != nil {
        return err
    }

    // Grab your testnet account ID and private key from the .env file
    h.AccountID, err = hederasdk.AccountIDFromString(os.Getenv("TESTNET_ACCOUNT_ID"))
    if err != nil {
        return err
    }

    // get the private key
    h.PrivateKey, err = hederasdk.PrivateKeyFromString(os.Getenv("TESTNET_PRIVATE_KEY"))
    if err != nil {
        return err
    }

    // derive the public key
    h.PublicKey = h.PrivateKey.PublicKey()

    return nil
}

// Update the public key of a Hedera account
func (h *HederaAccount) UpdateKey(m *PackageManager, newKey *hederasdk.PrivateKey) (string, error) {
    var err error

    // Updating the account with the new key
  	newAccountUpdateTransaction, err := hederasdk.NewAccountUpdateTransaction().
  		SetAccountID(h.AccountID).
  		// The new key
  		SetKey(newKey.PublicKey()).
  		FreezeWith(m.NetworkClient)
  	if err != nil {
  		return "", err
  	}
    println(newKey.PublicKey().String())
    println(newKey.String())

  	// Have to sign with both keys, the initial key first
  	newAccountUpdateTransaction.Sign(m.Operator.PrivateKey)
  	newAccountUpdateTransaction.Sign(*newKey)

    // Sign with client operator private key and submit the transaction to the Hedera network
    _, err = newAccountUpdateTransaction.Execute(m.NetworkClient)
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
    cost, err := newAccountInfoQuery.GetCost(m.NetworkClient)
    if err != nil {
        return "", err
    }

    //Sign with client operator private key and submit the query to a Hedera network
    h.Info, err = newAccountInfoQuery.Execute(m.NetworkClient)
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
    cost, err := newAccountBalanceQuery.GetCost(m.NetworkClient)
    if err != nil {
        return "", err
    }

    //Sign with client operator private key and submit the query to a Hedera network
    accountBalance, err := newAccountBalanceQuery.Execute(m.NetworkClient)
    if err != nil {
        return "", err
    }

    // update the internal balance
    h.Info.Balance = accountBalance.Hbars

    return cost.String(), nil
}

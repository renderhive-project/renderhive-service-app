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

/*

The hedera package handles all functions related to the Hedera Hashgraph
services. This also includes a crypto wallet with very basic functionality.

*/

import (

	// standard
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	// external
	hederasdk "github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/spf13/cobra"

	// internal
	. "renderhive/globals"
	"renderhive/logger"
)

// define the network types
const (
	NETWORK_TYPE_TESTNET    = iota
	NETWORK_TYPE_PREVIEWNET = iota
	NETWORK_TYPE_MAINNET    = iota
)

// empty structure to hold all methods for the Hedera interactions
type PackageManager struct {

	// network communication
	NetworkType   int
	NetworkClient *hederasdk.Client

	// Hedera account of this node
	Operator HederaAccount

	// Mirror Node
	MirrorNode MirrorNode

	// Command line interface
	Command      *cobra.Command
	CommandFlags struct {
		FlagPlaceholder bool
	}
}

// HEDERA MANAGER
// #############################################################################
// create the hedera manager variable
var Manager = PackageManager{}

// Initialize everything required for communication with the Hedera network
func (hm *PackageManager) Init(NetworkType int) error {
	var err error

	logger.Manager.Package["hedera"].Debug().Msg("Initializing the Hedera manager ...")

	switch NetworkType {
	case NETWORK_TYPE_TESTNET:

		// log information
		logger.Manager.Package["hedera"].Info().Msg(" [#] Initializing on Hedera Testnet ...")

		// Create your testnet client
		hm.NetworkClient = hederasdk.ClientForTestnet()

	case NETWORK_TYPE_PREVIEWNET:

		// log information
		logger.Manager.Package["hedera"].Debug().Msg("Initializing on Hedera Previewnet:")

		// Create your preview client
		hm.NetworkClient = hederasdk.ClientForPreviewnet()

	case NETWORK_TYPE_MAINNET:

		// log information
		logger.Manager.Package["hedera"].Debug().Msg("Initializing on Hedera Mainnet:")

		// Create your preview client
		hm.NetworkClient = hederasdk.ClientForPreviewnet()

	}

	// set network type
	hm.NetworkType = NetworkType

	// get the mirror node URL
	hm.MirrorNode.URL = HEDERA_TESTNET_MIRROR_NODE_URL

	// log info
	logger.Manager.Main.Info().Msg(fmt.Sprintf(" [#] Mirror node: %v", hm.MirrorNode.URL))

	return err
}

// Load the account from the local data
func (hm *PackageManager) LoadAccount(account_id string, passphrase string, publickey string) error {
	var err error

	// read the node account ID into the node manager
	hm.Operator.AccountID, err = hederasdk.AccountIDFromString(account_id)
	if err != nil {
		return err
	}

	// log info
	logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Account ID: %v", hm.Operator.AccountID))

	// TODO: From this point on, the private key is in memory in clear text
	//		 This needs to be improved from a security standpoint!!!

	// read the private key from the keystore file and decrypt it
	err = hm.Operator.FromFile(filepath.Join(RENDERHIVE_APP_DIRECTORY_CONFIG, strings.ReplaceAll(account_id, ".", "")+".key"), passphrase, publickey)
	if err != nil {
		return err
	}

	// log info
	logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Public key: %v", hm.Operator.PublicKey))

	// set this account as the operator
	hm.NetworkClient.SetOperator(hm.Operator.AccountID, hm.Operator.PrivateKey)

	// // query the account balance from the Hedera network
	// queryCost, err := hm.Operator.QueryBalance(hm)
	// logger.Manager.Package["hedera"].Info().Msg(fmt.Sprintf(" [#] Account Balance: %v", hm.Operator.Info.Balance))
	// logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Costs (QueryBalance): %v", queryCost))

	// query the complete account information from the Hedera network
	queryCost, err := hm.Operator.QueryInfo(hm)
	logger.Manager.Package["hedera"].Info().Msg(fmt.Sprintf(" [#] Account Balance: %v", hm.Operator.Info.Balance))
	logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Costs (QueryInfo): %v", queryCost))

	return err
}

// Set the account from the given account ID and private key
func (hm *PackageManager) SetAccount(account_id string, privatekey string) error {
	var err error

	// read the node account ID into the node manager
	hm.Operator.AccountID, err = hederasdk.AccountIDFromString(account_id)
	if err != nil {
		return err
	}

	// log info
	logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Account ID: %v", hm.Operator.AccountID))

	// TODO: From this point on, the private key is in memory in clear text
	//		 This needs to be improved from a security standpoint!!!

	// read the private key from the keystore file and decrypt it
	hm.Operator.PrivateKey, err = hederasdk.PrivateKeyFromString(privatekey)
	hm.Operator.PublicKey = hm.Operator.PrivateKey.PublicKey()

	// log info
	logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Public key: %v", hm.Operator.PublicKey))

	// set this account as the operator
	hm.NetworkClient.SetOperator(hm.Operator.AccountID, hm.Operator.PrivateKey)

	// // query the account balance from the Hedera network
	// queryCost, err := hm.Operator.QueryBalance(hm)
	// logger.Manager.Package["hedera"].Info().Msg(fmt.Sprintf(" [#] Account Balance: %v", hm.Operator.Info.Balance))
	// logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Costs (QueryBalance): %v", queryCost))

	// query the complete account information from the Hedera network
	queryCost, err := hm.Operator.QueryInfo(hm)
	logger.Manager.Package["hedera"].Info().Msg(fmt.Sprintf(" [#] Account Balance: %v", hm.Operator.Info.Balance))
	logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Costs (QueryInfo): %v", queryCost))

	return err
}

// Deinitialize the Hedera manager
func (hm *PackageManager) DeInit() error {
	var err error

	// log event
	logger.Manager.Package["hedera"].Debug().Msg("Deinitializing the Hedera manager ...")

	return err

}

// GENERAL TRANSACTION OPTIONS & FUNCTIONS
// #############################################################################
// define types to enable optional transaction options for functions
type TransactionOption func(*TransactionSettings) error

// define the type for the transaction settings
type TransactionSettings struct {
	Execute          bool
	ExecuteAccountID hederasdk.AccountID

	// NOTE: has not been implemented for any transaction type yet
	Schedule              bool
	ScheduleExpiration    time.Time
	ScheduleWaitForExpiry bool
}

// create a settings object with the given options
func MakeTransactionSettings(options ...TransactionOption) (*TransactionSettings, error) {

	// default settings
	settings := &TransactionSettings{
		Execute:            true,
		Schedule:           false,
		ScheduleExpiration: time.Unix(0, 0),
	}

	// apply each option to the settings object
	for _, txOption := range options {
		err := txOption(settings)
		if err != nil {
			return nil, err
		}
	}

	return settings, nil
}

// create a type to enable transaction option functions
type TransactionOpts struct{}

// create a global variable to access the options
var TransactionOptions TransactionOpts

// SetExecute specifies whether the transaction should be directly executed or not
func (TransactionOpts) SetExecute(use bool, accountID hederasdk.AccountID) TransactionOption {
	return func(settings *TransactionSettings) error {
		settings.Execute = use
		settings.ExecuteAccountID = accountID
		return nil
	}
}

// SetSchedule specifies whether the transaction should be scheduled
func (TransactionOpts) SetSchedule(use bool, experiationTime time.Time, wait bool) TransactionOption {
	return func(settings *TransactionSettings) error {
		settings.Schedule = use
		settings.ScheduleExpiration = experiationTime
		settings.ScheduleWaitForExpiry = wait
		return nil
	}
}

// helper function to freeze a transaction for signature by an external wallet
func _TransactionFreeze(_transaction interface{}, options ...TransactionOption) (interface{}, error) {
	var err error
	var transaction interface{}

	// get the settings for the transaction
	settings, err := MakeTransactionSettings(options...)
	if err != nil {
		return nil, err
	}

	// if the transaction shall NOT be directly executed
	if settings.Execute == false {

		// get network accounts
		nodeAccountIDs := []hederasdk.AccountID{}
		for _, node := range Manager.NetworkClient.GetNetwork() {
			nodeAccountIDs = append(nodeAccountIDs, node)
		}

		// set the transaction ID
		transaction, err = hederasdk.TransactionSetTransactionID(_transaction, hederasdk.TransactionIDGenerate(settings.ExecuteAccountID))
		if err != nil {
			return nil, err
		}

		// set the transaction ID
		transaction, err = hederasdk.TransactionSetNodeAccountIDs(transaction, nodeAccountIDs)
		if err != nil {
			return nil, err
		}

		// freeze the transaction
		switch i := transaction.(type) {
		case hederasdk.AccountCreateTransaction:
			return i.Freeze()
		case hederasdk.AccountDeleteTransaction:
			return i.Freeze()
		case hederasdk.AccountUpdateTransaction:
			return i.Freeze()
		case hederasdk.AccountAllowanceApproveTransaction:
			return i.Freeze()
		case hederasdk.AccountAllowanceDeleteTransaction:
			return i.Freeze()
		case hederasdk.ContractCreateTransaction:
			return i.Freeze()
		case hederasdk.ContractDeleteTransaction:
			return i.Freeze()
		case hederasdk.ContractExecuteTransaction:
			return i.Freeze()
		case hederasdk.ContractUpdateTransaction:
			return i.Freeze()
		case hederasdk.FileAppendTransaction:
			return i.Freeze()
		case hederasdk.FileCreateTransaction:
			return i.Freeze()
		case hederasdk.FileDeleteTransaction:
			return i.Freeze()
		case hederasdk.FileUpdateTransaction:
			return i.Freeze()
		case hederasdk.LiveHashAddTransaction:
			return i.Freeze()
		case hederasdk.LiveHashDeleteTransaction:
			return i.Freeze()
		case hederasdk.ScheduleCreateTransaction:
			return i.Freeze()
		case hederasdk.ScheduleDeleteTransaction:
			return i.Freeze()
		case hederasdk.ScheduleSignTransaction:
			return i.Freeze()
		case hederasdk.SystemDeleteTransaction:
			return i.Freeze()
		case hederasdk.SystemUndeleteTransaction:
			return i.Freeze()
		case hederasdk.TokenAssociateTransaction:
			return i.Freeze()
		case hederasdk.TokenBurnTransaction:
			return i.Freeze()
		case hederasdk.TokenCreateTransaction:
			return i.Freeze()
		case hederasdk.TokenDeleteTransaction:
			return i.Freeze()
		case hederasdk.TokenDissociateTransaction:
			return i.Freeze()
		case hederasdk.TokenFeeScheduleUpdateTransaction:
			return i.Freeze()
		case hederasdk.TokenFreezeTransaction:
			return i.Freeze()
		case hederasdk.TokenGrantKycTransaction:
			return i.Freeze()
		case hederasdk.TokenMintTransaction:
			return i.Freeze()
		case hederasdk.TokenRevokeKycTransaction:
			return i.Freeze()
		case hederasdk.TokenUnfreezeTransaction:
			return i.Freeze()
		case hederasdk.TokenUpdateTransaction:
			return i.Freeze()
		case hederasdk.TokenWipeTransaction:
			return i.Freeze()
		case hederasdk.TopicCreateTransaction:
			return i.Freeze()
		case hederasdk.TopicDeleteTransaction:
			return i.Freeze()
		case hederasdk.TopicMessageSubmitTransaction:
			return i.Freeze()
		case hederasdk.TopicUpdateTransaction:
			return i.Freeze()
		case hederasdk.TransferTransaction:
			return i.Freeze()
		case *hederasdk.AccountCreateTransaction:
			return i.Freeze()
		case *hederasdk.AccountDeleteTransaction:
			return i.Freeze()
		case *hederasdk.AccountUpdateTransaction:
			return i.Freeze()
		case *hederasdk.AccountAllowanceApproveTransaction:
			return i.Freeze()
		case *hederasdk.AccountAllowanceDeleteTransaction:
			return i.Freeze()
		case *hederasdk.ContractCreateTransaction:
			return i.Freeze()
		case *hederasdk.ContractDeleteTransaction:
			return i.Freeze()
		case *hederasdk.ContractExecuteTransaction:
			return i.Freeze()
		case *hederasdk.ContractUpdateTransaction:
			return i.Freeze()
		case *hederasdk.FileAppendTransaction:
			return i.Freeze()
		case *hederasdk.FileCreateTransaction:
			return i.Freeze()
		case *hederasdk.FileDeleteTransaction:
			return i.Freeze()
		case *hederasdk.FileUpdateTransaction:
			return i.Freeze()
		case *hederasdk.LiveHashAddTransaction:
			return i.Freeze()
		case *hederasdk.LiveHashDeleteTransaction:
			return i.Freeze()
		case *hederasdk.ScheduleCreateTransaction:
			return i.Freeze()
		case *hederasdk.ScheduleDeleteTransaction:
			return i.Freeze()
		case *hederasdk.ScheduleSignTransaction:
			return i.Freeze()
		case *hederasdk.SystemDeleteTransaction:
			return i.Freeze()
		case *hederasdk.SystemUndeleteTransaction:
			return i.Freeze()
		case *hederasdk.TokenAssociateTransaction:
			return i.Freeze()
		case *hederasdk.TokenBurnTransaction:
			return i.Freeze()
		case *hederasdk.TokenCreateTransaction:
			return i.Freeze()
		case *hederasdk.TokenDeleteTransaction:
			return i.Freeze()
		case *hederasdk.TokenDissociateTransaction:
			return i.Freeze()
		case *hederasdk.TokenFeeScheduleUpdateTransaction:
			return i.Freeze()
		case *hederasdk.TokenFreezeTransaction:
			return i.Freeze()
		case *hederasdk.TokenGrantKycTransaction:
			return i.Freeze()
		case *hederasdk.TokenMintTransaction:
			return i.Freeze()
		case *hederasdk.TokenRevokeKycTransaction:
			return i.Freeze()
		case *hederasdk.TokenUnfreezeTransaction:
			return i.Freeze()
		case *hederasdk.TokenUpdateTransaction:
			return i.Freeze()
		case *hederasdk.TokenWipeTransaction:
			return i.Freeze()
		case *hederasdk.TopicCreateTransaction:
			return i.Freeze()
		case *hederasdk.TopicDeleteTransaction:
			return i.Freeze()
		case *hederasdk.TopicMessageSubmitTransaction:
			return i.Freeze()
		case *hederasdk.TopicUpdateTransaction:
			return i.Freeze()
		case *hederasdk.TransferTransaction:
			return i.Freeze()
		default:
			return transaction, errors.New("(BUG) non-exhaustive switch statement")
		}

	} else {

		// don't change anything
		transaction = _transaction

		// freeze the transaction
		switch i := transaction.(type) {
		case hederasdk.AccountCreateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.AccountDeleteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.AccountUpdateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.AccountAllowanceApproveTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.AccountAllowanceDeleteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.ContractCreateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.ContractDeleteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.ContractExecuteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.ContractUpdateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.FileAppendTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.FileCreateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.FileDeleteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.FileUpdateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.LiveHashAddTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.LiveHashDeleteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.ScheduleCreateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.ScheduleDeleteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.ScheduleSignTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.SystemDeleteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.SystemUndeleteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.TokenAssociateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.TokenBurnTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.TokenCreateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.TokenDeleteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.TokenDissociateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.TokenFeeScheduleUpdateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.TokenFreezeTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.TokenGrantKycTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.TokenMintTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.TokenRevokeKycTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.TokenUnfreezeTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.TokenUpdateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.TokenWipeTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.TopicCreateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.TopicDeleteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.TopicMessageSubmitTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.TopicUpdateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case hederasdk.TransferTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.AccountCreateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.AccountDeleteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.AccountUpdateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.AccountAllowanceApproveTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.AccountAllowanceDeleteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.ContractCreateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.ContractDeleteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.ContractExecuteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.ContractUpdateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.FileAppendTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.FileCreateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.FileDeleteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.FileUpdateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.LiveHashAddTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.LiveHashDeleteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.ScheduleCreateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.ScheduleDeleteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.ScheduleSignTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.SystemDeleteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.SystemUndeleteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.TokenAssociateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.TokenBurnTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.TokenCreateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.TokenDeleteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.TokenDissociateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.TokenFeeScheduleUpdateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.TokenFreezeTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.TokenGrantKycTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.TokenMintTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.TokenRevokeKycTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.TokenUnfreezeTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.TokenUpdateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.TokenWipeTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.TopicCreateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.TopicDeleteTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.TopicMessageSubmitTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.TopicUpdateTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		case *hederasdk.TransferTransaction:
			return i.FreezeWith(Manager.NetworkClient)
		default:
			return transaction, errors.New("(BUG) non-exhaustive switch statement")
		}
	}

}

// helper function to schedule a transaction
func _TransactionSchedule(_transaction interface{}, options ...TransactionOption) (interface{}, error) {
	var err error
	var transaction interface{}

	// get the settings for the transaction
	settings, err := MakeTransactionSettings(options...)
	if err != nil {
		return nil, err
	}

	// if the transaction is to be scheduled
	if settings.Schedule == true {

		// create a scheduled transaction
		newScheduleTransaction := hederasdk.NewScheduleCreateTransaction().
			SetPayerAccountID(Manager.Operator.AccountID).
			SetExpirationTime(settings.ScheduleExpiration).
			SetWaitForExpiry(settings.ScheduleWaitForExpiry)

		// add the scheduled transaction to the schedule transaction
		switch i := _transaction.(type) {
		case hederasdk.AccountCreateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.AccountDeleteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.AccountUpdateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.AccountAllowanceApproveTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.AccountAllowanceDeleteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.ContractCreateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.ContractDeleteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.ContractExecuteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.ContractUpdateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.FileAppendTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.FileCreateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.FileDeleteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.FileUpdateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.LiveHashAddTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.LiveHashDeleteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.ScheduleCreateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.ScheduleDeleteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.ScheduleSignTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.SystemDeleteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.SystemUndeleteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.TokenAssociateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.TokenBurnTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.TokenCreateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.TokenDeleteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.TokenDissociateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.TokenFeeScheduleUpdateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.TokenFreezeTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.TokenGrantKycTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.TokenMintTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.TokenRevokeKycTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.TokenUnfreezeTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.TokenUpdateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.TokenWipeTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.TopicCreateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.TopicDeleteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.TopicMessageSubmitTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.TopicUpdateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case hederasdk.TransferTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(&i)
		case *hederasdk.AccountCreateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.AccountDeleteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.AccountUpdateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.AccountAllowanceApproveTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.AccountAllowanceDeleteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.ContractCreateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.ContractDeleteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.ContractExecuteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.ContractUpdateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.FileAppendTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.FileCreateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.FileDeleteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.FileUpdateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.LiveHashAddTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.LiveHashDeleteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.ScheduleCreateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.ScheduleDeleteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.ScheduleSignTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.SystemDeleteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.SystemUndeleteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.TokenAssociateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.TokenBurnTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.TokenCreateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.TokenDeleteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.TokenDissociateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.TokenFeeScheduleUpdateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.TokenFreezeTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.TokenGrantKycTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.TokenMintTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.TokenRevokeKycTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.TokenUnfreezeTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.TokenUpdateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.TokenWipeTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.TopicCreateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.TopicDeleteTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.TopicMessageSubmitTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.TopicUpdateTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		case *hederasdk.TransferTransaction:
			newScheduleTransaction, err = newScheduleTransaction.SetScheduledTransaction(i)
		default:
			return nil, errors.New("(BUG) non-exhaustive switch statement")
		}

		// check if the schedule transaction was successfully created
		if err != nil {
			return nil, err
		}

		// freeze the transaction for signing
		transaction, err = _TransactionFreeze(newScheduleTransaction, options...)
		if err != nil {
			return nil, err
		}

	} else {

		// don't change anything
		transaction = _transaction

	}

	return transaction, nil

}

// helper function that takes any transaction bytes, signs it with the current operator, and sends it to the Hedera network
func _ExecuteWithClient(transactionBytes []byte) (*hederasdk.TransactionReceipt, error) {
	var err error
	var transactionResponse hederasdk.TransactionResponse

	// create a new transaction from the bytes
	transactionInterface, err := hederasdk.TransactionFromBytes(transactionBytes)
	if err != nil {
		return nil, err
	}

	// execute the transaction
	transactionResponse, err = hederasdk.TransactionExecute(transactionInterface, Manager.NetworkClient)
	if err != nil {
		return nil, err
	}

	// get the transaction receipt
	transactionReceipt, err := transactionResponse.GetReceipt(Manager.NetworkClient)
	if err != nil {
		return nil, err
	}

	return &transactionReceipt, nil

}

// TOPIC MANAGEMENT
// #############################################################################
// Obtain the topic information from a TopicID given in string format
func (hm *PackageManager) TopicInfoFromString(topicID string) (*HederaTopic, error) {
	var err error

	logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf("Query topic information for TopicID string '%v':", topicID))
	// get the topic ID from a string
	hTopicID, err := hederasdk.TopicIDFromString(topicID)
	if err != nil {
		return nil, err
	}

	// create a HederaTopic variable and query the information
	topic := HederaTopic{ID: hTopicID}
	_, err = topic.QueryInfo(hm)
	if err != nil {
		return nil, err
	}
	logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Topic ID: %v", topic.ID))
	logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Topic Memo: %v", topic.Info.TopicMemo))

	return &topic, nil
}

// Subscribe to the topic
func (hm *PackageManager) TopicSubscribe(topic *HederaTopic, startTime time.Time, onNext func(message hederasdk.TopicMessage)) error {
	var err error

	logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf("Subscribe to topic with ID %v.", topic.ID))

	// subscribe to the topic
	err = topic.Subscribe(startTime, onNext)
	if err != nil {
		return err
	}

	return err
}

// HEDERA MANAGER COMMAND LINE INTERFACE
// #############################################################################
// Create the command for the command line interface
func (hm *PackageManager) CreateCommand() *cobra.Command {

	// create the package command
	hm.Command = &cobra.Command{
		Use:   "hedera",
		Short: "Commands for the interaction with the Hedera services",
		Long:  "This command and its sub-commands enable the interaction with the Hedera services required by the Renderhive network",
		Run: func(cmd *cobra.Command, args []string) {

			return

		},
	}

	return hm.Command

}

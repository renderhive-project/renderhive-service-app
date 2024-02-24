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
	"errors"
	"fmt"
	"time"

	// external
	// "github.com/joho/godotenv"
	hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

	// internal
	"renderhive/logger"
)

// Hedera consensus service topic data
type HederaTopic struct {
	ID   hederasdk.TopicID
	Info hederasdk.TopicInfo
}

// TODO
// #############################################################################
//
// (1) It would be better to make the query calls to a mirror node most of the time
//     unless in situations where we cannot cope with the time delay of mirror nodes.
//     (this would lower costs for users)
//
//

// TOPIC MANAGEMENT - LOWER LEVEL FUNCTIONS
// #############################################################################
// This section contains the lower level functions for the topic management.
// These functions are extended by higher level functions which are introduced
// for convenience purposes in the root file.

// Create a new topic with a given name as the topic memo, without signing and executing the transaction
func (topic *HederaTopic) New(adminKey interface{}, options ...TransactionOption) (*hederasdk.TransactionReceipt, []byte, error) {
	var err error

	// get the settings for the transaction
	settings, err := MakeTransactionSettings(options...)
	if err != nil {
		return nil, nil, err
	}

	// create a new topic
	newTopicCreateTransaction := hederasdk.NewTopicCreateTransaction()

	// if a topic memo was passed
	if topic.Info.TopicMemo != "" {
		newTopicCreateTransaction = newTopicCreateTransaction.SetTopicMemo(topic.Info.TopicMemo)
	}

	// if a admin key was passed
	if topic.Info.AdminKey != nil {

		// set it in the transaction
		newTopicCreateTransaction = newTopicCreateTransaction.SetAdminKey(topic.Info.AdminKey)

		// Freeze the transaction for signing (this prevents the transaction can be
		// while signing it)
		frozenTopicCreateTransaction := newTopicCreateTransaction //.FreezeWith(Manager.NetworkClient)
		if err != nil {
			return nil, nil, err
		}

		// if the type of the passed key is a PrivateKey
		thisKey, ok := adminKey.(hederasdk.PrivateKey)
		if ok == true {

			logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing message with private key of: %s", thisKey.PublicKey()))

			// and sign the transaction with this key
			newTopicCreateTransaction = frozenTopicCreateTransaction.Sign(thisKey)

		}

		// if the type of the passed key is a slice of PrivateKey
		keyList, ok := adminKey.([]hederasdk.PrivateKey)
		if ok == true {

			// iterate through all keys
			for i, thisKey := range keyList {

				logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing message with private key (%v) of: %s", i, thisKey.PublicKey()))

				// and sign the transaction with each key
				newTopicCreateTransaction = frozenTopicCreateTransaction.Sign(thisKey)

			}

		}
	}

	// if a submit key was passed, set it in the transaction
	if topic.Info.SubmitKey != nil {
		newTopicCreateTransaction = newTopicCreateTransaction.SetSubmitKey(topic.Info.SubmitKey)
	}

	// if the transaction should be directly executed
	if settings.Execute {

		// sign with client operator private key and submit the query to a Hedera network
		transactionResponse, err := newTopicCreateTransaction.Execute(Manager.NetworkClient)
		if err != nil {
			return nil, nil, err
		}

		// get the transaction receipt
		transactionReceipt, err := transactionResponse.GetReceipt(Manager.NetworkClient)
		if err != nil {
			return nil, nil, err
		}

		// log the receipt status of the transaction
		logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", transactionReceipt.TransactionID.String(), transactionReceipt.Status))

		// get the topic ID from the transaction receipt
		topic.ID = *transactionReceipt.TopicID

		// query the topic information from the network
		_, err = topic.QueryInfo(&Manager)
		if err != nil {
			return nil, nil, err
		}

		return &transactionReceipt, nil, err

	}

	// freeze the transaction for signing
	frozenTransaction, err := newTopicCreateTransaction.FreezeWith(Manager.NetworkClient)
	if err != nil {
		return nil, nil, err
	}

	// get the transaction bytes
	transactionBytes, err := frozenTransaction.ToBytes()
	if err != nil {
		return nil, nil, err
	}

	return nil, transactionBytes, err

}

// Update the topic
func (topic *HederaTopic) Update(updatedInfo *hederasdk.TopicInfo, oldAdminKey interface{}, options ...TransactionOption) (*hederasdk.TransactionReceipt, []byte, error) {
	var err error

	// get the settings for the transaction
	settings, err := MakeTransactionSettings(options...)
	if err != nil {
		return nil, nil, err
	}

	// delete the topic
	newTopicUpdateTransaction := hederasdk.NewTopicUpdateTransaction().
		SetTopicID(topic.ID)

	// if the AdminKey shall be updated
	if updatedInfo.AdminKey != nil {
		newTopicUpdateTransaction = newTopicUpdateTransaction.SetAdminKey(updatedInfo.AdminKey)
	}

	// if the SubmitKey shall be updated
	if updatedInfo.AdminKey != nil {
		newTopicUpdateTransaction = newTopicUpdateTransaction.SetSubmitKey(updatedInfo.SubmitKey)
	}

	// if the topic memo shall be updated
	if updatedInfo.TopicMemo != "" {
		newTopicUpdateTransaction = newTopicUpdateTransaction.SetTopicMemo(updatedInfo.TopicMemo)
	}

	// if the topic has a AdminKey
	if topic.Info.AdminKey != nil {

		// Freeze the transaction for signing (this prevents the transaction can be
		// while signing it)
		frozenTopicUpdateTransaction, err := newTopicUpdateTransaction.FreezeWith(Manager.NetworkClient)
		if err != nil {
			return nil, nil, err
		}

		// if the type of the passed key is a PrivateKey
		thisKey, ok := oldAdminKey.(hederasdk.PrivateKey)
		if ok == true {

			logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing message with private key of: %s", thisKey.PublicKey()))

			// and sign the transaction with this key
			newTopicUpdateTransaction = frozenTopicUpdateTransaction.Sign(thisKey)

		}

		// if the type of the passed key is a slice of PrivateKey
		keyList, ok := oldAdminKey.([]hederasdk.PrivateKey)
		if ok == true {

			// iterate through all keys
			for i, thisKey := range keyList {

				logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing message with private key (%v) of: %s", i, thisKey.PublicKey()))

				// and sign the transaction with each key
				newTopicUpdateTransaction = frozenTopicUpdateTransaction.Sign(thisKey)

			}

		}

	}

	// if the transaction should be directly executed
	if settings.Execute {

		// sign with client operator private key and submit the query to a Hedera network
		// NOTE: This will only delete the topic, if the operator account's key was set
		//       as admin key
		transactionResponse, err := newTopicUpdateTransaction.Execute(Manager.NetworkClient)
		if err != nil {
			return nil, nil, err
		}

		// get the transaction receipt
		transactionReceipt, err := transactionResponse.GetReceipt(Manager.NetworkClient)
		if err != nil {
			return nil, nil, err
		}

		// log the receipt status of the transaction
		logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", transactionReceipt.TransactionID.String(), transactionReceipt.Status))

		// free the pointer
		topic = nil //&HederaTopic{}

		return &transactionReceipt, nil, err

	}

	// freeze the transaction for signing
	frozenTransaction, err := newTopicUpdateTransaction.FreezeWith(Manager.NetworkClient)
	if err != nil {
		return nil, nil, err
	}

	// get the transaction bytes
	transactionBytes, err := frozenTransaction.ToBytes()
	if err != nil {
		return nil, nil, err
	}

	return nil, transactionBytes, err

}

// Delete the topic
func (topic *HederaTopic) Delete(adminKey interface{}, options ...TransactionOption) (*hederasdk.TransactionReceipt, []byte, error) {
	var err error
	var transaction interface{}

	// get the settings for the transaction
	settings, err := MakeTransactionSettings(options...)
	if err != nil {
		return nil, nil, err
	}

	// delete the topic
	transaction = hederasdk.NewTopicDeleteTransaction().
		SetTopicID(topic.ID)

	// freeze the transaction for signing
	transaction, err = _TransactionFreeze(transaction, options...)
	if err != nil {
		return nil, nil, err
	}

	// if the topic has a AdminKey
	if topic.Info.AdminKey != nil {

		// if the type of the passed key is a PrivateKey
		thisKey, ok := adminKey.(hederasdk.PrivateKey)
		if ok == true {

			logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing message with private key of: %s", thisKey.PublicKey()))

			// and sign the transaction with this key
			transaction, err = hederasdk.TransactionSign(transaction, thisKey)
			if err != nil {
				return nil, nil, err
			}

		}

		// if the type of the passed key is a slice of PrivateKey
		keyList, ok := adminKey.([]hederasdk.PrivateKey)
		if ok == true {

			// iterate through all keys
			for i, thisKey := range keyList {

				logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing message with private key (%v) of: %s", i, thisKey.PublicKey()))

				// and sign the transaction with each key
				transaction, err = hederasdk.TransactionSign(transaction, thisKey)
				if err != nil {
					return nil, nil, err
				}

			}

		}

	}

	// if the transaction should be directly executed
	if settings.Execute {

		// sign with client operator private key and submit the query to a Hedera network
		// NOTE: This will only delete the topic, if the operator account's key was set
		//       as admin key
		transactionResponse, err := hederasdk.TransactionExecute(transaction, Manager.NetworkClient)
		if err != nil {
			return nil, nil, err
		}

		// get the transaction receipt
		transactionReceipt, err := transactionResponse.GetReceipt(Manager.NetworkClient)
		if err != nil {
			return nil, nil, err
		}

		// log the receipt status of the transaction
		logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", transactionReceipt.TransactionID.String(), transactionReceipt.Status))

		// free the pointer
		topic = nil //&HederaTopic{}

		return &transactionReceipt, nil, err

	}

	// get the transaction bytes
	transactionBytes, err := hederasdk.TransactionToBytes(transaction)
	if err != nil {
		return nil, nil, err
	}

	return nil, transactionBytes, err

}

// Submit a message to the topic
func (topic *HederaTopic) SubmitMessage(message string, memo string, submitKey interface{}, options ...TransactionOption) (*hederasdk.TransactionReceipt, []byte, error) {
	var err error
	var transaction interface{}
	var transactionResponse hederasdk.TransactionResponse

	// get the settings for the transaction
	settings, err := MakeTransactionSettings(options...)
	if err != nil {
		return nil, nil, err
	}

	// prepare the submit message transaction
	transaction = hederasdk.NewTopicMessageSubmitTransaction().
		SetTopicID(topic.ID).
		SetTransactionMemo(memo).
		SetMessage([]byte(message))

	// freeze the transaction for signing
	transaction, err = _TransactionFreeze(transaction, options...)
	if err != nil {
		return nil, nil, err
	}

	// if the topic has a SubmitKey
	if topic.Info.SubmitKey != nil {

		// if the type of the passed key is a PrivateKey
		thisKey, ok := submitKey.(hederasdk.PrivateKey)
		if ok == true {

			logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing message with private key of: %s", thisKey.PublicKey()))

			// and sign the transaction with this key
			transaction, err = hederasdk.TransactionSign(transaction, thisKey)
			if err != nil {
				return nil, nil, err
			}

		}

		// if the type of the passed key is a slice of PrivateKey
		keyList, ok := submitKey.([]hederasdk.PrivateKey)
		if ok == true {

			// iterate through all keys
			for i, thisKey := range keyList {

				logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing message with private key (%v) of: %s", i, thisKey.PublicKey()))

				// and sign the transaction with each key
				transaction, err = hederasdk.TransactionSign(transaction, thisKey)
				if err != nil {
					return nil, nil, err
				}

			}

		}

	}

	// schedule the transaction if required
	transaction, err = _TransactionSchedule(transaction, options...)
	if err != nil {
		return nil, nil, err
	}

	// if the transaction should be directly executed
	if settings.Execute {

		// sign with client operator private key and submit the query to a Hedera network
		// NOTE: If a submit key was set, this will only work, if the operator account's
		//       key was set as submit key
		transactionResponse, err = hederasdk.TransactionExecute(transaction, Manager.NetworkClient)
		if err != nil {
			return nil, nil, err
		}

		// get the transaction receipt
		transactionReceipt, err := transactionResponse.GetReceipt(Manager.NetworkClient)
		if err != nil {
			return nil, nil, err
		}

		// log the receipt status of the transaction
		logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", transactionReceipt.TransactionID.String(), transactionReceipt.Status))
		if settings.Schedule == true {
			logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Schedule ID: %v", transactionReceipt.ScheduleID))
		}

		return &transactionReceipt, nil, err

	}

	// get the transaction bytes
	transactionBytes, err := hederasdk.TransactionToBytes(transaction)
	if err != nil {
		return nil, nil, err
	}

	return nil, transactionBytes, err

}

// Sign a scheduled submit message on the network
func (topic *HederaTopic) SignSubmitMessage(scheduleID hederasdk.ScheduleID, privateKey interface{}) (*hederasdk.TransactionReceipt, error) {
	var err error
	var transactionResponse hederasdk.TransactionResponse

	// Submit the first signature
	newScheduleSignTransaction, err := hederasdk.NewScheduleSignTransaction().
		SetScheduleID(scheduleID).
		FreezeWith(Manager.NetworkClient)
	if err != nil {
		return nil, err
	}

	// if the type of the passed key is a PrivateKey
	thisKey, ok := privateKey.(hederasdk.PrivateKey)
	if ok == true {
		// add the signature for the scheduled transaction
		newScheduleSignTransaction = newScheduleSignTransaction.Sign(thisKey)
	} else {
		return nil, errors.New("No valid private key was passed.")
	}

	// submit the transaction to the network
	transactionResponse, err = newScheduleSignTransaction.Execute(Manager.NetworkClient)
	if err != nil {
		return nil, err
	}

	// get the transaction receipt
	transactionReceipt, err := transactionResponse.GetReceipt(Manager.NetworkClient)
	if err != nil {
		return nil, err
	}

	// log the receipt status of the transaction
	logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", transactionReceipt.TransactionID.String(), transactionReceipt.Status))

	return &transactionReceipt, err

}

// Query the Hedera network for information on the topic
// NOTE: This should be used spareingly, since it has a network fee
func (topic *HederaTopic) QueryInfo(m *PackageManager) (string, error) {
	var err error

	// create the topic info query
	newTopicInfoQuery := hederasdk.NewTopicInfoQuery().
		SetTopicID(topic.ID).
		SetMaxQueryPayment(hederasdk.NewHbar(1))

	// get cost of this query
	cost, err := newTopicInfoQuery.GetCost(Manager.NetworkClient)
	if err != nil {
		return "", err
	}

	// Sign with client operator private key and submit the query to a Hedera network
	topic.Info, err = newTopicInfoQuery.Execute(Manager.NetworkClient)
	if err != nil {
		return "", err
	}

	return cost.String(), nil
}

// Subscribe to the topic to receive all messages
// NOTE: The messages are requested from Hedera mirror nodes
func (topic *HederaTopic) Subscribe(startTime time.Time, onNext func(message hederasdk.TopicMessage)) error {
	var err error

	// create the topic info query
	newTopicMessageQuery := hederasdk.NewTopicMessageQuery().
		SetTopicID(topic.ID).
		SetStartTime(startTime)

	// subscribe to the topic
	_, err = newTopicMessageQuery.Subscribe(Manager.NetworkClient, onNext)
	if err != nil {
		return err
	}

	return err
}

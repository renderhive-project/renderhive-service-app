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
  "time"
  "errors"

  // external
  // "github.com/joho/godotenv"
  hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

  // internal
  "renderhive/logger"

)

// Hedera consensus service topic data
type HederaTopic struct {

  ID hederasdk.TopicID
  Info hederasdk.TopicInfo

}



// TOPIC MANAGEMENT - LOWER LEVEL FUNCTIONS
// #############################################################################
// This section contains the lower level functions for the topic management.
// These functions are extended by higher level functions which are introduced
// for convenience purposes in the root file.

// Create a new topic with a given name as the topic memo
func (topic *HederaTopic) New(m *HederaManager, adminKey interface{}) (*hederasdk.TransactionReceipt, error) {
  var err error

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
    frozenTopicCreateTransaction := newTopicCreateTransaction//.FreezeWith(m.NetworkClient)
    if err != nil {
        return nil, err
    }

    // if the type of the passed key is a PrivateKey
    thisKey, ok := adminKey.(hederasdk.PrivateKey)
    if ok == true {

        logger.RenderhiveLogger.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing message with private key of: %s", thisKey.PublicKey()))

        // and sign the transaction with this key
        newTopicCreateTransaction = frozenTopicCreateTransaction.Sign(thisKey)

    }

    // if the type of the passed key is a slice of PrivateKey
    keyList, ok := adminKey.([]hederasdk.PrivateKey)
    if ok == true {

        // iterate through all keys
        for i, thisKey := range keyList {

          logger.RenderhiveLogger.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing message with private key (%v) of: %s", i, thisKey.PublicKey()))

          // and sign the transaction with each key
          newTopicCreateTransaction = frozenTopicCreateTransaction.Sign(thisKey)

        }

    }
  }

  // if a submit key was passed, set it in the transaction
  if topic.Info.SubmitKey != nil {
    newTopicCreateTransaction = newTopicCreateTransaction.SetSubmitKey(topic.Info.SubmitKey)
  }

  // sign with client operator private key and submit the query to a Hedera network
  transactionResponse, err := newTopicCreateTransaction.Execute(m.NetworkClient)
  if err != nil {
  	return nil, err
  }

  // get the transaction receipt
  transactionReceipt, err := transactionResponse.GetReceipt(m.NetworkClient)
  if err != nil {
  	return nil, err
  }

  // log the receipt status of the transaction
  logger.RenderhiveLogger.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", transactionReceipt.TransactionID.String(), transactionReceipt.Status))

  // get the topic ID from the transaction receipt
  topic.ID = *transactionReceipt.TopicID

  // query the topic information from the network
  _, err = topic.QueryInfo(m)
  if err != nil {
  	return nil, err
  }

  return &transactionReceipt, err

}

// Update the topic
func (topic *HederaTopic) Update(m *HederaManager, updatedInfo *hederasdk.TopicInfo, oldAdminKey interface{}) (*hederasdk.TransactionReceipt, error) {
  var err error

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
    frozenTopicUpdateTransaction, err := newTopicUpdateTransaction.FreezeWith(m.NetworkClient)
    if err != nil {
        return nil, err
    }

    // if the type of the passed key is a PrivateKey
    thisKey, ok := oldAdminKey.(hederasdk.PrivateKey)
    if ok == true {

        logger.RenderhiveLogger.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing message with private key of: %s", thisKey.PublicKey()))

        // and sign the transaction with this key
        newTopicUpdateTransaction = frozenTopicUpdateTransaction.Sign(thisKey)

    }

    // if the type of the passed key is a slice of PrivateKey
    keyList, ok := oldAdminKey.([]hederasdk.PrivateKey)
    if ok == true {

        // iterate through all keys
        for i, thisKey := range keyList {

          logger.RenderhiveLogger.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing message with private key (%v) of: %s", i, thisKey.PublicKey()))

          // and sign the transaction with each key
          newTopicUpdateTransaction = frozenTopicUpdateTransaction.Sign(thisKey)

        }

    }

  }

  // sign with client operator private key and submit the query to a Hedera network
  // NOTE: This will only delete the topic, if the operator account's key was set
  //       as admin key
  transactionResponse, err := newTopicUpdateTransaction.Execute(m.NetworkClient)
  if err != nil {
  	return nil, err
  }

  // get the transaction receipt
  transactionReceipt, err := transactionResponse.GetReceipt(m.NetworkClient)
  if err != nil {
  	return nil, err
  }

  // log the receipt status of the transaction
  logger.RenderhiveLogger.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", transactionReceipt.TransactionID.String(), transactionReceipt.Status))

  // free the pointer
  topic = nil//&HederaTopic{}

  return &transactionReceipt, err

}

// Delete the topic
func (topic *HederaTopic) Delete(m *HederaManager, adminKey interface{}) (*hederasdk.TransactionReceipt, error) {
  var err error

  // delete the topic
  newTopicDeleteTransaction := hederasdk.NewTopicDeleteTransaction().
    SetTopicID(topic.ID)

  // if the topic has a AdminKey
  if topic.Info.AdminKey != nil {

    // Freeze the transaction for signing (this prevents the transaction can be
    // while signing it)
    newTopicDeleteTransaction, err := newTopicDeleteTransaction.FreezeWith(m.NetworkClient)
    if err != nil {
        return nil, err
    }

    // if the type of the passed key is a PrivateKey
    thisKey, ok := adminKey.(hederasdk.PrivateKey)
    if ok == true {

        logger.RenderhiveLogger.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing message with private key of: %s", thisKey.PublicKey()))

        // and sign the transaction with this key
        newTopicDeleteTransaction = newTopicDeleteTransaction.Sign(thisKey)

    }

    // if the type of the passed key is a slice of PrivateKey
    keyList, ok := adminKey.([]hederasdk.PrivateKey)
    if ok == true {

        // iterate through all keys
        for i, thisKey := range keyList {

          logger.RenderhiveLogger.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing message with private key (%v) of: %s", i, thisKey.PublicKey()))

          // and sign the transaction with each key
          newTopicDeleteTransaction = newTopicDeleteTransaction.Sign(thisKey)

        }

    }

  }

  // sign with client operator private key and submit the query to a Hedera network
  // NOTE: This will only delete the topic, if the operator account's key was set
  //       as admin key
  transactionResponse, err := newTopicDeleteTransaction.Execute(m.NetworkClient)
  if err != nil {
  	return nil, err
  }

  // get the transaction receipt
  transactionReceipt, err := transactionResponse.GetReceipt(m.NetworkClient)
  if err != nil {
  	return nil, err
  }

  // log the receipt status of the transaction
  logger.RenderhiveLogger.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", transactionReceipt.TransactionID.String(), transactionReceipt.Status))

  // free the pointer
  topic = nil//&HederaTopic{}

  return &transactionReceipt, err

}

// Submit a message to the topic
func (topic *HederaTopic) SubmitMessage(m *HederaManager, message string, submitKey interface{}, scheduleTxn bool, expirationTime *time.Time, waitForExpiry bool) (*hederasdk.TransactionReceipt, error) {
  var err error
  var newTopicMessageSubmitTransaction *hederasdk.TopicMessageSubmitTransaction
  var newScheduledTransaction *hederasdk.ScheduleCreateTransaction
  var transactionResponse hederasdk.TransactionResponse

  // prepare the submit message transaction
  newTopicMessageSubmitTransaction = hederasdk.NewTopicMessageSubmitTransaction().
    SetTopicID(topic.ID).
    SetMessage([]byte(message))

  // if the transaction is to be scheduled
  if scheduleTxn == true {

      // create a scheduled transaction
      newScheduledTransaction, err = hederasdk.NewScheduleCreateTransaction().
    		SetPayerAccountID(m.Operator.AccountID).
        SetExpirationTime(*expirationTime).
        SetWaitForExpiry(waitForExpiry).
        SetScheduledTransaction(newTopicMessageSubmitTransaction)

  }

  // if the topic has a SubmitKey
  if topic.Info.SubmitKey != nil {

    // if the transaction is to be scheduled
    if scheduleTxn == true {

      // Freeze the transaction for signing (this prevents the transaction can be
      // while signing it)
      newScheduledTransaction, err = newScheduledTransaction.FreezeWith(m.NetworkClient)
      if err != nil {
          return nil, err
      }

      // if the type of the passed key is a PrivateKey
      thisKey, ok := submitKey.(hederasdk.PrivateKey)
      if ok == true {

          logger.RenderhiveLogger.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing message with private key of: %s", thisKey.PublicKey()))

          // and sign the transaction with this key
          newScheduledTransaction = newScheduledTransaction.Sign(thisKey)

      }

      // if the type of the passed key is a slice of PrivateKey
      keyList, ok := submitKey.([]hederasdk.PrivateKey)
      if ok == true {

          // iterate through all keys
          for i, thisKey := range keyList {

            logger.RenderhiveLogger.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing message with private key (%v) of: %s", i, thisKey.PublicKey()))

            // and sign the transaction with each key
            newScheduledTransaction = newScheduledTransaction.Sign(thisKey)

          }

      }

      // sign with client operator private key and submit the query to a Hedera network
      // NOTE: If a submit key was set, this will only work, if the operator account's
      //       key was set as submit key
      transactionResponse, err = newScheduledTransaction.Execute(m.NetworkClient)
      if err != nil {
        return nil, err
      }

    // if the transaction is NOT to be scheduled
    } else {

        // Freeze the transaction for signing (this prevents the transaction can be
        // while signing it)
        newTopicMessageSubmitTransaction, err = newTopicMessageSubmitTransaction.FreezeWith(m.NetworkClient)
        if err != nil {
            return nil, err
        }

        // if the type of the passed key is a PrivateKey
        thisKey, ok := submitKey.(hederasdk.PrivateKey)
        if ok == true {

            logger.RenderhiveLogger.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing message with private key of: %s", thisKey.PublicKey()))

            // and sign the transaction with this key
            newTopicMessageSubmitTransaction = newTopicMessageSubmitTransaction.Sign(thisKey)

        }

        // if the type of the passed key is a slice of PrivateKey
        keyList, ok := submitKey.([]hederasdk.PrivateKey)
        if ok == true {

            // iterate through all keys
            for i, thisKey := range keyList {

              logger.RenderhiveLogger.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Signing message with private key (%v) of: %s", i, thisKey.PublicKey()))

              // and sign the transaction with each key
              newTopicMessageSubmitTransaction = newTopicMessageSubmitTransaction.Sign(thisKey)

            }

        }

        // sign with client operator private key and submit the query to a Hedera network
        // NOTE: If a submit key was set, this will only work, if the operator account's
        //       key was set as submit key
        transactionResponse, err = newTopicMessageSubmitTransaction.Execute(m.NetworkClient)
        if err != nil {
        	return nil, err
        }
    }

  }


  // get the transaction receipt
  transactionReceipt, err := transactionResponse.GetReceipt(m.NetworkClient)
  if err != nil {
  	return nil, err
  }

  // log the receipt status of the transaction
  logger.RenderhiveLogger.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", transactionReceipt.TransactionID.String(), transactionReceipt.Status))
  if scheduleTxn == true {
      logger.RenderhiveLogger.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Schedule ID: %v", transactionReceipt.ScheduleID))
  }

  return &transactionReceipt, err

}


// Sign a scheduled submit message on the network
func (topic *HederaTopic) SignSubmitMessage(m *HederaManager, scheduleID hederasdk.ScheduleID, privateKey interface{}) (*hederasdk.TransactionReceipt, error) {
  var err error
  var transactionResponse hederasdk.TransactionResponse

  //Submit the first signature
  newScheduleSignTransaction, err := hederasdk.NewScheduleSignTransaction().
  		SetScheduleID(scheduleID).
  		FreezeWith(m.NetworkClient)
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
  transactionResponse, err = newScheduleSignTransaction.Execute(m.NetworkClient)
  if err != nil {
  	return nil, err
  }

  // get the transaction receipt
  transactionReceipt, err := transactionResponse.GetReceipt(m.NetworkClient)
  if err != nil {
  	return nil, err
  }

  // log the receipt status of the transaction
  logger.RenderhiveLogger.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", transactionReceipt.TransactionID.String(), transactionReceipt.Status))

  return &transactionReceipt, err

}

// Query the Hedera network for information on the topic
// NOTE: This should be used spareingly, since it has a network fee
func (topic *HederaTopic) QueryInfo(m *HederaManager) (string, error) {
    var err error

    // create the topic info query
    newTopicInfoQuery := hederasdk.NewTopicInfoQuery().
  		SetTopicID(topic.ID).
  		SetMaxQueryPayment(hederasdk.NewHbar(1))

    // get cost of this query
    cost, err := newTopicInfoQuery.GetCost(m.NetworkClient)
    if err != nil {
        return "", err
    }

    // Sign with client operator private key and submit the query to a Hedera network
    topic.Info, err = newTopicInfoQuery.Execute(m.NetworkClient)
    if err != nil {
        return "", err
    }

    return cost.String(), nil
}


// Subscribe to the topic to receive all messages
// NOTE: The messages are requested from Hedera mirror nodes
func (topic *HederaTopic) Subscribe(m *HederaManager, startTime time.Time) error {
    var err error

    // create the topic info query
    newTopicMessageQuery := hederasdk.NewTopicMessageQuery().
  		SetTopicID(topic.ID).
      SetStartTime(startTime)

    // subscribe to the topic
    _, err = newTopicMessageQuery.Subscribe(m.NetworkClient, func(message hederasdk.TopicMessage) {

      logger.RenderhiveLogger.Package["hedera"].Info().Msg(fmt.Sprintf("Message received: %s", string(message.Contents)))

    })
    if err != nil {
        return err
    }

    return err
}

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

package main

/*

  This file is just used to initialize the HCS topics and smart contract and should
  only be called once. It is mainly for development purposes and will be removed
  once a first version of the Renderhive software is ready to be tested.

*/

import (

  // standard
  "fmt"
  "sync"
  "os"
  "time"
  "encoding/json"

  // external
  hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

  // internal
  . "renderhive/globals"
  "renderhive/logger"
  "renderhive/hedera"
  "renderhive/node"

)

// is this a test run? (if so, all topics are deleted afterwards)
const testRun = false

// error value
var err error
var ServiceApp AppManager

// INITIALIZE APP
// #############################################################################
func init() {

  // INITIALIZE SERVICE APP
  // ***************************************************************************
  // TODO: use the signal library to catch interrupts, so that the app still
  //       shuts down decently?
  ServiceApp = AppManager{}
  ServiceApp.Quit = make(chan bool, 1)
  ServiceApp.WG = sync.WaitGroup{}

  // initialize service app
  err = ServiceApp.Init()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

}

// MAIN LOOP
// #############################################################################
func main () {

  // prepare end of program
  defer os.Exit(0)

  // deinitialize the service app at the end of the main function
  defer ServiceApp.DeInit()

  // LOGGER SYSTEM
  // ***************************************************************************
  // log the start of the Renderhive Service App
  logger.Manager.Main.Info().Msg("Renderhive initialization started.")

  // make sure the end of the program is logged
  defer logger.Manager.Main.Info().Msg("Initializer was stopped.")


  // LOAD OPERATOR ACCOUNT DETAILS
  // ***************************************************************************
  // var response *hederasdk.TransactionResponse
  var receipt *hederasdk.TransactionReceipt

  // initialize the Hedera Manager
  HederaManager := *ServiceApp.HederaManager



  // HEDERA SMART CONTRACT SERVICE CONTRACT
  // ***************************************************************************
  logger.Manager.Main.Info().Msg("Creating the required Hedera Smart Contract ...")

  // define the renderhive version for the topic names
  const renderhive_contract_main_version = "0"
  const renderhive_contract_sub_version = "1"
  const renderhive_contract_patch_version = "0"

  // RENDERHIVE SMART CONTRACT
  if RENDERHIVE_TESTNET_SMART_CONTRACT == "" {

      // Create the smart contract
      logger.Manager.Main.Info().Msg(" [#] Renderhive smart contract:")

      // prepare the topic information, which are used to create the topic
      contract := hedera.HederaSmartContract{Info: hederasdk.ContractInfo{ContractMemo: fmt.Sprintf("renderhive-v%s.%s.%s::smart-contract", renderhive_contract_main_version, renderhive_contract_sub_version, renderhive_contract_patch_version), AdminKey: HederaManager.Operator.PublicKey}}
      _, receipt, err = contract.New(&HederaManager, "./RenderhiveTestContract.json", HederaManager.Operator.PrivateKey, 100000)
      if err != nil {
        logger.Manager.Package["hedera"].Error().Err(err).Msg("")
        os.Exit(1)
      }
      if receipt != nil {
        logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] [*] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
      }

      // get contract information
      logger.Manager.Package["hedera"].Debug().Msg(" [#] [*] Query the contract information ...")
      contractID, err := hederasdk.ContractIDFromString(contract.ID.String())
      if err != nil {
        logger.Manager.Package["hedera"].Error().Err(err).Msg("")
        os.Exit(1)
      }
      contract = hedera.HederaSmartContract{ID: contractID}
      _, err = contract.QueryInfo(&HederaManager)
      if err != nil {
        logger.Manager.Package["hedera"].Error().Err(err).Msg("")
        os.Exit(1)
      }
      logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] [*] Contract ID: %v", contract.ID))
      logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] [*] Contract Memo: %v", contract.Info.ContractMemo))


      // if this is a test run, delete the topic again
      if testRun {
        // wait loop
        time.Sleep(10 * time.Second)


        // Delete the created contract
        logger.Manager.Package["hedera"].Debug().Msg("Delete the created contract again")
        _, receipt, err = contract.Delete(&HederaManager, nil)
        if err != nil {
          logger.Manager.Package["hedera"].Error().Err(err).Msg("")
          os.Exit(1)
        }
        if receipt != nil {
          logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
        }
      }

  }


  // HEDERA CONSENSUS SERVICE TOPICS
  // ***************************************************************************
  logger.Manager.Main.Info().Msg("Creating the required Hedera Consensus Service topics ...")

  // define the Renderhive version for the topic names
  const renderhive_topic_main_version = "0"
  const renderhive_topic_sub_version = "1"
  const renderhive_topic_patch_version = "0"

  // HIVE CYCLE SYNCHRONIZATION
  if RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_SYNCHRONIZATION == "" {

    // Create the hive cycle topics
    logger.Manager.Main.Info().Msg(" [#] Hive Cycle Synchronization topic:")


    // prepare the topic information, which are used to create the topic
    topic := hedera.HederaTopic{Info: hederasdk.TopicInfo{TopicMemo: fmt.Sprintf("renderhive-v%s.%s.%s::hive-cycle-synchronization", renderhive_topic_main_version, renderhive_topic_sub_version, renderhive_topic_patch_version), AdminKey: HederaManager.Operator.PublicKey, SubmitKey: HederaManager.Operator.PublicKey}}
    receipt, err = topic.New(&HederaManager, HederaManager.Operator.PrivateKey)
    if err != nil {
      logger.Manager.Package["hedera"].Error().Err(err).Msg("")
      os.Exit(1)
    }
    if receipt != nil {
      logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] [*] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
    }

    // get existing topic information
    logger.Manager.Package["hedera"].Debug().Msg(" [#] [*] Query the topic information ...")
    topicID, err := hederasdk.TopicIDFromString(topic.ID.String())
    if err != nil {
      logger.Manager.Package["hedera"].Error().Err(err).Msg("")
      os.Exit(1)
    }
    topic = hedera.HederaTopic{ID: topicID}
    _, err = topic.QueryInfo(&HederaManager)
    if err != nil {
      logger.Manager.Package["hedera"].Error().Err(err).Msg("")
      os.Exit(1)
    }
    logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] [*] Topic ID: %v", topic.ID))
    logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] [*] Topic Memo: %v", topic.Info.TopicMemo))


    // Submit the configuration message to the synchronization topic
    logger.Manager.Package["hedera"].Debug().Msg(" [#] [*] Send the configuration message to the synchronization topic:")
    message := node.HiveCycleConfigurationMessage{
      Iteration: 1,
      Duration:  300,
      Timestamp: time.Now(),
    }
    // Encode the message as JSON
    jsonMessage, err := json.Marshal(message)
    if err != nil {
        fmt.Println(err)
        return
    } else {

      receipt, err = topic.SubmitMessage(&HederaManager, string(jsonMessage), HederaManager.Operator.PrivateKey, false, nil, false)
      if err != nil {
        logger.Manager.Package["hedera"].Error().Err(err).Msg("")
        os.Exit(1)
      }
      if receipt != nil {
        logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] [*] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
      }

    }


    // if this is a test run, delete the topic again
    if testRun {
      // wait loop
      time.Sleep(10 * time.Second)


      // Delete the created topic
      logger.Manager.Package["hedera"].Debug().Msg("Delete the created topic again")
      receipt, err = topic.Delete(&HederaManager, nil)
      if err != nil {
        logger.Manager.Package["hedera"].Error().Err(err).Msg("")
        os.Exit(1)
      }
      if receipt != nil {
        logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
      }
    }

  }

  // HIVE CYCLE APPLICATION
  if RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_APPLICATION == "" {

    // Create the hive cycle topics
    logger.Manager.Main.Info().Msg(" [#] Hive Cycle Application topic:")


    // prepare the topic information, which are used to create the topic
    topic := hedera.HederaTopic{Info: hederasdk.TopicInfo{TopicMemo: fmt.Sprintf("renderhive-v%s.%s.%s::hive-cycle-application", renderhive_topic_main_version, renderhive_topic_sub_version, renderhive_topic_patch_version), AdminKey: HederaManager.Operator.PublicKey}}
    receipt, err = topic.New(&HederaManager, HederaManager.Operator.PrivateKey)
    if err != nil {
      logger.Manager.Package["hedera"].Error().Err(err).Msg("")
      os.Exit(1)
    }
    if receipt != nil {
      logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] [*] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
    }

    // get existing topic information
    logger.Manager.Package["hedera"].Debug().Msg(" [#] [*] Query the topic information ...")
    topicID, err := hederasdk.TopicIDFromString(topic.ID.String())
    if err != nil {
      logger.Manager.Package["hedera"].Error().Err(err).Msg("")
      os.Exit(1)
    }
    topic = hedera.HederaTopic{ID: topicID}
    _, err = topic.QueryInfo(&HederaManager)
    if err != nil {
      logger.Manager.Package["hedera"].Error().Err(err).Msg("")
      os.Exit(1)
    }
    logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] [*] Topic ID: %v", topic.ID))
    logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] [*] Topic Memo: %v", topic.Info.TopicMemo))



    // if this is a test run, delete the topic again
    if testRun {
      // wait loop
      time.Sleep(10 * time.Second)


      // Delete the created topic
      logger.Manager.Package["hedera"].Debug().Msg("Delete the created topic again")
      receipt, err = topic.Delete(&HederaManager, nil)
      if err != nil {
        logger.Manager.Package["hedera"].Error().Err(err).Msg("")
        os.Exit(1)
      }
      if receipt != nil {
        logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
      }
    }

  }

  // HIVE CYCLE VALIDATION
  if RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_VALIDATION == "" {

    // Create the hive cycle topics
    logger.Manager.Main.Info().Msg(" [#] Hive Cycle Validation topic:")


    // prepare the topic information, which are used to create the topic
    topic := hedera.HederaTopic{Info: hederasdk.TopicInfo{TopicMemo: fmt.Sprintf("renderhive-v%s.%s.%s::hive-cycle-validation", renderhive_topic_main_version, renderhive_topic_sub_version, renderhive_topic_patch_version), AdminKey: HederaManager.Operator.PublicKey}}
    receipt, err = topic.New(&HederaManager, HederaManager.Operator.PrivateKey)
    if err != nil {
      logger.Manager.Package["hedera"].Error().Err(err).Msg("")
      os.Exit(1)
    }
    if receipt != nil {
      logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] [*] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
    }

    // get existing topic information
    logger.Manager.Package["hedera"].Debug().Msg(" [#] [*] Query the topic information ...")
    topicID, err := hederasdk.TopicIDFromString(topic.ID.String())
    if err != nil {
      logger.Manager.Package["hedera"].Error().Err(err).Msg("")
      os.Exit(1)
    }
    topic = hedera.HederaTopic{ID: topicID}
    _, err = topic.QueryInfo(&HederaManager)
    if err != nil {
      logger.Manager.Package["hedera"].Error().Err(err).Msg("")
      os.Exit(1)
    }
    logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] [*] Topic ID: %v", topic.ID))
    logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] [*] Topic Memo: %v", topic.Info.TopicMemo))



    // if this is a test run, delete the topic again
    if testRun {
      // wait loop
      time.Sleep(10 * time.Second)


      // Delete the created topic
      logger.Manager.Package["hedera"].Debug().Msg("Delete the created topic again")
      receipt, err = topic.Delete(&HederaManager, nil)
      if err != nil {
        logger.Manager.Package["hedera"].Error().Err(err).Msg("")
        os.Exit(1)
      }
      if receipt != nil {
        logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
      }
    }

  }



  // RENDER JOB QUEUE
  if RENDERHIVE_TESTNET_RENDER_JOB_QUEUE == "" {

    // Create the hive cycle topics
    logger.Manager.Main.Info().Msg(" [#] Render Job Queue topic:")


    // prepare the topic information, which are used to create the topic
    topic := hedera.HederaTopic{Info: hederasdk.TopicInfo{TopicMemo: fmt.Sprintf("renderhive-v%s.%s.%s::render-job-queue", renderhive_topic_main_version, renderhive_topic_sub_version, renderhive_topic_patch_version), AdminKey: HederaManager.Operator.PublicKey}}
    receipt, err = topic.New(&HederaManager, HederaManager.Operator.PrivateKey)
    if err != nil {
      logger.Manager.Package["hedera"].Error().Err(err).Msg("")
      os.Exit(1)
    }
    if receipt != nil {
      logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] [*] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
    }

    // get existing topic information
    logger.Manager.Package["hedera"].Debug().Msg(" [#] [*] Query the topic information ...")
    topicID, err := hederasdk.TopicIDFromString(topic.ID.String())
    if err != nil {
      logger.Manager.Package["hedera"].Error().Err(err).Msg("")
      os.Exit(1)
    }
    topic = hedera.HederaTopic{ID: topicID}
    _, err = topic.QueryInfo(&HederaManager)
    if err != nil {
      logger.Manager.Package["hedera"].Error().Err(err).Msg("")
      os.Exit(1)
    }
    logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] [*] Topic ID: %v", topic.ID))
    logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] [*] Topic Memo: %v", topic.Info.TopicMemo))



    // if this is a test run, delete the topic again
    if testRun {
      // wait loop
      time.Sleep(10 * time.Second)


      // Delete the created topic
      logger.Manager.Package["hedera"].Debug().Msg("Delete the created topic again")
      receipt, err = topic.Delete(&HederaManager, nil)
      if err != nil {
        logger.Manager.Package["hedera"].Error().Err(err).Msg("")
        os.Exit(1)
      }
      if receipt != nil {
        logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
      }
    }

  }
}

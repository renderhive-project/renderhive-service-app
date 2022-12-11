/*
 * ************************** BEGIN LICENSE BLOCK ******************************
 *
 * Copyright © 2022 Christian Stolze
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

import (

  // standard
  "fmt"
  "os"
  "time"

  // external
  hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

  // internal
  "renderhive/logger"
  "renderhive/cli"
  "renderhive/hedera"
)

// COMPILER FLAGS
// #############################################################################


// MAIN LOOP
// #############################################################################
func main () {

  // prepare end of program
  defer os.Exit(0)

  // error value
  var err error

  // placeholder
  fmt.Println(time.Now().Add(30 * time.Second))

  // LOGGER SYSTEM
  // ***************************************************************************
  // initialize the logger system
  logger.Init()

  // add the package loggers
  logger.AddPackageLogger("node")
  logger.AddPackageLogger("hedera")
  logger.AddPackageLogger("ipfs")

  // log the start of the renderhive service
  logger.RenderhiveLogger.Main.Info().Msg("Renderhive service started.")

  // make sure the end of the program is logged
  defer logger.RenderhiveLogger.Main.Info().Msg("Renderhive service stopped.")



  // HEDERA
  // ***************************************************************************
  var receipt *hederasdk.TransactionReceipt

  // initialize the Hedera Manager
  HederaManager, err := hedera.InitHederaManager(hedera.NETWORK_TYPE_TESTNET, "hedera/testnet.env")
  if err != nil {
    logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
    os.Exit(1)
  }


	keys := make([]hederasdk.PrivateKey, 3)
	pubKeys := make([]hederasdk.PublicKey, 3)
	privateKeys := make([]hederasdk.PrivateKey, 3)

	fmt.Println("Keys: ")

	// Loop to generate keys for the KeyList
	for i := range keys {

		newKey, err := hederasdk.GeneratePrivateKey()
		if err != nil {
			println(err.Error(), ": error generating PrivateKey}")
			return
		}

		fmt.Printf("Key %v:\n", i)
		fmt.Printf("private = %v\n", newKey)
		fmt.Printf("public = %v\n", newKey.PublicKey())

		keys[i] = newKey
		pubKeys[i] = newKey.PublicKey()
		privateKeys[i] = newKey

  }

  // add all public keys to a public key list
	pubKeyList := hederasdk.NewKeyList().
		AddAllPublicKeys(pubKeys)

  fmt.Printf("Keylist: %i", len(pubKeys))
  fmt.Println(pubKeyList)


  // create a new topic
  logger.RenderhiveLogger.Package["hedera"].Debug().Msg("Creating the heartbeat topic:")
  // prepare the topic information, which are used to create the topic
  topic := hedera.HederaTopic{Info: hederasdk.TopicInfo{TopicMemo: "renderhive-v0.1.0::heartbeat", AdminKey: pubKeyList, SubmitKey: pubKeyList}}
  receipt, err = topic.New(&HederaManager, privateKeys)
  if err != nil {
    logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
    os.Exit(1)
  }
  if receipt != nil {
    logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
  }

  // get existing topic information
  logger.RenderhiveLogger.Package["hedera"].Debug().Msg("Query the heartbeat topic information:")
  topicID, err := hederasdk.TopicIDFromString(topic.ID.String())
  if err != nil {
    logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
    os.Exit(1)
  }
  topic = hedera.HederaTopic{ID: topicID}
  _, err = topic.QueryInfo(&HederaManager)
  if err != nil {
    logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
    os.Exit(1)
  }
  logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Topic ID: %v", topic.ID))
  logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Topic Memo: %v", topic.Info.TopicMemo))


  // subscribe to the topic
  logger.RenderhiveLogger.Package["hedera"].Debug().Msg("Subscribe to the heartbeat topic:")
  err = topic.Subscribe(&HederaManager, time.Unix(0, 0))
  if err != nil {
    logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
    os.Exit(1)
  }

  // Update the topic memo and adminkey
  logger.RenderhiveLogger.Package["hedera"].Debug().Msg("Updating the heartbeat topic information:")
  updateInfo := hederasdk.TopicInfo{TopicMemo: "renderhive-v0.1.1::heartbeat", AdminKey: HederaManager.Operator.PublicKey}
  receipt, err = topic.Update(&HederaManager, &updateInfo, privateKeys)
  if err != nil {
    logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
    os.Exit(1)
  }
  if receipt != nil {
    logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
  }

  // get existing topic information
  logger.RenderhiveLogger.Package["hedera"].Debug().Msg("Query the heartbeat topic information:")
  topicID, err = hederasdk.TopicIDFromString(topic.ID.String())
  if err != nil {
    logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
    os.Exit(1)
  }
  topic = hedera.HederaTopic{ID: topicID}
  _, err = topic.QueryInfo(&HederaManager)
  if err != nil {
    logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
    os.Exit(1)
  }
  logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Topic ID: %v", topic.ID))
  logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] New topic memo: %v", topic.Info.TopicMemo))

  // Submit a normal message to the topic
  message := "This is a test message!"
  logger.RenderhiveLogger.Package["hedera"].Debug().Msg("Send a message to the heartbeat topic:")
  receipt, err = topic.SubmitMessage(&HederaManager, message, privateKeys, false, nil, false)
  if err != nil {
    logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
    os.Exit(1)
  }
  if receipt != nil {
    logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
  }

  // wait for 2.5 seconds
  time.Sleep(5 * time.Second)

  // Submit a scheduled message to the topic
  message = "This is a scheduled test message!"
  expirationTime := time.Now().Add(30 * time.Second)
  logger.RenderhiveLogger.Package["hedera"].Debug().Msg("Send a message to the heartbeat topic:")
  receipt, err = topic.SubmitMessage(&HederaManager, message, privateKeys[0], true, &expirationTime, false)
  if err != nil {
    logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
    os.Exit(1)
  }
  scheduleID := *receipt.ScheduleID
  if receipt != nil {
    logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
    logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Schedule ID: %v", scheduleID))
  }

  // wait for 2.5 seconds
  time.Sleep(5 * time.Second)

  // Sign the scheduled transaction with the missing keys
  receipt, err = topic.SignSubmitMessage(&HederaManager, scheduleID, privateKeys[1])
  if err != nil {
    logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
    os.Exit(1)
  }
  if receipt != nil {
    logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
  }
  receipt, err = topic.SignSubmitMessage(&HederaManager, scheduleID, privateKeys[2])
  if err != nil {
    logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
    os.Exit(1)
  }
  if receipt != nil {
    logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
  }

  // wait loop
  time.Sleep(10 * time.Second)


  // Delete the created topic
  logger.RenderhiveLogger.Package["hedera"].Debug().Msg("Delete the created topic again")
  receipt, err = topic.Delete(&HederaManager, nil)
  if err != nil {
    logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
    os.Exit(1)
  }
  if receipt != nil {
    logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
  }
  // COMMAND LINE INTERFACE
  // ***************************************************************************
  // start the command line interface tool
  cli_tool := cli.CommandLine{}
  cli_tool.Start()

  // fmt.Println(transactionReceipt)

  //
	// // Create a schedule transaction
	// transaction, err := transactionToSchedule.Schedule()
  //
	// if err != nil {
	// panic(err)
	// }
  //
  // //Sign with the client operator key and submit the transaction to a Hedera network
  // txResponse, err := transaction.Execute(client)
  // ​
  // if err != nil {
  // 	panic(err)
  // }
  // ​
  // //Request the receipt of the transaction
  // receipt, err := txResponse.GetReceipt(client)
  // if err != nil {
  // 	panic(err)
  // }
  // ​
  // //Get the schedule ID from the receipt
  // scheduleId := *receipt.ScheduleID
  // ​
  // fmt.Printf("The new token ID is %v\n", scheduleId)

}

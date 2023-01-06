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
  . "renderhive/constants"
  "renderhive/logger"
  "renderhive/cli"
  "renderhive/hedera"
  // "renderhive/node"
)


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

  // log some informations about the used constants
  logger.RenderhiveLogger.Main.Info().Msg("This service app instance relies on the following smart contract(s) and HCS topic(s):")
  // the renderhive smart contract this instance calls
  logger.RenderhiveLogger.Main.Info().Msg(fmt.Sprintf(" [#] Smart Contract: %s", RENDERHIVE_TESTNET_SMART_CONTRACT))
  // Hive cycle
  logger.RenderhiveLogger.Main.Info().Msg(fmt.Sprintf(" [#] Hive Cycle Synchronization Topic: %s", RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_SYNCHRONIZATION))
  logger.RenderhiveLogger.Main.Info().Msg(fmt.Sprintf(" [#] Hive Cycle Application Topic: %s", RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_APPLICATION))
  logger.RenderhiveLogger.Main.Info().Msg(fmt.Sprintf(" [#] Hive Cycle Validation Topic: %s", RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_VALIDATION))
  // Render jobs
  logger.RenderhiveLogger.Main.Info().Msg(fmt.Sprintf(" [#] Render Job Topic: %s", RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_VALIDATION))


  // make sure the end of the program is logged
  defer logger.RenderhiveLogger.Main.Info().Msg("Renderhive service stopped.")



  // HEDERA
  // ***************************************************************************
  // initialize the Hedera Manager
  HederaManager, err := hedera.InitHederaManager(hedera.NETWORK_TYPE_TESTNET, "hedera/testnet.env")
  if err != nil {
    logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
    os.Exit(1)
  }

  // log the operator's public key
  logger.RenderhiveLogger.Main.Info().Msg("Loaded the account details from the environment file.")
  logger.RenderhiveLogger.Main.Info().Msg(fmt.Sprintf(" [#] Public key: %s",  HederaManager.Operator.PublicKey))



  // HEDERA SMART CONTRACT SERVICE CONTRACT
  // ***************************************************************************
  var response *hederasdk.TransactionResponse
  var receipt *hederasdk.TransactionReceipt
  logger.RenderhiveLogger.Main.Info().Msg("Creating the required Hedera Smart Contract ...")

  // define the renderhive version for the topic names
  const renderhive_contract_main_version = "0"
  const renderhive_contract_sub_version = "1"
  const renderhive_contract_patch_version = "0"

  // RENDERHIVE SMART CONTRACT
  if RENDERHIVE_TESTNET_SMART_CONTRACT == "" {

      // Create the smart contract
      logger.RenderhiveLogger.Main.Info().Msg(" [#] Renderhive smart contract:")

      // prepare the topic information, which are used to create the topic
      contract := hedera.HederaSmartContract{Info: hederasdk.ContractInfo{ContractMemo: fmt.Sprintf("renderhive-v%s.%s.%s::smart-contract", renderhive_contract_main_version, renderhive_contract_sub_version, renderhive_contract_patch_version), AdminKey: HederaManager.Operator.PublicKey}}
      response, receipt, err = contract.New(&HederaManager, "./RenderhiveTestContract.json", HederaManager.Operator.PrivateKey)
      if err != nil {
        logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
        os.Exit(1)
      }
      if receipt != nil {
        logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] [*] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
      }

      // get contract information
      logger.RenderhiveLogger.Package["hedera"].Debug().Msg(" [#] [*] Query the topic information ...")
      contractID, err := hederasdk.ContractIDFromString(contract.ID.String())
      if err != nil {
        logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
        os.Exit(1)
      }
      contract = hedera.HederaSmartContract{ID: contractID}
      _, err = contract.QueryInfo(&HederaManager)
      if err != nil {
        logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
        os.Exit(1)
      }
      logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] [*] Contract ID: %v", contract.ID))
      logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] [*] Contract Memo: %v", contract.Info.ContractMemo))


      // call a function of the contract
      // function parameters
      registerFunctionParams := hederasdk.NewContractFunctionParameters().
        AddString("TestUser")

      logger.RenderhiveLogger.Package["hedera"].Debug().Msg("Call the smart contract and register a new user:")
      response, receipt, err = contract.CallFunction(&HederaManager, "register", registerFunctionParams, 1000000)
      if err != nil {
        logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
        os.Exit(1)
      }
      if receipt != nil {
        logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
      }
      contract.GetEventLog(&HederaManager, response, "RegisteredUser")

      // wait
      time.Sleep(1 * time.Second)

      // unregister user again
      logger.RenderhiveLogger.Package["hedera"].Debug().Msg("Call the smart contract and delete a user:")
      response, receipt, err = contract.CallFunction(&HederaManager, "unregister", nil, 1000000)
      if err != nil {
        logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
        os.Exit(1)
      }
      if receipt != nil {
        logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
      }
      contract.GetEventLog(&HederaManager, response, "DeletedUser")

      // wait
      time.Sleep(1 * time.Second)

      // function parameters
      registerFunctionParams2 := hederasdk.NewContractFunctionParameters().
        AddString("Mr. Poe")

      logger.RenderhiveLogger.Package["hedera"].Debug().Msg("Call the smart contract and register a new user:")
      response, receipt, err = contract.CallFunction(&HederaManager, "register", registerFunctionParams2, 1000000)
      if err != nil {
        logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
        os.Exit(1)
      }
      if receipt != nil {
        logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
      }
      contract.GetEventLog(&HederaManager, response, "RegisteredUser")

      // wait
      time.Sleep(10 * time.Second)

      // Delete the created contract
      logger.RenderhiveLogger.Package["hedera"].Debug().Msg("Delete the created contract again")
      response, receipt, err = contract.Delete(&HederaManager, nil)
      if err != nil {
        logger.RenderhiveLogger.Package["hedera"].Error().Err(err).Msg("")
        os.Exit(1)
      }
      if receipt != nil {
        logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Receipt: %s (Status: %s)", receipt.TransactionID.String(), receipt.Status))
      }

  }


  // COMMAND LINE INTERFACE
  // ***************************************************************************
  // start the command line interface tool
  cli_tool := cli.CommandLine{}
  cli_tool.Start()

}

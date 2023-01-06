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
  // hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

  // internal
  . "renderhive/constants"
  "renderhive/logger"
  "renderhive/cli"
  //"renderhive/hedera"
  //"renderhive/node"
)


// APP
// #############################################################################
// INITIALIZE APP
func init() {

  // LOGGER SYSTEM
  // ***************************************************************************
  // initialize the logger system
  logger.Init()

  // add the package loggers
  logger.AddPackageLogger("node")
  logger.AddPackageLogger("hedera")
  logger.AddPackageLogger("ipfs")
  logger.AddPackageLogger("renderer")

}


// MAIN LOOP
func main () {

  // prepare end of program
  defer os.Exit(0)

  // error value
  var err error

  // placeholder
  fmt.Println(time.Now().Add(30 * time.Second))


  // BASIC INFORMATION
  // ***************************************************************************
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



  // INITIALIZE SERVICE APP
  // ***************************************************************************
  ServiceApp := ServiceApp{}
  err = ServiceApp.Init()
  if err != nil {
    logger.RenderhiveLogger.Main.Error().Err(err).Msg("")
    os.Exit(1)
  }



  // COMMAND LINE INTERFACE
  // ***************************************************************************
  // start the command line interface tool
  cli_tool := cli.CommandLine{}
  cli_tool.Start()

}

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
  "os"

  "rendera/logger"
  "rendera/cli"
  "rendera/hedera"
)

// COMPILER FLAGS
// #############################################################################
const COMPILER_HEDERA_TESTNET = true

// MAIN LOOP
// #############################################################################
func main () {

  // prepare end of program
  defer os.Exit(0)

  // error value
  var err error

  // LOGGER SYSTEM
  // ***************************************************************************
  // initialize the logger system
  logger.Init()

  // add the package loggers
  logger.AddPackageLogger("node")
  logger.AddPackageLogger("hedera")
  logger.AddPackageLogger("ipfs")

  // log the start of the rendera service
  logger.RenderaLogger.Main.Info().Msg("Rendera service started.")

  // make sure the end of the program is logged
  defer logger.RenderaLogger.Main.Info().Msg("Rendera service stopped.")


  // COMMAND LINE INTERFACE
  // ***************************************************************************
  // start the command line interface tool
  cli_tool := cli.CommandLine{}
  cli_tool.Start()



  // HEDERA TESTNET
  // ***************************************************************************
  // initialize the Hedera testnet, if required
  if COMPILER_HEDERA_TESTNET {
    hederaTestnet := hedera.HederaTestnet{}
    err = hederaTestnet.Init()
    if err != nil {
      logger.RenderaLogger.Package["hedera"].Error().Err(err).Msg("")
      os.Exit(1)
    }

    // create a new account
    err = hederaTestnet.CreateAccount()
    if err != nil {
      logger.RenderaLogger.Package["hedera"].Error().Err(err).Msg("")
      os.Exit(1)
    }
  }

}

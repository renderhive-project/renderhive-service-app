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
  "fmt"
  "os"

  "rendera/cli"
  "rendera/hedera"
)

// COMPILER FLAGS
// #############################################################################
const COMPILER_HEDERA_TESTNET = true

// MAIN LOOP
// #############################################################################
func main () {
  defer os.Exit(0)

  fmt.Println("[INFO] Rendera service started.")

  // initialize the Hedera testnet, if required
  if COMPILER_HEDERA_TESTNET {
    hederaTestnet := hedera.HederaTestnet{}
    err := hederaTestnet.Init()
    if err != nil {
      fmt.Println("[HEDERA]", "ERROR:", err)
      os.Exit(1)
    }

    // Print the testnet account ID and private key to the console
    fmt.Printf("[HEDERA] The account ID is: %v\n", hederaTestnet.AccountID)
    fmt.Printf("[HEDERA] The private key is: %v\n", hederaTestnet.PrivateKey)

  }

  // start the command line interface tool
	cli_tool := cli.CommandLine{}
	cli_tool.Start()

  fmt.Println("[INFO] Rendera service stopped.")
}

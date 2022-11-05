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

package hedera

/*

The hedera package handles all functions related to the Hedera Hashgraph
services. This also includes a crypto wallet with very basic functionality.

*/

import (
  "fmt"
  "os"

  "github.com/joho/godotenv"
  hederasdk "github.com/hashgraph/hedera-sdk-go/v2"
)

// empty structure to hold all methods
type HederaTestnet struct {
  AccountID hederasdk.AccountID
  PrivateKey hederasdk.PrivateKey
  Client *hederasdk.Client
}

// empty structure to hold all methods
type HederaServices struct {}


// TESTNET INTEGRATIONS
// #############################################################################
// Initialize everything required for communication with the Hedera Testnet
func (testnet *HederaTestnet) Init() error {

    fmt.Println("[HEDERA] Initializing the testnet account ...")

    // Loads the .env file and throws an error if it cannot load the variables
    // from that file correctly
    err := godotenv.Load("hedera/testnet.env")
    if err != nil {
        return err
    }

    // Grab your testnet account ID and private key from the .env file
    testnet.AccountID, err = hederasdk.AccountIDFromString(os.Getenv("TESTNET_ACCOUNT_ID"))
    if err != nil {
        return err
    }

    testnet.PrivateKey, err = hederasdk.PrivateKeyFromString(os.Getenv("TESTNET_PRIVATE_KEY"))
    if err != nil {
        return err
    }

    // Create your testnet client
    testnet.Client = hederasdk.ClientForTestnet()
    testnet.Client.SetOperator(testnet.AccountID, testnet.PrivateKey)

    return nil
}

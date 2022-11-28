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

  // standard
  "fmt"

  // external
  hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

  // internal
  "rendera/logger"

)

// define the network types
const (
  NETWORK_TYPE_TESTNET = iota
  NETWORK_TYPE_PREVIEWNET = iota
  NETWORK_TYPE_MAINNET = iota
)

// empty structure to hold all methods for the Hedera interactions
type HederaManager struct {

  // network communication
  NetworkType int
  NetworkClient *hederasdk.Client

  // account of the user
  Operator HederaAccount

}

// HEDERA MANAGER
// #############################################################################
// Initialize everything required for communication with the Hedera network
func InitHederaManager(NetworkType int, AccountFilePath string) (HederaManager, error) {
    var err error

    // create a new account
    Account := HederaAccount{}

    switch NetworkType {
    case NETWORK_TYPE_TESTNET:

        // log information
        logger.RenderaLogger.Package["hedera"].Info().Msg("Initializing on Hedera Testnet:")

        // Create your testnet client
        NetworkClient := hederasdk.ClientForTestnet()

        // get the testnet account information from file
        logger.RenderaLogger.Package["hedera"].Info().Msg(" [#] Load account information from encrypted file.")
        Account.FromFile(AccountFilePath)

        // create the Hedera manager
        m := HederaManager{NetworkType, NetworkClient, Account}

        // log the testnet account ID and private key to the console
        logger.RenderaLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Account ID: %v", Account.AccountID))
        // logger.RenderaLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Private key: %v", Account.PrivateKey))
        // logger.RenderaLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Public key: %v", Account.PublicKey))

        // set theis account as the operator
        NetworkClient.SetOperator(Account.AccountID, Account.PrivateKey)

        // query the complete account information from the Hedera network
        queryCost, err := Account.QueryInfo(&m)
        logger.RenderaLogger.Package["hedera"].Info().Msg(fmt.Sprintf(" [#] Account Balance: %v", Account.Info.Balance))
        logger.RenderaLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Costs (QueryInfo): %v", queryCost))

        // query the account balance from the Hedera network
        queryCost, err = Account.QueryBalance(&m)
        logger.RenderaLogger.Package["hedera"].Info().Msg(fmt.Sprintf(" [#] Account Balance: %v", Account.Info.Balance))
        logger.RenderaLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Costs (QueryBalance): %v", queryCost))

        // return the initialized Hedera manager
        return m, err

    case NETWORK_TYPE_PREVIEWNET:

        // log information
        logger.RenderaLogger.Package["hedera"].Debug().Msg("Initializing on Hedera Previewnet:")

        // return the initialized Hedera manager
        return HederaManager{}, err

    case NETWORK_TYPE_MAINNET:

        // log information
        logger.RenderaLogger.Package["hedera"].Debug().Msg("Initializing on Hedera Mainnet:")

        // return the initialized Hedera manager
        return HederaManager{}, err

    default:

        // return the initialized Hedera manager
        return HederaManager{}, err
    }
}

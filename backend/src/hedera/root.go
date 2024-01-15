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

/*

The hedera package handles all functions related to the Hedera Hashgraph
services. This also includes a crypto wallet with very basic functionality.

*/

import (

	// standard
	"fmt"
	"path/filepath"
	"strings"
	"time"

	// external
	hederasdk "github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/spf13/cobra"

	// internal
	. "renderhive/globals"
	"renderhive/logger"
)

// define the network types
const (
	NETWORK_TYPE_TESTNET    = iota
	NETWORK_TYPE_PREVIEWNET = iota
	NETWORK_TYPE_MAINNET    = iota
)

// empty structure to hold all methods for the Hedera interactions
type PackageManager struct {

	// network communication
	NetworkType   int
	NetworkClient *hederasdk.Client

	// Hedera account of this node
	Operator HederaAccount

	// Mirror Node
	MirrorNode MirrorNode

	// Command line interface
	Command      *cobra.Command
	CommandFlags struct {
		FlagPlaceholder bool
	}
}

// HEDERA MANAGER
// #############################################################################
// create the hedera manager variable
var Manager = PackageManager{}

// Initialize everything required for communication with the Hedera network
func (hm *PackageManager) Init(NetworkType int) error {
	var err error

	logger.Manager.Package["hedera"].Debug().Msg("Initializing the Hedera manager ...")

	switch NetworkType {
	case NETWORK_TYPE_TESTNET:

		// log information
		logger.Manager.Package["hedera"].Info().Msg(" [#] Initializing on Hedera Testnet ...")

		// Create your testnet client
		hm.NetworkClient = hederasdk.ClientForTestnet()

	case NETWORK_TYPE_PREVIEWNET:

		// log information
		logger.Manager.Package["hedera"].Debug().Msg("Initializing on Hedera Previewnet:")

		// Create your preview client
		hm.NetworkClient = hederasdk.ClientForPreviewnet()

	case NETWORK_TYPE_MAINNET:

		// log information
		logger.Manager.Package["hedera"].Debug().Msg("Initializing on Hedera Mainnet:")

		// Create your preview client
		hm.NetworkClient = hederasdk.ClientForPreviewnet()

	}

	// set network type
	hm.NetworkType = NetworkType

	// get the mirror node URL
	hm.MirrorNode.URL = HEDERA_TESTNET_MIRROR_NODE_URL

	// log info
	logger.Manager.Main.Info().Msg(fmt.Sprintf(" [#] Mirror node: %v", hm.MirrorNode.URL))

	return err
}

// Load the account from the local data
func (hm *PackageManager) LoadAccount(account_id string, passphrase string, publickey string) error {
	var err error

	// read the node account ID into the node manager
	hm.Operator.AccountID, err = hederasdk.AccountIDFromString(account_id)
	if err != nil {
		return err
	}

	// log info
	logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Account ID: %v", hm.Operator.AccountID))

	// TODO: From this point on, the private key is in memory in clear text
	//		 This needs to be improved from a security standpoint!!!

	// read the private key from the keystore file and decrypt it
	err = hm.Operator.FromFile(filepath.Join(RENDERHIVE_APP_DIRECTORY_CONFIG, strings.ReplaceAll(account_id, ".", "")+".key"), passphrase, publickey)
	if err != nil {
		return err
	}

	// log info
	logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Public key: %v", hm.Operator.PublicKey))

	// set this account as the operator
	hm.NetworkClient.SetOperator(hm.Operator.AccountID, hm.Operator.PrivateKey)

	// // query the account balance from the Hedera network
	// queryCost, err := hm.Operator.QueryBalance(hm)
	// logger.Manager.Package["hedera"].Info().Msg(fmt.Sprintf(" [#] Account Balance: %v", hm.Operator.Info.Balance))
	// logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Costs (QueryBalance): %v", queryCost))

	// query the complete account information from the Hedera network
	queryCost, err := hm.Operator.QueryInfo(hm)
	logger.Manager.Package["hedera"].Info().Msg(fmt.Sprintf(" [#] Account Balance: %v", hm.Operator.Info.Balance))
	logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Costs (QueryInfo): %v", queryCost))

	return err
}

// Deinitialize the Hedera manager
func (hm *PackageManager) DeInit() error {
	var err error

	// log event
	logger.Manager.Package["hedera"].Debug().Msg("Deinitializing the Hedera manager ...")

	return err

}

// TOPIC MANAGEMENT
// #############################################################################
// Obtain the topic information from a TopicID given in string format
func (hm *PackageManager) TopicInfoFromString(topicID string) (*HederaTopic, error) {
	var err error

	logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf("Query topic information for TopicID string '%v':", topicID))
	// get the topic ID from a string
	hTopicID, err := hederasdk.TopicIDFromString(topicID)
	if err != nil {
		return nil, err
	}

	// create a HederaTopic variable and query the information
	topic := HederaTopic{ID: hTopicID}
	_, err = topic.QueryInfo(hm)
	if err != nil {
		return nil, err
	}
	logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Topic ID: %v", topic.ID))
	logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Topic Memo: %v", topic.Info.TopicMemo))

	return &topic, nil
}

// Subscribe to the topic
func (hm *PackageManager) TopicSubscribe(topic *HederaTopic, startTime time.Time, onNext func(message hederasdk.TopicMessage)) error {
	var err error

	logger.Manager.Package["hedera"].Debug().Msg(fmt.Sprintf("Subscribe to topic with ID %v.", topic.ID))

	// subscribe to the topic
	err = topic.Subscribe(startTime, onNext)
	if err != nil {
		return err
	}

	return err
}

// HEDERA MANAGER COMMAND LINE INTERFACE
// #############################################################################
// Create the command for the command line interface
func (hm *PackageManager) CreateCommand() *cobra.Command {

	// create the package command
	hm.Command = &cobra.Command{
		Use:   "hedera",
		Short: "Commands for the interaction with the Hedera services",
		Long:  "This command and its sub-commands enable the interaction with the Hedera services required by the Renderhive network",
		Run: func(cmd *cobra.Command, args []string) {

			return

		},
	}

	return hm.Command

}

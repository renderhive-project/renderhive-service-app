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

This file wraps all the required REST API calls in order to communicate with
the Hedera mirror nodes.

*/

import (

	// standard
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	// external
	// hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

	// internal
	"renderhive/logger"
)

// Mirror node information and calls
type MirrorNode struct {

	// Mirror node url
	URL string
}

// Response structures
type AccountInfoResponse struct {
	Accounts []AccountInfo `json:"accounts"`
	Links    struct {
		Next string `json:"next"`
	} `json:"links"`
}

type AccountInfo struct {
	Account         string `json:"account"`
	Alias           string `json:"alias"`
	AutoRenewPeriod int64  `json:"auto_renew_period"`
	Balance         struct {
		Balance   int64         `json:"balance"`
		Timestamp string        `json:"timestamp"`
		Tokens    []interface{} `json:"tokens"`
	} `json:"balance"`
	CreatedTimestamp string `json:"created_timestamp"`
	DeclineReward    bool   `json:"decline_reward"`
	Deleted          bool   `json:"deleted"`
	EthereumNonce    int    `json:"ethereum_nonce"`
	EvmAddress       string `json:"evm_address"`
	ExpiryTimestamp  string `json:"expiry_timestamp"`
	Key              struct {
		Type string `json:"_type"`
		Key  string `json:"key"`
	} `json:"key"`
	MaxAutomaticTokenAssociations int    `json:"max_automatic_token_associations"`
	Memo                          string `json:"memo"`
	PendingReward                 int    `json:"pending_reward"`
	ReceiverSigRequired           bool   `json:"receiver_sig_required"`
	StakedAccountID               string `json:"staked_account_id"`
	StakedNodeID                  int    `json:"staked_node_id"`
	StakePeriodStart              string `json:"stake_period_start"`
}

// Response structures
type TransactionInfoResponse struct {
	Transactions []TransactionInfo `json:"transactions"`
	Links        struct {
		Next string `json:"next"`
	} `json:"links"`
}

type TransactionInfo struct {
	Bytes                    interface{} `json:"bytes"`
	ChargedTxFee             int         `json:"charged_tx_fee"`
	ConsensusTimestamp       string      `json:"consensus_timestamp"`
	EntityID                 string      `json:"entity_id"`
	MaxFee                   string      `json:"max_fee"`
	MemoBase64               interface{} `json:"memo_base64"`
	Name                     string      `json:"name"`
	Node                     string      `json:"node"`
	Nonce                    int         `json:"nonce"`
	ParentConsensusTimestamp string      `json:"parent_consensus_timestamp"`
	Result                   string      `json:"result"`
	Scheduled                bool        `json:"scheduled"`
	StakingRewardTransfers   []struct {
		Account int `json:"account"`
		Amount  int `json:"amount"`
	} `json:"staking_reward_transfers"`
	TransactionHash string `json:"transaction_hash"`
	TransactionID   string `json:"transaction_id"`
	TokenTransfers  []struct {
		TokenID    string `json:"token_id"`
		Account    string `json:"account"`
		Amount     int    `json:"amount"`
		IsApproval bool   `json:"is_approval"`
	} `json:"token_transfers"`
	Transfers []struct {
		Account    string `json:"account"`
		Amount     int    `json:"amount"`
		IsApproval bool   `json:"is_approval"`
	} `json:"transfers"`
	ValidDurationSeconds string `json:"valid_duration_seconds"`
	ValidStartTimestamp  string `json:"valid_start_timestamp"`
}

// MIRROR NODE API
// #############################################################################
// Query account information
// https://mainnet-public.mirrornode.hedera.com/api/v1/accounts?account.id=x.x.xxxx
func (m *MirrorNode) GetAccountInfo(accountID string, limit int, order string) (*[]AccountInfo, error) {
	var err error
	var command []string
	var parameters []string

	// log query
	logger.Manager.Package["hedera"].Trace().Msg("Query account information:")

	// prepare the base command
	command = append(command, m.URL, "api", "v1", "accounts?")

	// prepare the parameters
	if accountID != "" {
		parameters = append(parameters, "account.id="+accountID)
	}
	if limit > 0 {
		parameters = append(parameters, "limit="+strconv.Itoa(limit))
	}
	if order != "" {
		parameters = append(parameters, "order="+order)
	}

	// log the command
	logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Command: %v", strings.Join(command, "/")+strings.Join(parameters, "&")))

	// query the transaction list
	httpResponse, err := http.Get(strings.Join(command, "/") + strings.Join(parameters, "&"))
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	// read the complete data
	httpResponseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	// parse the transaction response
	var AccountInfoResponse AccountInfoResponse
	json.Unmarshal(httpResponseBody, &AccountInfoResponse)

	// log number of transactions
	logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Mirror node responded with %v accounts", len(AccountInfoResponse.Accounts)))

	return &AccountInfoResponse.Accounts, err

}

// Query a list of transactions
// https://mainnet-public.mirrornode.hedera.com/api/v1/transactions?order=desc&limit=1
func (m *MirrorNode) Transactions(accountID string, limit int, order string, transactionType string, result string, balanceType string) (*[]TransactionInfo, error) {
	var err error
	var command []string
	var parameters []string

	// log query
	logger.Manager.Package["hedera"].Trace().Msg("Query a list of transactions:")

	// prepare the base command
	command = append(command, m.URL, "api", "v1", "transactions?")

	// prepare the parameters
	if limit > 0 {
		parameters = append(parameters, "limit="+strconv.Itoa(limit))
	}
	if order != "" {
		parameters = append(parameters, "order="+order)
	}
	if transactionType != "" {
		parameters = append(parameters, "transactiontype="+transactionType)
	}
	if result != "" {
		parameters = append(parameters, "result="+result)
	}
	if balanceType != "" {
		parameters = append(parameters, "type="+balanceType)
	}

	// log the command
	logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Command: %v", strings.Join(command, "/")+strings.Join(parameters, "&")))

	// query the transaction list
	httpResponse, err := http.Get(strings.Join(command, "/") + strings.Join(parameters, "&"))
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	// read the complete data
	httpResponseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	// parse the transaction response
	var TransactionResponse TransactionInfoResponse
	json.Unmarshal(httpResponseBody, &TransactionResponse)

	// log number of transactions
	logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Mirror node responded with %v transactions", len(TransactionResponse.Transactions)))

	return &TransactionResponse.Transactions, err

}

// Query a specific transaction ID
// https://mainnet-public.mirrornode.hedera.com/api/v1/transactions/${transactionID}
func (m *MirrorNode) GetTransactionInfo(transactionID string) (*TransactionInfo, error) {
	var err error
	var command []string

	// log query
	logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf("Query the transaction with Id: %v", transactionID))

	// bring transactionID into the coorect form
	parts := strings.Split(transactionID, "@")
	parts[1] = strings.ReplaceAll(parts[1], ".", "-")
	transactionID = parts[0] + "-" + parts[1]

	// prepare the base command
	command = append(command, m.URL, "api", "v1", "transactions", transactionID)

	// log the command
	logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Command: %v", strings.Join(command, "/")))

	// query the transaction
	httpResponse, err := http.Get(strings.Join(command, "/"))
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	// read the complete data
	httpResponseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	// parse the transaction response
	var TransactionResponse TransactionInfoResponse
	err = json.Unmarshal(httpResponseBody, &TransactionResponse)
	if err != nil {
		return nil, err
	}

	// log number of transactions
	logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Mirror node responded with %v transactions", len(TransactionResponse.Transactions)))

	// parse the transaction response
	if len(TransactionResponse.Transactions) > 0 {

		// get the transaction information
		TransactionInfo := TransactionResponse.Transactions[0]

		// log number of transactions
		logger.Manager.Package["hedera"].Trace().Msg(fmt.Sprintf(" [#] Mirror node responded with: %v", TransactionInfo))

		return &TransactionInfo, err
	}

	return nil, fmt.Errorf("transaction not found")
}

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

package hedera

/*

This file wraps all the required REST API calls in order to communicate with
the Hedera mirror nodes.

*/

import (

  // standard
  "fmt"
  "strconv"
  "strings"
  "net/http"
  "io/ioutil"
  "encoding/json"
  "time"

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
type TransactionResponse struct {
  Transactions []struct {
		Bytes                   interface{}   `json:"bytes"`
		ChargedTxFee            int           `json:"charged_tx_fee"`
		ConsensusTimestamp      string        `json:"consensus_timestamp"`
		EntityID                string        `json:"entity_id"`
		MaxFee                  int           `json:"max_fee"`
		MemoBase64              interface{}   `json:"memo_base64"`
		Name                    string        `json:"name"`
		Node                    string        `json:"node"`
		Nonce                   int           `json:"nonce"`
		ParentConsensusTimestamp string        `json:"parent_consensus_timestamp"`
		Result                  string        `json:"result"`
		Scheduled               bool          `json:"scheduled"`
		StakingRewardTransfers  []struct {
			Account int `json:"account"`
			Amount  int `json:"amount"`
		} `json:"staking_reward_transfers"`
		TransactionHash      string        `json:"transaction_hash"`
		TransactionID        string        `json:"transaction_id"`
		TokenTransfers       []struct {
			TokenID     string  `json:"token_id"`
			Account     string  `json:"account"`
			Amount      int     `json:"amount"`
			IsApproval  bool    `json:"is_approval"`
		} `json:"token_transfers"`
		Transfers []struct {
			Account    string  `json:"account"`
			Amount     int     `json:"amount"`
			IsApproval bool    `json:"is_approval"`
		} `json:"transfers"`
		ValidDurationSeconds    int     `json:"valid_duration_seconds"`
		ValidStartTimestamp     string  `json:"valid_start_timestamp"`
	} `json:"transactions"`
	Links struct {
		Next string `json:"next"`
	} `json:"links"`
}

// MIRROR NODE API
// #############################################################################
// Query a list of transactions
// https://mainnet-public.mirrornode.hedera.com/api/v1/transactions?order=desc&limit=1
func (m *MirrorNode) Transactions(hm *HederaManager, accountID string, limit int, order string, transactionType string, result string, balanceType string) (*TransactionResponse, error) {
    var err error
    var command []string
    var parameters []string

    // log query
    logger.RenderhiveLogger.Package["hedera"].Debug().Msg("Query a list of transactions:")

    // prepare the base command
    command = append(command, m.URL, "api", "v1", "transactions?")

    // prepare the parameters
    if limit > 0              { parameters = append(parameters, "limit=" + strconv.Itoa(limit)) }
    if order != ""            { parameters = append(parameters, "order=" + order) }
    if transactionType != ""  { parameters = append(parameters, "transactiontype=" + transactionType) }
    if result != ""           { parameters = append(parameters, "result=" + result) }
    if balanceType != ""      { parameters = append(parameters, "type=" + balanceType) }

    // log the command
    logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Command: %v", strings.Join(command, "/") + strings.Join(parameters, "&")))

    // query the transaction list
    httpResponse, err := http.Get(strings.Join(command, "/") + strings.Join(parameters, "&"))
    if err != nil {
        return nil, err
    }
    defer httpResponse.Body.Close()

    // read the complete data
    httpResponseBody, err := ioutil.ReadAll(httpResponse.Body)
    if err != nil {
        return nil, err
    }

    // parse the transaction response
    var TransactionResponse TransactionResponse
    json.Unmarshal(httpResponseBody, &TransactionResponse)

    // log number of transactions
    logger.RenderhiveLogger.Package["hedera"].Debug().Msg(fmt.Sprintf(" [#] Mirror node responded with %v transactions", len(TransactionResponse.Transactions)))

    return &TransactionResponse, err

}

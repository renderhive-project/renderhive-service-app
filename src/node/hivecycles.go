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

package node

/*

  This file contains all structures and functions required to handle the hive
  cycle calculations and to keep the node in synchronization with all the other
  nodes.

*/

import (

  // standard
  "fmt"
  "encoding/json"
  // "os"
  "time"

  // external
  hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

  // internal
  "renderhive/logger"
  // "renderhive/constants"
  // "renderhive/hedera"
)

// Structure to manage the hive cycle
type HiveCycle struct {

  Configurations []HiveCycleConfigurationMessage
  Current int

}

// Define a struct to represent the JSON message in the Hive Cycle Synchronization Topic
type HiveCycleConfigurationMessage struct {
    Iteration   int `json:"iteration"`       // hive cycle iteration (how many times was the configuration changed)
    Duration    int `json:"duration"`        // hive cycle duration in seconds
    Timestamp   time.Time `json:"timestamp"` // timestamp of the creator's clock when creating the message
}


// HIVE CYCLE MANAGEMENT
// #############################################################################
// Get the most recent hive cycle configuration from a mirror node
func (hc *HiveCycle) MessageCallback() (func(message hederasdk.TopicMessage)) {

    return func(message hederasdk.TopicMessage) {
      var err error

      //
      // Import and parse the compiled contract from the contract file
    	jsonData := message.Contents

      // Parse the HiveCycleConfigurationMessage from the JSON string
      var configuration HiveCycleConfigurationMessage
    	err = json.Unmarshal(jsonData, &configuration)
    	if err != nil {

        logger.RenderhiveLogger.Package["hedera"].Info().Msg(fmt.Sprintf("Message received but not processed: %s", string(message.Contents)))
        return

    	}

      // add the message to the array of configuration messages
      hc.Configurations = append(hc.Configurations, configuration)

      // use the consensus timestamp of the message
      hc.Configurations[len(hc.Configurations) - 1].Timestamp = message.ConsensusTimestamp

      return

    }

}

// Synchronize with the Hedera network consensus time and calculate the current
// hive cycle of the network
func (hc *HiveCycle) Synchronize() (error) {
    var err error

    // log information
    logger.RenderhiveLogger.Package["node"].Info().Msg("Synchronize with hive cycle")

    // TODO: check the current Hedera network consensus time by looking at the last
    // transaction that reached consensus in the network


    // iterate through all configurations messages to calculate the hive cycle
    for i, configuration := range hc.Configurations {

      logger.RenderhiveLogger.Package["node"].Info().Msg(fmt.Sprintf("Obtained new configuration message (no. %v):", i))
      logger.RenderhiveLogger.Package["node"].Info().Msg(fmt.Sprintf(" [#] Iteration: %v", configuration.Iteration))
      logger.RenderhiveLogger.Package["node"].Info().Msg(fmt.Sprintf(" [#] Duration: %v", configuration.Duration))
      logger.RenderhiveLogger.Package["node"].Info().Msg(fmt.Sprintf(" [#] Timestamp: %v", configuration.Timestamp))

    }

    // TODO: Validate that the message was from a valid source.
    //       This could be the admin account ID. But maybe it is not necessary
    //       at all, if we use a SubmitKey in the topic?

    return err

}

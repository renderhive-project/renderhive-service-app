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
  "math"
  "time"

  // external
  hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

  // internal
  "renderhive/logger"
  // "renderhive/constants"
  "renderhive/hedera"
)

// structure for the time synchronization
type HiveClock struct {

  NetworkTime time.Time
  LocalTime time.Time
  Difference time.Duration

}

// Structure to manage the hive cycle
type HiveCycle struct {

  Configurations []HiveCycleConfigurationMessage
  NetworkClock HiveClock
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

      // TODO: Validate that the message was from a valid source.
      //       This could be the admin account ID. But maybe it is not necessary
      //       at all, if we use a SubmitKey in the topic?

      // add the message to the array of configuration messages
      hc.Configurations = append(hc.Configurations, configuration)

      // use the consensus timestamp of the message
      hc.Configurations[len(hc.Configurations) - 1].Timestamp = message.ConsensusTimestamp

      return

    }

}

// Synchronize with the Hedera network consensus time and calculate the current
// hive cycle of the network
func (hc *HiveCycle) Synchronize(hm *hedera.HederaManager) (error) {
    var err error
    var transactions *[]hedera.TransactionInfo

    // Get the last transaction on the Hedera mirror node
    transactions, err = hm.MirrorNode.Transactions(hm, "", 1, "desc", "", "", "")
    if err != nil {
      return err
    }

    // local time
    hc.NetworkClock.LocalTime = time.Now()

    // Parse the duration represented by the input string
    duration, err := time.ParseDuration((*transactions)[0].ConsensusTimestamp + "s")
    if err != nil {
        return err
    }

    // Add the duration to the Unix epoch to obtain a time.Time value
    hc.NetworkClock.NetworkTime = time.Unix(0, 0).Add(duration)

    // calculate the difference between the local node time
    hc.NetworkClock.Difference = hc.NetworkClock.LocalTime.Sub(hc.NetworkClock.NetworkTime)

    // reset hive cycle value
    hc.Current = 0

    // iterate through all configurations messages to calculate the current
    // hive cycle
    for _, configuration := range hc.Configurations {

      // if there is more than one configuration message
      if len(hc.Configurations) > 1 {

          // calculate the hive cycles in this iteration (i)
          hc.Current += int(math.Ceil(float64(hc.NetworkClock.NetworkTime.Sub(configuration.Timestamp) / (time.Duration(configuration.Duration) * time.Second))))

      } else {

          // calculate the hive cycles
          hc.Current += int(math.Ceil(float64(hc.NetworkClock.NetworkTime.Sub(configuration.Timestamp) / (time.Duration(configuration.Duration) * time.Second))))

      }

      logger.RenderhiveLogger.Package["node"].Debug().Msg(fmt.Sprintf("Configuration message (iteration: %v):", configuration.Iteration))
      logger.RenderhiveLogger.Package["node"].Debug().Msg(fmt.Sprintf(" [#] Consensus time: %v", configuration.Timestamp))
      logger.RenderhiveLogger.Package["node"].Debug().Msg(fmt.Sprintf(" [#] Duration: %v", configuration.Duration))

    }

    // log information
    logger.RenderhiveLogger.Package["node"].Debug().Msg("Synchronized with HCS time and calculated hive cycle:")
    logger.RenderhiveLogger.Package["node"].Debug().Msg(fmt.Sprintf(" [#] Consensus time: %v", hc.NetworkClock.NetworkTime))
    logger.RenderhiveLogger.Package["node"].Debug().Msg(fmt.Sprintf(" [#] Difference to local time: %v", hc.NetworkClock.Difference))
    logger.RenderhiveLogger.Package["node"].Debug().Msg(fmt.Sprintf(" [#] Current hive cycle: %v", hc.Current))

    return err

}

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
	"encoding/json"
	"fmt"

	// "os"
	"errors"
	"math"
	"time"

	// external
	hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

	// internal
	. "renderhive/globals"
	"renderhive/hedera"
	"renderhive/logger"
)

// structure for the time synchronization
type HiveClock struct {

	// Time on the Hedera (mirror node) network and this Renderhive node during
	// the last synchronization
	NetworkTime time.Time
	LocalTime   time.Time
	Difference  time.Duration

	// Start time of the current cycle
	NetworkStartTime time.Time
	LocalStartTime   time.Time
}

// Structure to manage the hive cycle
type HiveCycle struct {
	Configurations []HiveCycleConfigurationMessage
	Clock          HiveClock
	Current        int
	LastSyncTime   time.Time
}

// Define a struct to represent the JSON message in the Hive Cycle Synchronization Topic
type HiveCycleConfigurationMessage struct {
	Iteration int       `json:"iteration"` // hive cycle iteration (how many times was the configuration changed)
	Duration  int       `json:"duration"`  // hive cycle duration in seconds
	Timestamp time.Time `json:"timestamp"` // consensus timestamp of this configuration message
}

// HIVE CYCLE MANAGEMENT
// #############################################################################
// Get the most recent hive cycle configuration from a mirror node
func (hc *HiveCycle) MessageCallback() func(message hederasdk.TopicMessage) {

	return func(message hederasdk.TopicMessage) {
		var err error

		//
		// Import and parse the compiled contract from the contract file
		jsonData := message.Contents

		// Parse the HiveCycleConfigurationMessage from the JSON string
		var configuration HiveCycleConfigurationMessage
		err = json.Unmarshal(jsonData, &configuration)
		if err != nil {

			logger.Manager.Package["hedera"].Error().Msg(fmt.Sprintf("Message received but not processed: %s", string(message.Contents)))
			return

		}

		// TODO: Validate that the message was from a valid source.
		//       This could be the admin account ID. But maybe it is not necessary
		//       at all, if we use a SubmitKey in the topic?

		// add the message to the array of configuration messages
		hc.Configurations = append(hc.Configurations, configuration)

		// use the consensus timestamp of the message
		hc.Configurations[len(hc.Configurations)-1].Timestamp = message.ConsensusTimestamp

		return

	}

}

// Synchronize with the Hedera network consensus time and calculate the current
// hive cycle of the network
func (hc *HiveCycle) Synchronize(hm *hedera.PackageManager) error {
	var err error
	var transactions *[]hedera.TransactionInfo

	// Only synchronize with the mirror node once per hour
	// NOTE: We use a local time between the synchronizations to lower the amount
	//       of mirror node calls
	if hc.LastSyncTime.IsZero() || time.Now().Sub(hc.LastSyncTime) > RENDERHIVE_CONFIG_HIVE_CYCLE_SYNCHRONIZATION_INTERVAL {

		// TODO: Check the following approach
		// Instead of using all the mirror node transactions and querying the
		// mirror node on a regular basis, we could just subscribe to the job
		// queue topic. Whenever a new message comes in (e.g., a new render job,
		// a render job status update, etc.) each node will check the consensus
		// time of this transaction and update the hive cycle based on this.
		// When there is no network activitiy (e.g., because all nodes are busy)
		// we will not query the mirror node network, because it makes no sense.
		//
		// Problem:
		// - What if a new node goes online, which could make processing of a
		//   previously skipped render job possible?

		// Get the last transaction on the Hedera mirror node
		transactions, err = hm.MirrorNode.Transactions("", 1, "desc", "", "", "")
		if err != nil {
			return err
		}

		// local time
		hc.Clock.LocalTime = time.Now()

		// Parse the duration represented by the input string
		duration, err := time.ParseDuration((*transactions)[0].ConsensusTimestamp + "s")
		if err != nil {
			return err
		}

		// Add the duration to the Unix epoch to obtain a time.Time value
		hc.Clock.NetworkTime = time.Unix(0, 0).Add(duration)

		// calculate the difference between the local node time and the network time
		hc.Clock.Difference = hc.Clock.LocalTime.Sub(hc.Clock.NetworkTime)

		// remember the time of the last synchronization
		hc.LastSyncTime = time.Now()

		// log trace event
		logger.Manager.Package["node"].Trace().Msg("Synchronized with mirror node consensus time.")

	} else {

		// local time
		hc.Clock.LocalTime = time.Now()

		// network time
		hc.Clock.NetworkTime = hc.Clock.LocalTime.Add(-hc.Clock.Difference)

	}

	// reset hive cycle value
	oldCycle := hc.Current
	hc.Current = 0

	// reset start time to configuration message's consensus timestamp
	hc.Clock.NetworkStartTime = hc.Configurations[0].Timestamp

	// iterate through all configurations messages to calculate the current
	// hive cycle
	for _, configuration := range hc.Configurations {

		// log trace event
		logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf("Configuration message (iteration: %v):", configuration.Iteration))
		logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Consensus time: %v", configuration.Timestamp))
		logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Duration: %v", configuration.Duration))

		// if there is more than one configuration message
		if len(hc.Configurations) > 1 {

			// calculate the hive cycles in this iteration (i)
			hc.Current += int(math.Ceil(float64(hc.Clock.NetworkTime.Sub(configuration.Timestamp) / (time.Duration(configuration.Duration) * time.Second))))

		} else {

			// calculate the hive cycles
			hc.Current += int(math.Ceil(float64(hc.Clock.NetworkTime.Sub(configuration.Timestamp) / (time.Duration(configuration.Duration) * time.Second))))

		}

		// calculate the start time of this cycle (in Consensus Time)
		hc.Clock.NetworkStartTime = hc.Clock.NetworkStartTime.Add(time.Duration(hc.Current*configuration.Duration) * time.Second)

	}

	// calculate the start time of this hive cycle in the local time of the node
	hc.Clock.LocalStartTime = hc.Clock.NetworkStartTime.Add(hc.Clock.Difference)

	// log information
	logger.Manager.Package["node"].Trace().Msg("Synchronized with HCS time and calculated hive cycle:")
	logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Consensus time: %v", hc.Clock.NetworkTime))
	logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Difference to local time: %v", hc.Clock.Difference))
	logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf(" [#] Current hive cycle: %v", hc.Current))

	// if the hive cycle just changed
	if hc.Current != oldCycle {

		// log trace event
		logger.Manager.Package["node"].Trace().Msg(fmt.Sprintf("New hive cycle %v detected at consensus time %v / local time %v", hc.Current, hc.Clock.NetworkTime, hc.Clock.LocalTime))

		// Enter the hive cycle application phase
		err = hc.ApplicationPhase(hm)
		if err != nil {
			return err
		}

		// TODO: Enter the hive cycle distribution phase
		// ...

		// TODO: Enter hive cycle render contract phase
		// ...

		// TODO: Enter hive cycle validation phase
		// ...

		// TODO: Enter hive cycle claiming phase
		// ...

	}

	return err

}

// Enter the application phase for this hive cycle
func (hc *HiveCycle) ApplicationPhase(hm *hedera.PackageManager) error {
	var err error
	// var transactions *[]hedera.TransactionInfo

	// // Get the last transaction on the Hedera mirror node
	// transactions, err = hm.MirrorNode.Transactions("", 1, "desc", "", "", "")
	// if err != nil {
	//   return err
	// }
	//

	// if the node is busy rendering, skip this cycle
	if Manager.Renderer.Busy {
		return errors.New(fmt.Sprintf("The node is busy rendering and will skip the hive cycle %v.", hc.Current))
	}

	// TODO: Check if the hive cycle is still in the application phase
	// ...

	return err

}

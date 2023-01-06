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
  // "fmt"
  // "os"
  "time"

  // external
  // hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

  // internal
  // "renderhive/logger"
  // "renderhive/constants"
  // "renderhive/hedera"
)


// Define a struct to represent the JSON message in the Hive Cycle Synchronization Topic
type HiveCycleConfigurationMessage struct {
    Iteration   int `json:"iteration"`       // hive cycle iteration (how many times was the condiguration changed)
    Duration    int `json:"duration"`        // hive cycle duration in seconds
    Created     time.Time `json:"timestamp"` // timestamp of the creator's clock when creating the message
}

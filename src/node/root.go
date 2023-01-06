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

package node

/*

The node package handles all the functionality of the Renderhive nodes. The node
types are:

    (1) Render nodes
    (2) Client nodes
    (3) Mediator nodes (not implemented)

*/

import (

  // standard
  // "fmt"
  // "os"
  // "time"

  // external
  hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

  // internal
  "renderhive/logger"
  // "renderhive/constants"
  "renderhive/hedera"
)

// Data required to manage the nodes
type NodeManager struct {

  User UserData
  Node NodeData

}

// User data of the node's owner
type UserData struct {

  ID int
  Username string
  UserAccount hederasdk.AccountID     // Hedera account ID of the user's main account
  NodeAccounts []hederasdk.AccountID  // Hedera account IDs of the user's node accounts

}

// Node data of the node running this service app instance
type NodeData struct {

  ID int                      // Renderhive ID of the node
  ClientNode bool             // True, if the node acts as a client node
  RenderNode bool             // True, if the node acts as a render node

  UserData *UserData
  NodeAccount *hedera.HederaAccount

}


// NODE MANAGER
// #############################################################################
// Initialize everything required for the node management
func InitNodeManager() (NodeManager, error) {
    var err error

    // log information
    logger.RenderhiveLogger.Package["node"].Info().Msg("Initializing node:")

    // create a new node manager
    nm := NodeManager{}

    // return the initialized node manager
    return nm, err

}

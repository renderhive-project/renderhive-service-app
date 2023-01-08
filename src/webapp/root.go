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

package webapp

/*

The webapp package provides the backend and front end for the user UI, which will
be served locally as a webapp.

*/

import (

  // standard
  // "fmt"
  // "os"
  // "time"

  // external
  // hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

  // internal
  "renderhive/logger"
  // "renderhive/constants"
  // "renderhive/hedera"
)

// structure for the web app manager
type WebAppManager struct {

  Placeholder string

}


// WEBAPP MANAGER
// #############################################################################
// Initialize everything required for the web app management
func InitWebAppManager() (WebAppManager, error) {
    var err error

    // log information
    logger.RenderhiveLogger.Package["webapp"].Info().Msg("Initializing web app manager:")

    // create a new IPFS manager
    webm := WebAppManager{}

    return webm, err

}

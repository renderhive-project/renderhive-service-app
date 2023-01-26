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

package utility

/*

This package contains utility/helper functions for the service app.

*/

import (

  // standard
  // "fmt"
  // "os"
  // "time"
  // "sync"

  // external
  // hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

  // internal
  // . "renderhive/globals"
  // "renderhive/logger"
  // "renderhive/node"
  // "renderhive/hedera"
  // "renderhive/ipfs"
  // "renderhive/webapp"
  // "renderhive/cli"
)


// FUNCTIONS
// #############################################################################
// Initialize the Renderhive Service App session
func InStringSlice(slice []string, test string) (bool) {

    // loop through all elements and check if one element is the 'test' string
    for _, str := range slice {
      if str == test {
          return true
      }
    }

    return false

}
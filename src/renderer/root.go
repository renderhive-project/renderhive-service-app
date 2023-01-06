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

package renderer

/*

The renderer package handles all the functionality for rendering. It manages
render jobs, the data, calls Blender, etc.

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

// structure for the render manager
type RenderManager struct {

  Placeholder string

}


// RENDER MANAGER
// #############################################################################
// Initialize everything required for the render management
func InitRenderManager() (RenderManager, error) {
    var err error

    // log information
    logger.RenderhiveLogger.Package["renderer"].Info().Msg("Initializing render manager:")

    // create a new render manager
    rm := RenderManager{}

    return rm, err

}

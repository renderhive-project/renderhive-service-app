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

package constants

import (

  // external
  "github.com/rs/zerolog"

)

// COMPILER FLAGS
// #############################################################################
const COMPILER_RENDERHIVE_LOGGER_LEVEL = zerolog.DebugLevel

// HEDERA CONSTANTS
//
const HEDERA_TESTNET_MIRROR_NODE_URL = "https://testnet.mirrornode.hedera.com:443"

// RENDERHIVE CONSTANTS
// #############################################################################
// Account ID of the Renderhive smart contract
const RENDERHIVE_TESTNET_SMART_CONTRACT = "0.0.49230883"

// Account IDs of the Hedera Consensus Service topics
// Hive cycle
const RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_SYNCHRONIZATION = "0.0.49139787"
const RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_APPLICATION = "0.0.49139788"
const RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_VALIDATION = "0.0.49139789"
// Render jobs
const RENDERHIVE_TESTNET_RENDER_JOB_QUEUE = "0.0.49139793"

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

package main

/*

This file contains all functions and other declarations for the service app.

*/

import (

  // standard
  "fmt"
  // "os"
  "time"

  // external
  // hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

  // internal
  . "renderhive/constants"
  "renderhive/logger"
  //"renderhive/cli"
  "renderhive/node"
  "renderhive/hedera"
  "renderhive/ipfs"
  "renderhive/renderer"
  "renderhive/webapp"
)



// STRUCTURES
// #############################################################################
// Data required to manage the nodes
type ServiceApp struct {

  // Managers
  NodeManager node.NodeManager
  HederaManager hedera.HederaManager
  IPFSManager ipfs.IPFSManager
  RenderManager renderer.RenderManager
  WebAppManager webapp.WebAppManager

  // Hedera consensus service topics
  // Hive cycle topics
  HiveCycleSynchronizationTopic hedera.HederaTopic
  HiveCycleApplicationTopic hedera.HederaTopic
  HiveCycleValidationTopic hedera.HederaTopic

  // Render job topics
  JobQueueTopic hedera.HederaTopic
  JobTopics []hedera.HederaTopic

}


// FUNCTIONS
// #############################################################################
// Initialize the Renderhive Service App session
func (service *ServiceApp) Init() (error) {
    var err error
    var topic *hedera.HederaTopic

    // INITIALIZE INTERNAL MANAGERS
    // *************************************************************************
    // initialize the node manager
    service.NodeManager, err = node.InitNodeManager()
    if err != nil {
      return err
    }

    // initialize the Hedera manager
    service.HederaManager, err = hedera.InitHederaManager(hedera.NETWORK_TYPE_TESTNET, "hedera/testnet.env")
    if err != nil {
      return err
    }
    logger.RenderhiveLogger.Main.Info().Msg("Loaded the account details from the environment file.")
    logger.RenderhiveLogger.Main.Info().Msg(fmt.Sprintf(" [#] Public key: %s", service.HederaManager.Operator.PublicKey))

    // initialize the IPFS manager
    service.IPFSManager, err = ipfs.InitIPFSManager()
    if err != nil {
      return err
    }

    // initialize the render manager
    service.RenderManager, err = renderer.InitRenderManager()
    if err != nil {
      return err
    }

    // initialize the web app manager
    service.WebAppManager, err = webapp.InitWebAppManager()
    if err != nil {
      return err
    }

    // READ HCS TOPIC INFORMATION & SUBSCRIBE
    // *************************************************************************
    // render job queue
    topic, err = service.HederaManager.TopicInfoFromString(RENDERHIVE_TESTNET_RENDER_JOB_QUEUE)
    if err != nil {
      return err
    }
    err = service.HederaManager.TopicSubscribe(topic, time.Unix(0, 0))
    if err != nil {
      return err
    }

    // hive cycle synchronization topic
    topic, err = service.HederaManager.TopicInfoFromString(RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_SYNCHRONIZATION)
    if err != nil {
      return err
    }
    err = service.HederaManager.TopicSubscribe(topic, time.Unix(0, 0))
    if err != nil {
      return err
    }

    // hive cycle application topic
    topic, err = service.HederaManager.TopicInfoFromString(RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_APPLICATION)
    if err != nil {
      return err
    }
    err = service.HederaManager.TopicSubscribe(topic, time.Unix(0, 0))
    if err != nil {
      return err
    }

    // hive cycle validation topic
    topic, err = service.HederaManager.TopicInfoFromString(RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_VALIDATION)
    if err != nil {
      return err
    }
    err = service.HederaManager.TopicSubscribe(topic, time.Unix(0, 0))
    if err != nil {
      return err
    }


    // STATE CHECKS
    // *************************************************************************
    // perform important state checks
    // ...

    return nil
}

/*
 * ************************** BEGIN LICENSE BLOCK ******************************
 *
 * Copyright © 2024 Christian Stolze
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
	"time"

	// "os"
	"sync"

	// external

	// internal
	"renderhive/cli"
	. "renderhive/globals"
	"renderhive/hedera"
	"renderhive/ipfs"
	"renderhive/jsonrpc"
	"renderhive/logger"
	"renderhive/node"
)

// Data required to manage the nodes
type AppManager struct {

	// Managers
	LoggerManager  *logger.PackageManager
	NodeManager    *node.PackageManager
	HederaManager  *hedera.PackageManager
	IPFSManager    *ipfs.PackageManager
	JsonRpcManager *jsonrpc.PackageManager
	CLIManager     *cli.PackageManager

	// Signaling channels
	Quit chan bool
	WG   sync.WaitGroup
}

// FUNCTIONS
// #############################################################################
// Initialize the Renderhive Service App session
func (service *AppManager) Init() error {
	var err error

	// INITIALIZE LOGGER
	// *************************************************************************
	// initialize the logger manager
	service.LoggerManager = &logger.Manager
	err = service.LoggerManager.Init()
	if err != nil {
		return err
	}

	// INITIALIZE INTERNAL MANAGERS
	// *************************************************************************
	// log the start of the renderhive service
	logger.Manager.Main.Info().Msg("Starting Renderhive service app.")

	// log debug event
	logger.Manager.Package["logger"].Debug().Msg("Initialized the logger manager.")
	logger.Manager.Package["logger"].Debug().Msg(fmt.Sprintf(" [#] The log file is located at '%s'", logger.Manager.FileWriter.Name()))

	// initialize the Hedera manager
	service.HederaManager = &hedera.Manager
	err = service.HederaManager.Init(hedera.NETWORK_TYPE_TESTNET)
	if err != nil {
		return err
	}

	// initialize the IPFS manager
	service.IPFSManager = &ipfs.Manager
	err = service.IPFSManager.Init()
	if err != nil {
		return err
	}

	// initialize the node manager
	service.NodeManager = &node.Manager
	err = service.NodeManager.Init()
	if err != nil {
		return err
	}

	// initialize the JSON-RPC manager
	service.JsonRpcManager = &jsonrpc.Manager
	err = service.JsonRpcManager.Init()
	if err != nil {
		return err
	}

	// initialize the command line interfae manager
	service.CLIManager = &cli.Manager
	err = service.CLIManager.Init()
	if err != nil {
		return err
	}

	// HIVE CYCLE
	// *************************************************************************
	if RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_SYNCHRONIZATION != "" {

		// add call to wait group
		service.WG.Add(1)

		go func() {

			// loop forever
			for {

				select {

				// app is quitting
				case <-service.Quit:
					logger.Manager.Main.Debug().Msg("Stopped hive cycle synchronization loop.")
					service.WG.Done()
					return

				// app is running
				default:

					// synchronize the hive cycle
					err := service.NodeManager.HiveCycle.Synchronize(service.HederaManager)
					if err != nil {
						logger.Manager.Main.Error().Msg(fmt.Sprintf("Error in hive cycle synchronization: %v", err))
					}

					// wait for 100 milliseconds to next check
					time.Sleep(100 * time.Millisecond)

				}
			}
		}()

	}

	// STATE CHECKS
	// *************************************************************************
	// perform important state checks
	// ...

	// LOG BASIC APP INFORMATION
	// *************************************************************************

	// log some informations about the used constants
	logger.Manager.Main.Info().Msg("This service app instance relies on the following smart contract(s) and HCS topic(s):")
	// the renderhive smart contract this instance calls
	logger.Manager.Main.Info().Msg(fmt.Sprintf(" [#] Smart Contract: %s", RENDERHIVE_TESTNET_SMART_CONTRACT))
	// Hive cycle
	logger.Manager.Main.Info().Msg(fmt.Sprintf(" [#] Hive Cycle Synchronization Topic: %s", RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_SYNCHRONIZATION))
	logger.Manager.Main.Info().Msg(fmt.Sprintf(" [#] Hive Cycle Application Topic: %s", RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_APPLICATION))
	logger.Manager.Main.Info().Msg(fmt.Sprintf(" [#] Hive Cycle Validation Topic: %s", RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_VALIDATION))
	// Render jobs
	logger.Manager.Main.Info().Msg(fmt.Sprintf(" [#] Render Job Topic: %s", RENDERHIVE_TESTNET_RENDER_JOB_QUEUE))

	return nil

}

// Deinitialize the Renderhive Service App session
func (service *AppManager) DeInit() error {
	var err error

	// log event
	logger.Manager.Main.Info().Msg("Stopping Renderhive service app ... ")

	// send the Quit signal to all concurrent go functions
	service.Quit <- true

	// log event
	logger.Manager.Main.Info().Msg("Waiting for background operations to terminate ... ")
	service.WG.Wait()

	// DEINITIALIZE INTERNAL MANAGERS
	// *************************************************************************

	// deinitialize the commmand line interface manager
	err = service.CLIManager.DeInit()
	if err != nil {
		return err
	}

	// deinitialize the JSON-RPC manager
	err = service.JsonRpcManager.DeInit()
	if err != nil {
		return err
	}

	// deinitialize the IPFS manager
	service.IPFSManager.DeInit()
	if err != nil {
		return err
	}

	// deinitialize the Hedera manager
	err = service.HederaManager.DeInit()
	if err != nil {
		return err
	}

	// deinitialize the node manager
	err = service.NodeManager.DeInit()
	if err != nil {
		return err
	}

	// LOG BASIC APP INFORMATION
	// *************************************************************************

	logger.Manager.Main.Info().Msg("Renderhive service app stopped.")

	return err

}

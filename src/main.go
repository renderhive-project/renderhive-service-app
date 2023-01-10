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

import (

  // standard
  "fmt"
  "os"
  "time"
  "sync"

  // external
  // hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

  // internal
  //. "renderhive/constants"
  "renderhive/logger"
  "renderhive/cli"
  //"renderhive/hedera"
  //"renderhive/node"
)


// APP
// #############################################################################
// INITIALIZE APP
func init() {

  // LOGGER SYSTEM
  // ***************************************************************************
  // initialize the logger system
  logger.Init()

  // add the package loggers
  logger.AddPackageLogger("node")
  logger.AddPackageLogger("hedera")
  logger.AddPackageLogger("ipfs")
  logger.AddPackageLogger("renderer")
  logger.AddPackageLogger("webapp")

}


// MAIN LOOP
func main () {

  // prepare end of program
  defer os.Exit(0)

  // error value
  var err error

  // placeholder
  fmt.Println(time.Now().Add(30 * time.Second))



  // INITIALIZE SERVICE APP
  // ***************************************************************************
  ServiceApp := ServiceApp{}

  // TODO: use the signal library to catch interrupts, so that the app still
  //       shuts down decently?
  ServiceApp.Quit = make(chan bool, 1)
  ServiceApp.WG = sync.WaitGroup{}

  // initialize service app
  err = ServiceApp.Init()
  if err != nil {
    logger.RenderhiveLogger.Main.Error().Err(err).Msg("")
    os.Exit(1)
  }

  // deinitialize the service app at the end of the main function
  defer ServiceApp.DeInit()



  // COMMAND LINE INTERFACE
  // ***************************************************************************
  // start the command line interface tool
  cli_tool := cli.CommandLine{}
  cli_tool.Start()


  // MAIN LOOP
  // ***************************************************************************
  time.Sleep(91 * time.Second)


}

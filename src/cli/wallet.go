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

package cli

import (
  "fmt"
  "flag"
)

// define a data structure for this subcommand
type WalletSubCommand struct {
	fs *flag.FlagSet

  // arguments
  idArg *bool

  // internal structure of the wallet wallet data
	id int
}

// initialize this subcommand
func NewSubCommandWallet() *WalletSubCommand {

  // create a new flag set for this subcommand
	subcmd := &WalletSubCommand{
		fs: flag.NewFlagSet("wallet", flag.ContinueOnError),
	}

  // define the arguments of this subcommand
	subcmd.idArg = subcmd.fs.Bool("id", false, "returns the wallet identifier of the local IPFS wallet")

	return subcmd
}

// Methods of the subcommand
// #############################################################################
// Get the name of the command
func (wallet *WalletSubCommand) Name() string {
	return wallet.fs.Name()
}

// Initialize the subcommand
func (wallet *WalletSubCommand) Init(args []string) error {
	return wallet.fs.Parse(args)
}

// Run the subcommand
func (wallet *WalletSubCommand) Run() error {

  if *wallet.idArg {
    fmt.Println("[INFO] ID:", wallet.id)
  }
	return nil
}

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

package jsonrpc

import "net/http"

/*

 A ping service to test if the backend <-> fronend connection is working

*/

// internal
// "renderhive/hedera"

// export the PingService
type PingService struct{}

// SERVICE DEFINITION
// #############################################################################

// Arguments and reply
type Args struct {
	Who string
}
type Reply struct {
	Message string
}

// Just say hellow to the name provided by the client
func (ps *PingService) SayHello(r *http.Request, args *Args, reply *Reply) error {

	// lock the mutex
	Manager.Mutex.Lock()
	defer Manager.Mutex.Unlock()

	// return the string
	reply.Message = "Hello, " + args.Who + "!"
	return nil

}

// INTERNAL HELPER FUNCTIONS
// #############################################################################

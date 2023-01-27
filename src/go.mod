// ************************** BEGIN LICENSE BLOCK ******************************
// *
// * Copyright © 2023 Christian Stolze
// *
// * Licensed under the Apache License, Version 2.0 (the "License");
// * you may not use this file except in compliance with the License.
// * You may obtain a copy of the License at
// *
// * http://www.apache.org/licenses/LICENSE-2.0
// *
// * Unless required by applicable law or agreed to in writing, software
// * distributed under the License is distributed on an "AS IS" BASIS,
// * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// * See the License for the specific language governing permissions and
// * limitations under the License.
// *
// ************************** END LICENSE BLOCK ********************************
// */

module renderhive

go 1.16

// The Hedera Go SDK states:
// google.golang.org/protobuf v1.27.1 Breaks the SDK as it contains multiple
// protobuf files with the same name. Make sure to use v1.26.1 instead.
// The follow snippet can be used in go.mod to force the project to use v1.26.1
replace google.golang.org/protobuf v1.27.1 => google.golang.org/protobuf v1.26.1-0.20210525005349-febffdd88e85

require github.com/hashgraph/hedera-sdk-go/v2 v2.20.0

require github.com/rs/zerolog v1.28.0

require (
	github.com/ethereum/common v0.0.0-20150727083859-e5cdbecceb9d // indirect
	github.com/ethereum/go-ethereum v1.10.26 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	// this is only required for the testnet
	github.com/joho/godotenv v1.4.0
	github.com/mattn/go-shellwords v1.0.12 // indirect
	github.com/spf13/cobra v1.6.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/exp v0.0.0-20230118134722-a68e582fa157 // indirect
)

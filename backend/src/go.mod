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

// Hedera
// google.golang.org/protobuf v1.27.1 Breaks the SDK as it contains multiple
// protobuf files with the same name.
replace google.golang.org/protobuf v1.27.1 => google.golang.org/protobuf v1.26.1-0.20210525005349-febffdd88e85

require (
	github.com/ethereum/go-ethereum v1.10.26
	github.com/hashgraph/hedera-sdk-go/v2 v2.20.0
)

// IPFS
require github.com/ipfs/kubo v0.18.1

// General
require (
	github.com/gorilla/rpc v1.2.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/ipfs/go-cid v0.3.2
	github.com/ipfs/go-libipfs v0.2.0
	github.com/ipfs/interface-go-ipfs-core v0.8.2
	github.com/jbenet/goprocess v0.1.4
	github.com/joho/godotenv v1.4.0
	github.com/klauspost/cpuid/v2 v2.2.3 // indirect
	github.com/libp2p/go-libp2p v0.24.2
	github.com/mattn/go-shellwords v1.0.12
	github.com/multiformats/go-multiaddr v0.8.0
	github.com/polydawn/refmt v0.89.0 // indirect
	github.com/rs/zerolog v1.28.0
	github.com/spf13/cobra v1.6.1
	github.com/spf13/pflag v1.0.5
	github.com/whyrusleeping/cbor-gen v0.0.0-20230126041949-52956bd4c9aa // indirect
	go4.org v0.0.0-20201209231011-d4a079459e60 // indirect
	golang.org/x/crypto v0.5.0 // indirect
	golang.org/x/exp v0.0.0-20230118134722-a68e582fa157 // indirect
)

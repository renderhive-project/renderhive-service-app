// ************************** BEGIN LICENSE BLOCK ******************************
// *
// * Copyright © 2022 Christian Stolze
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

module rendera

go 1.16

// The Hedera Go SDK states:
// google.golang.org/protobuf v1.27.1 Breaks the SDK as it contains multiple
// protobuf files with the same name. Make sure to use v1.26.1 instead.
// The follow snippet can be used in go.mod to force the project to use v1.26.1
replace google.golang.org/protobuf v1.27.1 => google.golang.org/protobuf v1.26.1-0.20210525005349-febffdd88e85

require github.com/hashgraph/hedera-sdk-go/v2 v2.17.7

require (
	github.com/btcsuite/btcd/btcec/v2 v2.2.1 // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0 // indirect
	github.com/ethereum/go-ethereum v1.10.25 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/hashgraph/hedera-protobufs-go v0.2.1-0.20220831114249-138cd7171d62 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rs/zerolog v1.28.0 // indirect
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	golang.org/x/crypto v0.0.0-20220919173607-35f4265a4bc0 // indirect
	golang.org/x/net v0.0.0-20220919232410-f2f64ebce3c1 // indirect
	golang.org/x/sys v0.0.0-20220919091848-fb04ddd9f9c8 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220919141832-68c03719ef51 // indirect
	google.golang.org/grpc v1.49.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

// this is only required for the testnet
require github.com/joho/godotenv v1.4.0

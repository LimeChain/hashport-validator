/*
 * Copyright 2022 LimeChain Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package constants

import (
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
)

var (
	Networks = map[int64]*parser.Network{
		0: {
			Tokens: map[string]parser.Token{
				constants.Hbar: {
					Networks: map[int64]string{
						33: "0x0000000000000000000000000000000000000001",
					},
				},
			},
		},
		1: {
			Tokens: map[string]parser.Token{
				"0xsomeethaddress": {
					Networks: map[int64]string{
						33: "0x0000000000000000000000000000000000000123",
					},
				},
			},
		},
		2: {
			Tokens: map[string]parser.Token{
				"0x0000000000000000000000000000000000000000": {
					Networks: map[int64]string{
						0: "",
					},
				},
			},
		},
		3: {
			Tokens: map[string]parser.Token{
				"0x0000000000000000000000000000000000000000": {
					Networks: map[int64]string{
						0: "",
					},
				},
			},
		},
		32: {
			Tokens: map[string]parser.Token{
				"0x0000000000000000000000000000000000000000": {
					Networks: map[int64]string{
						0: "",
					},
				},
			},
		},
		33: {
			Tokens: map[string]parser.Token{
				"0x0000000000000000000000000000000000000000": {
					Networks: map[int64]string{
						0: constants.Hbar,
						1: "0xsome-other-eth-address",
					},
				},
			}},
	}
)

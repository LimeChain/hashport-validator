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

package verify

import (
	"encoding/json"
	"testing"

	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/setup"
)

func EventTransactionIDFromValidatorAPI(t *testing.T, setupEnv *setup.Setup, eventID, expectedTxID string) {
	t.Helper()
	actualTxID, err := setupEnv.Clients.ValidatorClient.GetEventTransactionID(eventID)
	if err != nil {
		t.Fatalf("[%s] - Failed to get event transaction ID. Error: [%s]", eventID, err)
	}

	if actualTxID != expectedTxID {
		t.Fatalf("Expected Event TX ID [%s] did not match actual TX ID [%s]", expectedTxID, actualTxID)
	}
}

func FungibleTransferFromValidatorAPI(t *testing.T, setupEnv *setup.Setup, evm setup.EVMUtils, txId, tokenID, expectedSendAmount, targetAsset string) *service.FungibleTransferData {
	t.Helper()
	bytes, err := setupEnv.Clients.ValidatorClient.GetTransferData(txId)
	if err != nil {
		t.Fatalf("Cannot fetch transaction data - Error: [%s].", err)
	}
	var transferDataResponse *service.FungibleTransferData
	err = json.Unmarshal(bytes, &transferDataResponse)
	if err != nil {
		t.Fatalf("Failed to parse JSON transaction data [%s]. Error: [%s]", bytes, err)
	}

	if transferDataResponse.IsNft {
		t.Fatalf("Transaction data mismatch: Expected response data to not be NFT related.")
	}
	if transferDataResponse.Amount != expectedSendAmount {
		t.Fatalf("Transaction data mismatch: Expected [%s], but was [%s]", expectedSendAmount, transferDataResponse.Amount)
	}
	if transferDataResponse.NativeAsset != tokenID {
		t.Fatalf("Native Token mismatch: Expected [%s], but was [%s]", setupEnv.TokenID.String(), transferDataResponse.NativeAsset)
	}
	if transferDataResponse.Recipient != evm.Receiver.String() {
		t.Fatalf("Receiver address mismatch: Expected [%s], but was [%s]", evm.Receiver.String(), transferDataResponse.Recipient)
	}
	if transferDataResponse.TargetAsset != targetAsset {
		t.Fatalf("Token address mismatch: Expected [%s], but was [%s]", targetAsset, transferDataResponse.TargetAsset)
	}

	return transferDataResponse
}

func NonFungibleTransferFromValidatorAPI(t *testing.T, setupEnv *setup.Setup, evm setup.EVMUtils, txId string, tokenID string, metadata string, tokenIdOrSerialNum int64, targetAsset string) *service.NonFungibleTransferData {
	t.Helper()
	bytes, err := setupEnv.Clients.ValidatorClient.GetTransferData(txId)
	if err != nil {
		t.Fatalf("Cannot fetch transaction data - Error: [%s].", err)
	}
	var transferDataResponse *service.NonFungibleTransferData
	err = json.Unmarshal(bytes, &transferDataResponse)
	if err != nil {
		t.Fatalf("Failed to parse JSON transaction data [%s]. Error: [%s]", bytes, err)
	}

	if !transferDataResponse.IsNft {
		t.Fatalf("Transaction data mismatch: Expected response data to be NFT related.")
	}
	if transferDataResponse.Metadata != metadata {
		t.Fatalf("Transaction data mismatch: Expected [%s], but was [%s]", metadata, transferDataResponse.Metadata)
	}
	if transferDataResponse.TokenId != tokenIdOrSerialNum {
		t.Fatalf("Transaction tokenId/serialNum mismatch: Expected [%d], but was [%d]", tokenIdOrSerialNum, transferDataResponse.TokenId)
	}
	if transferDataResponse.NativeAsset != tokenID {
		t.Fatalf("Native Token mismatch: Expected [%s], but was [%s]", setupEnv.TokenID.String(), transferDataResponse.NativeAsset)
	}
	if transferDataResponse.Recipient != evm.Receiver.String() {
		t.Fatalf("Receiver address mismatch: Expected [%s], but was [%s]", evm.Receiver.String(), transferDataResponse.Recipient)
	}
	if transferDataResponse.TargetAsset != targetAsset {
		t.Fatalf("Token address mismatch: Expected [%s], but was [%s]", targetAsset, transferDataResponse.TargetAsset)
	}

	return transferDataResponse
}

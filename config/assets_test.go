package config

import (
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

var (
	networks = map[int64]*parser.Network{
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

func Test_LoadAssets(t *testing.T) {
	assets := LoadAssets(networks)
	if reflect.TypeOf(assets).String() != "config.Assets" {
		t.Fatalf(`Expected to return assets type *config.Assets, but returned: [%s]`, reflect.TypeOf(assets).String())
	}
}

func Test_IsNative(t *testing.T) {
	assets := LoadAssets(networks)

	actual := assets.IsNative(0, constants.Hbar)
	assert.Equal(t, true, actual)

	actual = assets.IsNative(0, "0x0000000000000000000000000000000000000000")
	assert.Equal(t, false, actual)
}

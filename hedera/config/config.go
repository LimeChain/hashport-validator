package config

var AccountData = struct {
	Operator struct {
		AccountId  string
		PublicKey  string
		PrivateKey string
	}
	Network map[string]string
}{
	Operator: struct {
		AccountId  string
		PublicKey  string
		PrivateKey string
	}{
		"0.0.99661",
		"302a300506032b6570032100d8646c0c1a84c77210302879dadd7fd67ada72ae8f2c49df04e01f2e2fa27ef7",
		"302e020100300506032b657004220420fc79b49c62c4637437292814eccd640a6173933c30698ddb41c36c195f0b6629",
	},
	Network: map[string]string{
		"0.testnet.hedera.com:50211": "0.0.3",
		"1.testnet.hedera.com:50211": "0.0.4",
		"2.testnet.hedera.com:50211": "0.0.5",
		"3.testnet.hedera.com:50211": "0.0.6",
	},
}

var MirrorNodeClientAddress = "hcs.testnet.mirrornode.hedera.com:5600"
var MirrorNodeAPIAddress = "https://testnet.mirrornode.hedera.com/api/v1/"

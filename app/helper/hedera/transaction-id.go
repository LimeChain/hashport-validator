package hedera

import (
	"fmt"
	"strings"
)

// ToMirrorNodeTransactionID parses TX with format `0.0.X@{seconds}.{nanos}` to format `0.0.X-{seconds}-{nanos}`
func ToMirrorNodeTransactionID(txId string) string {
	split := strings.Split(txId, "?")
	split = strings.Split(split[0], "@")
	accId := split[0]
	split = strings.Split(split[1], ".")
	return fmt.Sprintf("%s-%s-%s", accId, split[0], fmt.Sprintf("%09s", split[1]))
}

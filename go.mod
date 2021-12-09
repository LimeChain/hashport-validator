module github.com/limechain/hedera-eth-bridge-validator

go 1.13

require (
	github.com/caarlos0/env/v6 v6.4.0
	github.com/ethereum/go-ethereum v1.10.8
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-chi/render v1.0.1
	github.com/golang/protobuf v1.5.2
	github.com/hashgraph/hedera-sdk-go/v2 v2.6.0
	github.com/pkg/errors v0.9.1
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20210907225631-ff17edfbf26d
	google.golang.org/protobuf v1.26.1-0.20210525005349-febffdd88e85
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/driver/postgres v1.0.5
	gorm.io/gorm v1.20.6
)

replace github.com/hashgraph/hedera-sdk-go/v2 => github.com/limechain/hedera-sdk-go/v2 v2.0.0-20211209111032-975e06a0142d

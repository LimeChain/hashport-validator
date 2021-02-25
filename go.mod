module github.com/limechain/hedera-eth-bridge-validator

go 1.13

require (
	github.com/btcsuite/btcd v0.21.0-beta // indirect
	github.com/caarlos0/env/v6 v6.4.0
	github.com/ethereum/go-ethereum v1.9.24
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-chi/render v1.0.1
	github.com/golang/protobuf v1.4.3
	github.com/hashgraph/hedera-sdk-go/v2 v2.1.4
	github.com/jackc/pgx/v4 v4.9.2 // indirect
	github.com/limechain/hedera-watcher-sdk v0.0.0-20210219143218-2592d3168472
	github.com/rs/cors v0.0.0-20160617231935-a62a804a8a00
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.6.1
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/net v0.0.0-20210224082022-3d97a244fca7 // indirect
	golang.org/x/sys v0.0.0-20210223212115-eede4237b368 // indirect
	google.golang.org/genproto v0.0.0-20210223151946-22b48be4551b // indirect
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v2 v2.3.0
	gorm.io/driver/postgres v1.0.5
	gorm.io/gorm v1.20.6
)

package config_bridge

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"net/http"
)

var (
	Route = "/config/bridge"
)

//Router for bridge config
func NewRouter(bridgeConfig parser.Bridge) http.Handler {
	r := chi.NewRouter()
	r.Get("/", configBridgeResponse(bridgeConfig))
	return r
}

// GET: .../config/bridge
func configBridgeResponse(bridgeConfig parser.Bridge) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, bridgeConfig)
	}
}

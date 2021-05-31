package webhook

import (
	"fmt"
	"net/http"

	"github.com/k8s-at-home/gateway-admision-controller/internal/config"
	"github.com/k8s-at-home/gateway-admision-controller/internal/log"
)

// Config is the handler configuration.
type Config struct {
	CmdConfig config.CmdConfig
	Logger    log.Logger
}

func (c *Config) defaults() error {

	if c.Logger == nil {
		c.Logger = log.Dummy
	}

	return nil
}

type handler struct {
	handler   http.Handler
	cmdConfig config.CmdConfig
	logger    log.Logger
}

// New returns a new webhook handler.
func New(config Config) (http.Handler, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("handler configuration is not valid: %w", err)
	}

	mux := http.NewServeMux()

	h := handler{
		handler:   mux,
		cmdConfig: config.CmdConfig,
		logger:    config.Logger.WithKV(log.KV{"service": "webhook-handler"}),
	}

	// Register all the routes with our router.
	err = h.routes(mux)
	if err != nil {
		return nil, fmt.Errorf("could not register routes on handler: %w", err)
	}

	return h, nil
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}

package webhook

import (
	"net/http"
)

// routes wires the routes to handlers on a specific router.
func (h handler) routes(router *http.ServeMux) error {
	gatewayPodMutator, err := h.gatewayPodMutator()
	if err != nil {
		return err
	}
	router.Handle("/wh/mutating/gatewayPodMutator", gatewayPodMutator)

	return nil
}

package webhook

import (
	"encoding/json"
	"net/http"
)

// routes wires the routes to handlers on a specific router.
func (h handler) routes(router *http.ServeMux) error {

	// Add gatewayPodMutator
	gatewayPodMutator, err := h.gatewayPodMutator()
	if err != nil {
		return err
	}
	router.Handle("/wh/mutating/setgateway", gatewayPodMutator)

	//Add health
	router.HandleFunc("/wh/health", func(w http.ResponseWriter, r *http.Request) {
		// an example API handler
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	return nil
}

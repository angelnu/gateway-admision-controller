package webhook

import (
	"context"
	"fmt"
	"net/http"

	kwhhttp "github.com/slok/kubewebhook/v2/pkg/http"
	kwhlog "github.com/slok/kubewebhook/v2/pkg/log"
	kwhmutating "github.com/slok/kubewebhook/v2/pkg/webhook/mutating"

	"github.com/k8s-at-home/gateway-admision-controller/internal/log"
	"github.com/k8s-at-home/gateway-admision-controller/internal/mutation"
)

// kubewebhookLogger is a small proxy to use our logger with Kubewebhook.
type kubewebhookLogger struct {
	log.Logger
}

func (l kubewebhookLogger) WithValues(kv map[string]interface{}) kwhlog.Logger {
	return kubewebhookLogger{Logger: l.Logger.WithKV(kv)}
}
func (l kubewebhookLogger) WithCtxValues(ctx context.Context) kwhlog.Logger {
	return l.WithValues(kwhlog.ValuesFromCtx(ctx))
}
func (l kubewebhookLogger) SetValuesOnCtx(parent context.Context, values map[string]interface{}) context.Context {
	return kwhlog.CtxWithValues(parent, values)
}

// allmark sets up the webhook handler for marking all kubernetes resources using Kubewebhook library.
func (h handler) gatewayPodMutator() (http.Handler, error) {

	logger := kubewebhookLogger{Logger: h.logger.WithKV(log.KV{"lib": "kubewebhook", "webhook": "gatewayPodMutator"})}

	// Create our mutator
	gwPodMutator, err := gatewayPodMutator.NewGatewayPodMutator(h.cmdConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("error creating webhook mutator: %w", err)
	}
	mt := kwhmutating.MutatorFunc(gwPodMutator.GatewayPodMutator)

	wh, err := kwhmutating.NewWebhook(kwhmutating.WebhookConfig{
		ID:      "gatewayPodMutator",
		Logger:  logger,
		Mutator: mt,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create webhook: %w", err)
	}
	whHandler, err := kwhhttp.HandlerFor(kwhhttp.HandlerConfig{
		Webhook: wh,
		Logger:  logger,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create handler from webhook: %w", err)
	}

	return whHandler, nil
}

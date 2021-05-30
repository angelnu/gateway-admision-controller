package gatewayPodMutator_test

import (
	"context"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mutator "github.com/k8s-at-home/gateway-admision-controller/internal/mutation"
)

func getExpectedObj(gatewayIP string, keepDNS bool) metav1.Object {
	pod := corev1.Pod{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{
				corev1.Container{
					Name:  "addGateway",
					Image: "alpine",
					Command: append(
						strings.Split("ip route add default via", " "),
						gatewayIP,
					),
					SecurityContext: &corev1.SecurityContext{
						Privileged: &[]bool{true}[0],
					},
				},
			},
		},
	}
	if !keepDNS {

		pod.Spec.DNSPolicy = "None"
		pod.Spec.DNSConfig = &corev1.PodDNSConfig{
			Nameservers: []string{
				gatewayIP,
			},
		}
	}
	return &pod
}

func TestGatewayPodMutator(t *testing.T) {

	exampleGatewayName := "example.com"
	exampleGatewayNameIPs, _ := net.LookupIP(exampleGatewayName)

	tests := map[string]struct {
		gateway string
		keepDNS bool
		obj     metav1.Object
		expObj  metav1.Object
	}{
		"Gateway IP, keepDNS=false - Having a pod, gateway should be added": {
			gateway: "1.2.3.4",
			keepDNS: false,
			obj:     &corev1.Pod{},
			expObj:  getExpectedObj("1.2.3.4", false),
		},
		"Gateway IP, keepDNS=true - Having a pod, gateway should be added": {
			gateway: "1.2.3.4",
			keepDNS: true,
			obj:     &corev1.Pod{},
			expObj:  getExpectedObj("1.2.3.4", true),
		},
		"Gateway name, keepDNS=true - Having a pod, gateway should be added": {
			gateway: exampleGatewayName,
			keepDNS: false,
			obj:     &corev1.Pod{},
			expObj:  getExpectedObj(exampleGatewayNameIPs[0].String(), false),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			m, err := mutator.NewGatewayPodMutator(test.gateway, test.keepDNS)
			require.NoError(err)

			_, err = m.GatewayPodMutator(context.TODO(), nil, test.obj)
			require.NoError(err)

			assert.Equal(test.expObj, test.obj)
		})
	}
}

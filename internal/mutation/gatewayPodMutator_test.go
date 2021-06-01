package gatewayPodMutator_test

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/k8s-at-home/gateway-admision-controller/internal/config"
	mutator "github.com/k8s-at-home/gateway-admision-controller/internal/mutation"
)

const (
	testGatewayIP           = "1.2.3.4"
	testGatewayName         = "example.com"
	testDNSPolicy           = "None"
	testInitImage           = "initImg"
	testInitImagePullPol    = "Always"
	testInitCmd             = "initCmd"
	testInitMountPoint      = "/media"
	testSidecarImage        = "sidecarImg"
	testSidecarImagePullPol = "IfNotPresent"
	testSidecarCmd          = "sidecarCmd"
	testSidecarMountPoint   = "/mnt"
	testConfigmapName       = "settings"
)

func getExpectedPodSpec(gateway string) corev1.PodSpec {

	exampleGatewayNameIPs, _ := net.LookupIP(gateway)

	spec := corev1.PodSpec{
		InitContainers: []corev1.Container{
			corev1.Container{
				Name:    mutator.GATEWAY_INIT_CONTAINER_NAME,
				Image:   testInitImage,
				Command: []string{testInitCmd},
				Env: []corev1.EnvVar{
					{
						Name:  "gateway",
						Value: gateway,
					},
				},
				ImagePullPolicy: corev1.PullPolicy(testInitImagePullPol),
				SecurityContext: &corev1.SecurityContext{
					Privileged: &[]bool{true}[0],
				},
				VolumeMounts: []corev1.VolumeMount{
					corev1.VolumeMount{
						Name:      mutator.GATEWAY_CONFIGMAP_VOLUME_NAME,
						ReadOnly:  true,
						MountPath: testInitMountPoint,
					},
				},
			},
		},
		DNSConfig: &corev1.PodDNSConfig{
			Nameservers: []string{
				exampleGatewayNameIPs[0].String(),
			},
		},
	}

	spec.Volumes = append(spec.Volumes, corev1.Volume{
		Name: mutator.GATEWAY_CONFIGMAP_VOLUME_NAME,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: testConfigmapName,
				},
				DefaultMode: &mutator.GATEWAY_CONFIGMAP_VOLUME_MODE,
			},
		},
	})
	return spec
}

func getExpectedPodSpec_keepDNS(gateway string) corev1.PodSpec {
	spec := getExpectedPodSpec(gateway)
	spec.DNSPolicy = ""
	spec.DNSConfig = nil
	return spec
}

func getExpectedPodSpec_sidecar(gateway string) corev1.PodSpec {
	spec := getExpectedPodSpec(gateway)

	container := corev1.Container{
		Name:    mutator.GATEWAY_SIDECAR_CONTAINER_NAME,
		Image:   testSidecarImage,
		Command: []string{testSidecarCmd},
		Env: []corev1.EnvVar{
			{
				Name:  "gateway",
				Value: gateway,
			},
		},
		ImagePullPolicy: corev1.PullPolicy(testSidecarImagePullPol),
		SecurityContext: &corev1.SecurityContext{
			Privileged: &[]bool{true}[0],
		},
		VolumeMounts: []corev1.VolumeMount{
			corev1.VolumeMount{
				Name:      mutator.GATEWAY_CONFIGMAP_VOLUME_NAME,
				ReadOnly:  true,
				MountPath: testSidecarMountPoint,
			},
		},
	}

	//Add container to pod
	spec.Containers = append(spec.Containers, container)

	return spec
}

func getExpectedPodSpec_DNSPolicy(gatewayIP string) corev1.PodSpec {
	spec := getExpectedPodSpec(gatewayIP)
	spec.DNSPolicy = testDNSPolicy
	return spec
}

func TestGatewayPodMutator(t *testing.T) {

	tests := map[string]struct {
		cmdConfig config.CmdConfig
		obj       metav1.Object
		expObj    metav1.Object
	}{
		"Gateway IP - Having a pod, gateway should be added": {
			cmdConfig: config.CmdConfig{
				Gateway:           testGatewayIP,
				SetGatewayDefault: true,
				InitImage:         testInitImage,
				InitCmd:           testInitCmd,
				InitImagePullPol:  testInitImagePullPol,
				InitMountPoint:    testInitMountPoint,
				ConfigmapName:     testConfigmapName,
			},
			obj: &corev1.Pod{},
			expObj: &corev1.Pod{
				Spec: getExpectedPodSpec(testGatewayIP),
			},
		},
		"Gateway name - Having a pod, gateway should be added": {
			cmdConfig: config.CmdConfig{
				Gateway:           testGatewayName,
				SetGatewayDefault: true,
				InitImage:         testInitImage,
				InitCmd:           testInitCmd,
				InitImagePullPol:  testInitImagePullPol,
				InitMountPoint:    testInitMountPoint,
				ConfigmapName:     testConfigmapName,
			},
			obj: &corev1.Pod{},
			expObj: &corev1.Pod{
				Spec: getExpectedPodSpec(testGatewayName),
			},
		},
		"Gateway IP, keepDNS=true - Having a pod, gateway should be added": {
			cmdConfig: config.CmdConfig{
				Gateway:           testGatewayIP,
				SetGatewayDefault: true,
				InitImage:         testInitImage,
				InitCmd:           testInitCmd,
				InitImagePullPol:  testInitImagePullPol,
				InitMountPoint:    testInitMountPoint,
				ConfigmapName:     testConfigmapName,
				KeepDNS:           true,
			},
			obj: &corev1.Pod{},
			expObj: &corev1.Pod{
				Spec: getExpectedPodSpec_keepDNS(testGatewayIP),
			},
		},
		"Gateway IP, no SetGatewayDefault - it should be a NOP": {
			cmdConfig: config.CmdConfig{
				Gateway: testGatewayIP,
			},
			obj:    &corev1.Pod{},
			expObj: &corev1.Pod{},
		},
		"Gateway IP, setGatewayLabel='setGateway' - it should be a NOP": {
			cmdConfig: config.CmdConfig{
				Gateway:           testGatewayIP,
				SetGatewayDefault: true,
				InitImage:         testInitImage,
				InitCmd:           testInitCmd,
				InitImagePullPol:  testInitImagePullPol,
				InitMountPoint:    testInitMountPoint,
				ConfigmapName:     testConfigmapName,
				SetGatewayLabel:   "setGateway",
			},
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"setGateway": "false",
					},
				},
			},
			expObj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"setGateway": "false",
					},
				},
			},
		},
		"Gateway IP, setGatewayLabel='setGateway' - it should set gateway since label is true": {
			cmdConfig: config.CmdConfig{
				Gateway:           testGatewayIP,
				SetGatewayDefault: true,
				InitImage:         testInitImage,
				InitCmd:           testInitCmd,
				InitImagePullPol:  testInitImagePullPol,
				InitMountPoint:    testInitMountPoint,
				ConfigmapName:     testConfigmapName,
				SetGatewayLabel:   "setGateway",
			},
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"setGateway": "true",
					},
				},
			},
			expObj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"setGateway": "true",
					},
				},
				Spec: getExpectedPodSpec(testGatewayIP),
			},
		},
		"Gateway IP, setGatewayAnnotation='setGateway' - it should be a NOP": {
			cmdConfig: config.CmdConfig{
				Gateway:              testGatewayIP,
				SetGatewayDefault:    true,
				InitImage:            testInitImage,
				InitCmd:              testInitCmd,
				InitImagePullPol:     testInitImagePullPol,
				InitMountPoint:       testInitMountPoint,
				ConfigmapName:        testConfigmapName,
				SetGatewayAnnotation: "setGateway",
			},
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"setGateway": "false",
					},
				},
			},
			expObj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"setGateway": "false",
					},
				},
			},
		},
		"Gateway IP, setGatewayAnnotation='setGateway' - it should set gateway since label is true": {
			cmdConfig: config.CmdConfig{
				Gateway:              testGatewayIP,
				SetGatewayDefault:    true,
				InitImage:            testInitImage,
				InitCmd:              testInitCmd,
				InitImagePullPol:     testInitImagePullPol,
				InitMountPoint:       testInitMountPoint,
				ConfigmapName:        testConfigmapName,
				SetGatewayAnnotation: "setGateway",
			},
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"setGateway": "true",
					},
				},
			},
			expObj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"setGateway": "true",
					},
				},
				Spec: getExpectedPodSpec(testGatewayIP),
			},
		},
		"Gateway IP, sidecar cmd": {
			cmdConfig: config.CmdConfig{
				Gateway:             testGatewayIP,
				SetGatewayDefault:   true,
				InitImage:           testInitImage,
				InitCmd:             testInitCmd,
				InitImagePullPol:    testInitImagePullPol,
				InitMountPoint:      testInitMountPoint,
				ConfigmapName:       testConfigmapName,
				SidecarImage:        testSidecarImage,
				SidecarCmd:          testSidecarCmd,
				SidecarImagePullPol: testSidecarImagePullPol,
				SidecarMountPoint:   testSidecarMountPoint,
			},
			obj: &corev1.Pod{},
			expObj: &corev1.Pod{
				Spec: getExpectedPodSpec_sidecar(testGatewayIP),
			},
		},
		"Gateway IP, DNSPolicy": {
			cmdConfig: config.CmdConfig{
				Gateway:           testGatewayIP,
				SetGatewayDefault: true,
				InitImage:         testInitImage,
				InitCmd:           testInitCmd,
				InitImagePullPol:  testInitImagePullPol,
				InitMountPoint:    testInitMountPoint,
				ConfigmapName:     testConfigmapName,
				SetDNSPolicy:      testDNSPolicy,
			},
			obj: &corev1.Pod{},
			expObj: &corev1.Pod{
				Spec: getExpectedPodSpec_DNSPolicy(testGatewayIP),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			m, err := mutator.NewGatewayPodMutator(test.cmdConfig)
			require.NoError(err)

			_, err = m.GatewayPodMutator(context.TODO(), nil, test.obj)
			require.NoError(err)

			assert.Equal(test.expObj, test.obj)
		})
	}
}

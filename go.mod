module github.com/k8s-at-home/gateway-admision-controller

go 1.16

require (
	github.com/oklog/run v1.1.0
	github.com/sirupsen/logrus v1.8.1
	github.com/slok/kubewebhook/v2 v2.1.0
	github.com/stretchr/testify v1.7.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.21.3
	k8s.io/apimachinery v0.21.3
)

replace k8s.io/client-go/v12 => k8s.io/client-go v12.0.0

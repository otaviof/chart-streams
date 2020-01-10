module github.com/otaviof/chart-streams

go 1.13

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/gin-gonic/contrib v0.0.0-20190923054218-35076c1b2bea
	github.com/gin-gonic/gin v1.4.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	github.com/ugorji/go v1.1.7 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.2
	gopkg.in/src-d/go-git.v4 v4.13.1
	helm.sh/helm/v3 v3.0.1 // indirect
	k8s.io/apimachinery v0.0.0-20191014065749-fb3eea214746 // indirect
	k8s.io/client-go v11.0.0+incompatible // indirect
	k8s.io/helm v2.16.1+incompatible
)

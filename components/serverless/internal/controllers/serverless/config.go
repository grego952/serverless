package serverless

import (
	"time"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type FunctionConfig struct {
	PublisherProxyAddress                      string         `envconfig:"optional"`
	TraceCollectorEndpoint                     string         `envconfig:"optional"`
	ImageRegistryDefaultDockerConfigSecretName string         `envconfig:"default=serverless-registry-config-default"`
	PackageRegistryConfigSecretName            string         `envconfig:"default=serverless-package-registry-config"`
	ImagePullAccountName                       string         `envconfig:"default=serverless-function"`
	TargetCPUUtilizationPercentage             int32          `envconfig:"default=50"`
	RequeueDuration                            time.Duration  `envconfig:"default=1m"`
	FunctionReadyRequeueDuration               time.Duration  `envconfig:"default=5m"`
	GitFetchRequeueDuration                    time.Duration  `envconfig:"default=30s"`
	ResourceConfig                             ResourceConfig `envconfig:"optional"`
	Build                                      BuildConfig
}

type BuildConfig struct {
	ExecutorArgs        []string `envconfig:"default=--insecure;--skip-tls-verify;--skip-unused-stages;--log-format=text;--cache=true;--use-new-run;--compressed-caching=false"`
	ExecutorImage       string   `envconfig:"default=gcr.io/kaniko-project/executor:v1.9.2"`
	RepoFetcherImage    string   `envconfig:"default=europe-docker.pkg.dev/kyma-project/prod/function-build-init:v20230426-37b02524"`
	MaxSimultaneousJobs int      `envconfig:"default=5"`
}

type DockerConfig struct {
	ActiveRegistryConfigSecretName string
	PushAddress                    string
	PullAddress                    string
}

type ResourceConfig struct {
	Function FunctionResourceConfig `yaml:"function"`
	BuildJob BuildJobResourceConfig `yaml:"buildJob"`
}

var _ envconfig.Unmarshaler = &ResourceConfig{}

func (rc *ResourceConfig) Unmarshal(input string) error {
	err := yaml.Unmarshal([]byte(input), rc)
	return err
}

type FunctionResourceConfig struct {
	Resources Resources `yaml:"resources"`
}

type BuildJobResourceConfig struct {
	Resources Resources `yaml:"resources"`
}

type Resources struct {
	Presets            Preset         `yaml:"presets"`
	RuntimePresets     RuntimePresets `yaml:"runtimePresets"`
	DefaultPreset      string         `yaml:"defaultPreset"`
	MinRequestedCPU    Quantity       `yaml:"minRequestedCPU"`
	MinRequestedMemory Quantity       `yaml:"minRequestedMemory"`
}

type RuntimePresets map[string]Preset

type Preset map[string]Resource

type Resource struct {
	RequestCPU    Quantity `yaml:"requestCpu"`
	RequestMemory Quantity `yaml:"requestMemory"`
	LimitCPU      Quantity `yaml:"limitCpu"`
	LimitMemory   Quantity `yaml:"limitMemory"`
}

func (p Preset) ToResourceRequirements() map[string]corev1.ResourceRequirements {
	resources := map[string]corev1.ResourceRequirements{}
	for k, v := range p {
		resources[k] = corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    v.LimitCPU.Quantity,
				corev1.ResourceMemory: v.LimitMemory.Quantity,
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    v.RequestCPU.Quantity,
				corev1.ResourceMemory: v.RequestMemory.Quantity,
			},
		}
	}
	return resources
}

type Quantity struct {
	resource.Quantity
}

func (q *Quantity) UnmarshalYAML(unmarshal func(interface{}) error) error {
	quantity := ""
	err := unmarshal(&quantity)
	if err != nil {
		return errors.Wrap(err, "while unmarshalling quantity")
	}
	out, err := resource.ParseQuantity(quantity)
	if err != nil {
		return errors.Wrap(err, "while parsing quantity")
	}
	q.Quantity = out
	return nil
}

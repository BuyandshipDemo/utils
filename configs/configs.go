package configs

import (
	"fmt"
	"os"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	cc "github.com/kitex-contrib/config-etcd/etcd"
	etcdServer "github.com/kitex-contrib/config-etcd/server"
	prometheus "github.com/kitex-contrib/monitor-prometheus"
	etcd "github.com/kitex-contrib/registry-etcd"
	"gopkg.in/yaml.v3"
)

// KitexInfoConfig
type KitexInfoConfig struct {
	ServiceName string `yaml:"ServiceName"`
	ToolVersion string `yaml:"ToolVersion"`
}

// DiscoveryConfig
type DiscoveryConfig struct {
	Component string `yaml:"component"`
	Address   string `yaml:"address"`
}

// ConfigCenterConfig
type ConfigCenterConfig struct {
	Component string `yaml:"component"`
	Address   string `yaml:"address"`
}

// LoggingConfig
type LoggingConfig struct {
	Path  string `yaml:"path"`
	Level string `yaml:"level"`
}

// MonitoringConfig
type MonitoringConfig struct {
	Component string `yaml:"component"`
	Address   string `yaml:"address"`
}

// ObservabilityConfig
type ObservabilityConfig struct {
	Logging    LoggingConfig    `yaml:"logging"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
}

// GovernanceConfig
type GovernanceConfig struct {
	Discovery    DiscoveryConfig    `yaml:"discovery"`
	ConfigCenter ConfigCenterConfig `yaml:"config_center"`
}

// AppConfig
type AppConfig struct {
	KitexInfo     KitexInfoConfig     `yaml:"kitexinfo"`
	Governance    GovernanceConfig    `yaml:"governance"`
	Observability ObservabilityConfig `yaml:"observability"`
}

func loadConfig(fpath string) *AppConfig {
	c, err := os.ReadFile(fpath)
	if err != nil {
		panic(fmt.Sprintf("Failed to open YAML file: %v", err))
	}

	config := &AppConfig{}
	err = yaml.Unmarshal(c, config)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal YAML: %v", err))
	}
	return config
}

// Load
func Load(fpath string) []server.Option {
	cfg := loadConfig(fpath)

	var opts []server.Option

	opts = append(opts, server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
		ServiceName: cfg.KitexInfo.ServiceName,
	}))
	// service registry
	if cfg.Governance.Discovery.Address != "" {
		r, err := etcd.NewEtcdRegistry([]string{cfg.Governance.Discovery.Address})
		if err != nil {
			panic(err)
		}
		opts = append(opts, server.WithRegistry(r))
	}

	// config center
	if cfg.Governance.ConfigCenter.Address != "" {
		configCli, err := cc.NewClient(cc.Options{
			Node: []string{cfg.Governance.ConfigCenter.Address},
		})
		if err != nil {
			panic(err)
		}
		opts = append(opts, server.WithSuite(etcdServer.NewSuite(cfg.KitexInfo.ServiceName, configCli)))
	}

	// monitoring
	if cfg.Observability.Monitoring.Address != "" {
		opts = append(opts, server.WithTracer(prometheus.NewServerTracer(cfg.Observability.Monitoring.Address, "/kitexserver")))
	}
	return opts
}

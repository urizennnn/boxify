package config

// ContainerConfig represents the configuration for launching a container
type ContainerConfig struct {
	ImageName      string         `yaml:"image_name" json:"image_name"`
	ResourceLimits ResourceLimits `yaml:"resource_limits" json:"resource_limits"`
}

// ResourceLimits defines cgroup resource constraints for containers
type ResourceLimits struct {
	MemoryLimit string `yaml:"memory_limit" json:"memory_limit"`
	CpuLimit    string `yaml:"cpu_limit" json:"cpu_limit"`
}

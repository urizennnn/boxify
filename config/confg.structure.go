package config

type ConfigStructure struct {
	ImageName  string   `yaml:"image_name" json:"image_name"`
	Settings Settings `yaml:"settings" json:"settings"`
}

type Settings struct {
	MemoryLimit string `yaml:"memory_limit" json:"memory_limit"`
	CpuLimit    string `yaml:"cpu_limit" json:"cpu_limit"`
}

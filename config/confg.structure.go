package config

type ConfigStructure struct {
	ImageName  string   `json:"image_name"`
	Settings Settings `json:"settings"`
}

type Settings struct {
	MemoryLimit string `json:"memory_limit"`
	CpuLimit    string `json:"cpu_limit"`
}

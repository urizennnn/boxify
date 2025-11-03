package requests

type InitContainerRequest struct {
	Name         string `json:"name"`
	OriginFolder string `json:"origin_folder"`
	MemoryLimit  string `json:"memory_limit"`
	CpuLimit     string `json:"cpu_limit"`
}

package container

type ContainerStructure struct {
	Id       string                      `json:"Id"`
	Name     string                      `json:"Name"`
	Image    string                      `json:"Image"`
	Status   string                      `json:"Status"`
	State    string                      `json:"State"`
	Networks map[string]NetworkStructure `json:"Networks"`
}

type NetworkStructure struct {
	IpAddress string `json:"IPAddress"`
	Gateway   string `json:"Gateway"`
	Bridge    string `json:"Bridge"`
	Cidr      string `json:"Cidr"`
}

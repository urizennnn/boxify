package container

// ContainerListItem represents a container's information for listing/display purposes
type ContainerListItem struct {
	Id       string                           `json:"Id"`
	Name     string                           `json:"Name"`
	Image    string                           `json:"Image"`
	Status   string                           `json:"Status"`
	State    string                           `json:"State"`
	Networks map[string]ContainerNetworkInfo `json:"Networks"`
}

// ContainerNetworkInfo represents network information for display purposes
type ContainerNetworkInfo struct {
	IpAddress string `json:"IPAddress"`
	Gateway   string `json:"Gateway"`
	Bridge    string `json:"Bridge"`
	Cidr      string `json:"Cidr"`
}

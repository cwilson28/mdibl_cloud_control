package datamodels

type EC2InstanceDetails struct {
	Name          string `json:"name"`
	InstanceID    string `json:"instance_id"`
	InstanceType  string `json:"instance_type"`
	InstanceState string `json:"instance_state"`
	PrivateIP     string `json:"private_ip"`
	PublicIP      string `json:"public_ip"`
}

type EC2InstanceReport struct {
	Instances []EC2InstanceDetails `json:"instances"`
}

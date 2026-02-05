package domain

type HealthReport struct {
	Live    bool   `json:"live"`
	Ready   bool   `json:"ready"`
	DB      bool   `json:"db"`
	Version string `json:"version,omitempty"`
	Time    string `json:"time"`
}

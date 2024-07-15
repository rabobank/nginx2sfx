package model

// VcapApplication - structure for VCAP_APPLICATION environment variable
type VcapApplication struct {
	ApplicationID      string   `json:"application_id"`
	ApplicationName    string   `json:"application_name"`
	ApplicationUris    []string `json:"application_uris"`
	ApplicationVersion string   `json:"application_version"`
	CfAPI              string   `json:"cf_api"`
	Limits             struct {
		Disk int64 `json:"disk"`
		Fds  int64 `json:"fds"`
		Mem  int64 `json:"mem"`
	} `json:"limits"`
	Name             string   `json:"name"`
	OrganizationID   string   `json:"organization_id"`
	OrganizationName string   `json:"organization_name"`
	ProcessID        string   `json:"process_id"`
	ProcessType      string   `json:"process_type"`
	SpaceID          string   `json:"space_id"`
	SpaceName        string   `json:"space_name"`
	Uris             []string `json:"uris"`
	Version          string   `json:"version"`
}

// VcapServices - structure for VCAP_SERVICES environment variable
type VcapServices struct {
	Credhub []VcapService `json:"credhub"`
}

type VcapService struct {
	Credentials struct {
		Token string `json:"token,omitempty"`
	} `json:"credentials"`
	InstanceName string `json:"instance_name"`
}

// LogLine - structure for the logline from the nginx access log
type LogLine struct {
	Uri            string  `json:"uri"`
	Method         string  `json:"method"`
	ServerProtocol string  `json:"server_protocol"`
	Request        string  `json:"request"`
	Status         string  `json:"status"`
	BodyBytesSent  float64 `json:"body_bytes_sent"`
	RequestTime    float64 `json:"request_time"`
}

type CounterMetrics struct {
	Counter []CounterMetric `json:"counter"`
}

type CounterMetric struct {
	Metric     string     `json:"metric"`
	Value      float64    `json:"value"`
	Dimensions Dimensions `json:"dimensions"`
}

type Dimensions struct {
	Uri             string `json:"uri,omitempty"`
	Method          string `json:"method"`
	ServerProtocol  string `json:"server_protocol"`
	StatusCode      string `json:"status_code"`
	Cfenv           string `json:"cfenv"`
	CfInstanceIndex string `json:"cf_instance_index"`
	CfAppName       string `json:"cf_app_name"`
	CfAppId         string `json:"cf_app_id"`
	CfSpaceName     string `json:"cf_space_name"`
	CfOrgName       string `json:"cf_org_name"`
}

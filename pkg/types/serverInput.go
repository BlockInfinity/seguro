package types

type ScanPostReq struct {
	ProjectName       string
	ProjectRemoteUrls []string
	Revision          string
	Findings          []UnifiedFinding
}

type DevicePostReq struct {
	DeviceName string
}

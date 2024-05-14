package types

type ScanPostReq struct {
	ProjectName       string
	ProjectRemoteUrls []string
	Revision          string
	Findings          []UnifiedFinding
}

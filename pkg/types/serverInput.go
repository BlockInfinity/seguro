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

type DependencycheckScanPostReq struct {
	ManifestFiles []FileReq
}

type FileReq struct {
	Path    string
	Content []byte
}

type FixedFileContentPostReq struct {
	FileContent       string
	ProblemLineNumber int
	Hint              string
}

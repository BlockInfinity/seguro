package reporting

import "github.com/secguro/secguro-cli/pkg/types"

type ScanPostReq struct {
	ProjectName       string
	ProjectRemoteUrls []string
	Revision          string
	Findings          []types.UnifiedFinding
}

type ConfirmationRes struct {
	Status string
}

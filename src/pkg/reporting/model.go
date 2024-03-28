package reporting

import "secguro.com/secguro/pkg/types"

type ScanPostReq struct {
	ProjectName string
	Revision    string
	Findings    []types.UnifiedFinding
}

type ConfirmationRes struct {
	Status string
}

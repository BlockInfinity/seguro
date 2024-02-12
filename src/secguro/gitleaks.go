package main

type GitleaksFinding struct {
	File      string
	StartLine int
}

func convertGitleaksFindingToUnifiedFinding(gitleaksFinding GitleaksFinding) UnifiedFinding {
	return UnifiedFinding{
		File: gitleaksFinding.File,
		Line: gitleaksFinding.StartLine,
	}
}

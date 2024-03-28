package scan

import "secguro.com/secguro/pkg/functional"

func IsSecretDetectionRule(rule string) bool {
	secretDetectionRules := []string{
		"generic-api-key",
		"generic.secrets.security.detected-generic-api-key.detected-generic-api-key",
	}

	return functional.ArrayIncludes(secretDetectionRules, rule)
}

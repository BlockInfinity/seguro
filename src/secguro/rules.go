package main

func isSecretDetectionRule(rule string) bool {
	secretDetectionRules := []string{
		"generic-api-key",
		"generic.secrets.security.detected-generic-api-key.detected-generic-api-key",
	}

	return arrayIncludes(secretDetectionRules, rule)
}

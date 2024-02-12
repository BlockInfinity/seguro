## Installation
- Download: `wget 'https://secguro.github.io/secguro-cli/secguro'`
- Make executable: `chmod +x secguro`
- Move to directory in `$PATH`

## Use
### Locally
- `secguro scan`

### Github Workflow
```yaml
    - name: Check for Secguro Violations
      run: wget 'https://secguro.github.io/secguro-cli/secguro' && chmod +x secguro &&  ./secguro scan
```

### Azure Pipeline
```yaml
    - task: CmdLine@2
      displayName: Check for Secguro Violations
      inputs:
        script: wget 'https://secguro.github.io/secguro-cli/secguro' && chmod +x secguro &&  ./secguro scan
        workingDirectory: .
        failOnStderr: false # because wget writes to stderr
```

## Deveploment
- Go version: 1.21.7

### Linter
- Installation: `yay -S golangci-lint` or https://golangci-lint.run/usage/install/#local-installation
- Manual invocation: `make lint`
- Activation of pre-push hook: `git config core.hooksPath hooks`

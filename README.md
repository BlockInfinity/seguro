## Installation
- Download: `wget 'https://secguro.github.io/secguro-cli/secguro'`
- Make executable: `chmod +x secguro`
- Move to directory in `$PATH`; e.g. `mv secguro ~/.local/bin`

## Use
### Locally
```bash
secguro scan
```

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

## Exit Code
Exit codes ranging from 0 to 250 (inclusive) indicate the number of findings. Exit code 250 indicates 250 or more findings. Ignored findings are not counted.

Exit codes not equal to 0 are useful to make Github Workflows and Azure Pipelines fail.

Switch `--tolerance n` (or `--tolerance=n`) may be used to make secguro yield exit code 0 if the number of findigs does not exceed `n`.

## Deveploment
- Go version: 1.21.7

### Linter
- Installation: `yay -S golangci-lint` or https://golangci-lint.run/usage/install/#local-installation
- Manual invocation: `make lint`
- Activation of pre-push hook: `git config core.hooksPath hooks`

### Compilation
```bash
make
```

Location of generated binary: `build/secguro`

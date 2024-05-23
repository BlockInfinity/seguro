## Prerequisites
### Operating System and Architecture
Either:
- GNU/Linux AMD64
- Darwin ARM64

### Dependencies for Use
- Python 3 + pipx
- Java 8 or above; If Java 8, update version must be 251 or above
- Git (only necessary when using git mode or detector gitleaks)

### Dependencies for Development
- [Dependencies for use](#dependencies-for-use)
- Go version 1.21.7

## Installation
- Download: `wget 'https://secguro.github.io/secguro-cli/secguro'`
- Make executable: `chmod +x secguro`
- Move to directory in `$PATH`; e.g. `mv secguro ~/.local/bin`

## Scanning for Problems
### Locally
```bash
secguro scan [path]
```

### Github Workflow
```yaml
    - name: Check for Secguro Violations
      run: wget 'https://secguro.github.io/secguro-cli/secguro' && chmod +x secguro && SECGURO_CI_TOKEN="GET THIS TOKEN FROM THE SECGURO WEBAPP" ./secguro scan
```

### Azure Pipeline
```yaml
    - task: CmdLine@2
      displayName: Check for Secguro Violations
      inputs:
        script: wget 'https://secguro.github.io/secguro-cli/secguro' && chmod +x secguro && SECGURO_CI_TOKEN="GET THIS TOKEN FROM THE SECGURO WEBAPP" ./secguro scan
        workingDirectory: .
        failOnStderr: false # because wget writes to stderr
```

## Fixing Problems
```bash
secguro fix [path]
```

## Exit Code
Exit codes ranging from 0 to 250 (inclusive) indicate the number of findings. Exit code 250 indicates 250 or more findings. Ignored findings are not counted.

Exit codes not equal to 0 are useful to make Github Workflows and Azure Pipelines fail.

Switch `--tolerance n` (or `--tolerance=n`) may be used to make secguro yield exit code 0 if the number of findigs does not exceed `n`.

## Options
```
$ secguro scan --help
NAME:
   secguro scan - scan for problems

USAGE:
   secguro scan [command options] [arguments...]

OPTIONS:
   --git                                                      set to scan git history and print commit information (default: false)
   --disabled-detectors value [ --disabled-detectors value ]  list of detectors to disable (semgrep,gitleaks,dependencycheck)
   --format value                                             text or json (default: "text")
   --output value, -o value                                   path to output destination
   --tolerance value                                          number of findings to tolerate when choosing exit code (default: 0)
   --help, -h                                                 show help
```

```
$ secguro fix --help
NAME:
   secguro fix - scan for problems and then switch to an interactive mode to fix them

USAGE:
   secguro fix [command options] [arguments...]

OPTIONS:
   --git                                                      set to scan git history and print commit information (default: false)
   --disabled-detectors value [ --disabled-detectors value ]  list of detectors to disable (semgrep,gitleaks,dependencycheck)
   --help, -h                                                 show help
```

### Development
### Linter
- Installation: `yay -S golangci-lint` or https://golangci-lint.run/usage/install/#local-installation
- Manual invocation: `make lint`
- Activation of pre-push hook: `git config core.hooksPath hooks`

### Compilation
To generate a binary that communicates with the CD server, run:
```bash
make
```

For a developer build that communicates with localhost, run:
```bash
make compile-dev
```

Location of the generated binary: `build/secguro`

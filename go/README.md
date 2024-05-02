# compliance-to-policy
Compliance-to-Policy (C2P) provides the framework to bridge Compliance administration and Policy administration by [OSCAL](https://pages.nist.gov/OSCAL/). OSCAL (Open Security Controls Assessment Language) is a standardized framework developed by NIST for expressing and automating the assessment and management of security controls in machine-readable format (xml, json, yaml)

![C2P Overview](/go/docs/images/e2e-pm.png)

## Usage of C2P CLI
```
$ c2pcli -h        
C2P CLI

Usage:
  c2pcli [flags]
  c2pcli [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  kyverno     C2P CLI Kyverno plugin
  ocm         C2P CLI OCM plugin
  version     Display version

Flags:
  -h, --help   help for c2pcli

Use "c2pcli [command] --help" for more information about a command.
```

C2P is targeting a plugin architecture to cover not only OCM Policy Framework but also other types of PVPs. 
Please go to the docs for each usage.
- [C2P for OCM](/go/docs/ocm/README.md) 
- [C2P for Kyverno](/go/docs/kyverno/README.md) 

## Build at local
```
make build
```
```
./bin/c2pcli_<version>_<os>_<arch> -h
```

## Test
```
make test
```

## Release
1. Create a git tag of the following format `go/<version>` (e.g. `go/v0.1.2`)
1. Run release command
    ```
    echo $PAT | gh auth login --with-token -h github.com
    make release 
    ```
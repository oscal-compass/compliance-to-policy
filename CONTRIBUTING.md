## Contributing In General

Our project welcomes external contributions. If you have an itch, please feel
free to scratch it.

To contribute code or documentation, please submit a [pull request](https://github.com/oscal-compass/compliance-to-policy/pulls).

A good way to familiarize yourself with the codebase and contribution process is
to look for and tackle low-hanging fruit in the [issue tracker](https://github.com/oscal-compass/compliance-to-policy/issues).
Before embarking on a more ambitious contribution, please quickly [get in touch](/MAINTAINERS.md) with us.

**Note: We appreciate your effort, and want to avoid a situation where a contribution
requires extensive rework (by you or by us), sits in backlog for a long time, or
cannot be accepted at all!**

We have also adopted [Contributor Covenant Code of Conduct](/CODE_OF_CONDUCT.md).

### Proposing new features

If you would like to implement a new feature, please [raise an issue](https://github.com/oscal-compass/compliance-to-policy/issues)
labelled `enhancement` before sending a pull request so the feature can be discussed. This is to avoid
you wasting your valuable time working on a feature that the project developers
are not interested in accepting into the code base.

### Fixing bugs

If you would like to fix a bug, please [raise an issue](https://github.com/oscal-compass/compliance-to-policy/issues) labelled `bug` before sending a
pull request so it can be tracked.

### Merge approval

The project maintainers use LGTM (Looks Good To Me) in comments on the code
review to indicate acceptance. A change requires LGTMs from one of the maintainers.

For a list of the maintainers, see the [maintainers](/MAINTAINERS.md) page.

### C2P merging and release workflow

`C2P` is operating on a simple, yet opinionated, method for continuous integration. It's designed to give developers a coherent understanding of the objectives of other past developers.
The criteria for this are below. Trestle effectively uses a gitflow workflow with one modification: PR's merge into develop are squash merged as one commit.

In trestle's CI environment this results in the following rules:

1. All Commit's *MUST* be signed off with `git commit --signoff` irrespective of the author's affiliation. This ensures all code can be attributed.
   1. This is enforced by DCO bot and can be overrided by maintainers presuming at least one commit is signed-off.
1. All commits *SHOULD* use [conventional commits](https://www.conventionalcommits.org/en/v1.0.0-beta.2/)
   1. This is as github, when only one commit is in a PR, will use the native git commit message as the merge commit title.
      1. When only a single commit is provided the commit MUST be an conventional commit and will be checked the `Lint PR` aciton.
1. All PR's title's MUST be formed as an [convention commit](https://www.conventionalcommits.org/en/v1.0.0-beta.2/)
   1. This is checked by the `Lint PR` action
1. All PR's to `main` should close at least one issue by [linking the PR to an issue](https://docs.github.com/en/issues/tracking-your-work-with-issues/linking-a-pull-request-to-an-issue#linking-a-pull-request-to-an-issue-using-a-keyword).
1. C2P will release on demand.
1. Each feature/fix/chore (PR into develop) be represented by a single commit into develop / main with a coherent title (in the PR).
   1. The C2P preference for doing this is to use squash merge functionality when merging a PR into develop.
1. Developers *MUST* pass the required CI checks for each PR.
1. Developers are encouraged to use GitHub's automated merge process where possible to keep the number of active PR's low.

## Typing, docstrings and documentation

`C2P` has a goal of using [PEP 484](https://www.python.org/dev/peps/pep-0484/) type annotations where possible / practical.
The devops process does not _strictly_ enforce typing, however, the expectation is that type coverage is added for new
commits with a focus on quality over quantity (e.g. don't add `Any` everywhere just to meet coverage requirements).
Python typing of functions is an active work in progress.

## Legal

Each source file must include a license header for the Apache
Software License 2.0. Using the SPDX format is the simplest approach.
e.g.

```text
# Copyright (c) 2020 IBM Corp. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
```

We have tried to make it as easy as possible to make contributions. This
applies to how we handle the legal aspects of contribution. We use the
same approach - the [Developer's Certificate of Origin 1.1 (DCO)](https://oscal-compass.github.io/compliance-trestle/contributing/DCO/) - that the Linux® Kernel [community](https://elinux.org/Developer_Certificate_Of_Origin)
uses to manage code contributions.

We simply ask that when submitting a patch for review, the developer
must include a sign-off statement in the commit message.

Here is an example Signed-off-by line, which indicates that the
submitter accepts the DCO:

```text
Signed-off-by: John Doe <john.doe@example.com>
```

You can include this automatically when you commit a change to your
local git repository using the following command:

```bash
git commit --signoff
```

Note that DCO signoff is enforced by [DCO bot](https://github.com/probot/dco). Missing DCO's will be required to be rebased
with a signed off commit before being accepted.

## Setup - Developing `C2P`

### Does `C2P` run correctly on my platform

- Setup a venv for python in .venv directory in the repository root directory.
- Run `make install-dev`
  - This will install all python dependencies.
  - It will also checkout the submodules required for testing.
- Run `make test`
  - This *should* run on all platforms.

### Setting up `vscode` for python.

- Use the following commands to setup python:

```bash
python3 -m venv venv
. ./venv/bin/activate
make install-dev
```

- Install vscode plugin [Python extension for Visual Studio Code](https://marketplace.visualstudio.com/items?itemName=ms-python.python)

- Install vscode plugin [Formatter extension for Visual Studio Code using the Black formatter](https://marketplace.visualstudio.com/items?itemName=ms-python.black-formatter)

- Configure vscode setting with the black-formatter enabled. The example setting.json is as follows:
      ```
      {
        "[python]": {
          "diffEditor.ignoreTrimWhitespace": false,
          "gitlens.codeLens.symbolScopes": [
            "!Module"
          ],
          "editor.formatOnType": true,
          "editor.formatOnSave": true,
          "editor.wordBasedSuggestions": "off",
          "editor.defaultFormatter": "ms-python.black-formatter",
          "editor.tabSize": 4,
        },
        "isort.args":["--profile", "black"],
        "black-formatter.args": [
          "--line-length=120",
          "--skip-string-normalization"
        ]
      }
      ```
### Testing python

Tests should be in the test subdirectory. Each file should be named test\_\*.py and each test function should be named test\_\*().
Tests can be executed by `make test`.

If you want to debug test, here is the example launch.json.
```
{
  "version": "0.2.0",
  "configurations": [{
      "name": "Pytest current file",
      "type": "debugpy",
      "request": "launch",
      "module": "pytest",
      "console": "integratedTerminal",
      "args": ["${file}"],
      "justMyCode": false
   }]
}
```

### Code style and formating

`C2P` uses [Black](https://black.readthedocs.io/en/stable/) for code formatting and [isort](https://pycqa.github.io/isort/) for sorting imports. `make format` runs both tools.

`C2P` also uses [pre-commit](https://pre-commit.com/) hooks that are integrated into the development process with [detect-secrets](https://github.com/IBM/detect-secrets) to prevent from contaminating any confidential data.  

## For Go project
Please refer to [go/README.md](/go)
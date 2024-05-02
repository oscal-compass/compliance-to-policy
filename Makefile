PYTHON := $(shell pwd)/.venv/bin/python

.venv:
	@echo Please create venv firstly

build: .venv
	@$(PYTHON) -m build

install: .venv
	@$(PYTHON) -m pip install .

install-dev: .venv
	@$(PYTHON) -m pip install ".[dev]"

uninstall: .venv
	@$(PYTHON) -m pip uninstall compliance-to-policy


format: .venv
	@$(PYTHON) -m isort .
	@$(PYTHON) -m black .

lint: .venv
	@$(PYTHON) -m pylint ./c2p ./tests

.PHONY: docs
docs: .venv
	@$(PYTHON) -m mkdocs build

.PHONY: gh-pages
 gh-pages: .venv
	@$(PYTHON) -m mkdocs gh-deploy

# make test ARGS="-n 2 --dist loadscope --log-cli-level DEBUG" TARGET="tests/c2p/test_cli.py"
# TODO: -n 2 (pytest-xdist plugin) results in no logs displayed.
test: ARGS ?= 
test: TARGET ?= tests/
test: .venv test-plugin
	@OUTPUT_PATH=/dev/null $(PYTHON) -m pytest $(ARGS) $(TARGET)

test-plugin: ARGS ?= 
test-plugin: TARGET ?= plugins_public/tests/
test-plugin: .venv
	@OUTPUT_PATH=/dev/null $(PYTHON) -m pytest $(ARGS) $(TARGET)

# After published, the branch must be merged first-forwardly. TODO: Integrate with CI
publish: GIT_TAG ?=
publish:
	@toml set --toml-path pyproject.toml project.version $(GIT_TAG)
	@git add pyproject.toml
	@git commit -S -s -m "update version to $(GIT_TAG)"
	@git tag $(GIT_TAG)

clean: .venv
	@rm -rf build *.egg-info dist
	@find ./plugins -type d \( -name '*.egg-info' -o -name 'dist' \) | while read x; do echo $$x; rm -r $$x ; done 
	@$(PYTHON) -m pyclean -v .
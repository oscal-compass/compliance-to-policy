.PHONY: build
build:
	python -m build

.PHONY: install
install:
	python -m pip install .

.PHONY: install-dev
install-dev:
	python -m pip install ".[dev]"

# Direct dependency is not allowed for Pypi packaging even if the dependant module is defined as extra dependencies. 
# Workaround: Move to manual installation by make
.PHONY: install-detect-descret
install-detect-descret:
	python -m pip install detect-secrets@git+https://github.com/ibm/detect-secrets.git@master#egg=detect-secrets

.PHONY: uninstall
uninstall:
	python -m pip uninstall compliance-to-policy

.PHONY: format
format:
	python -m isort .
	python -m black .

.PHONY: lint
lint:
	python -m pylint ./c2p ./tests

.PHONY: docs
docs:
	python -m mkdocs build

.PHONY: gh-pages
 gh-pages:
	python -m mkdocs gh-deploy

# make test ARGS="-n 2 --dist loadscope --log-cli-level DEBUG" TARGET="tests/c2p/test_cli.py"
# TODO: -n 2 (pytest-xdist plugin) results in no logs displayed.
.PHONY: test
test: ARGS ?= 
test: TARGET ?= tests/
test: test-plugin
	@OUTPUT_PATH=/dev/null python -m pytest $(ARGS) $(TARGET)

.PHONY: test-plugin
test-plugin: ARGS ?= 
test-plugin: TARGET ?= plugins_public/tests/
test-plugin:
	@OUTPUT_PATH=/dev/null python -m pytest $(ARGS) $(TARGET)

.PHONY: it
it:
	python samples_public/kyverno/compliance_to_policy.py
	python samples_public/kyverno/result_to_compliance.py
	python samples_public/ocm/compliance_to_policy.py
	python samples_public/ocm/result_to_compliance.py
	python samples_public/auditree/compliance_to_policy.py
	python samples_public/auditree/result_to_compliance.py

.PHONY: clean
clean:
	@rm -rf build *.egg-info dist
	@find ./plugins -type d \( -name '*.egg-info' -o -name 'dist' \) | while read x; do echo $$x; rm -r $$x ; done 
	python -m pyclean -v .
.PHONY: test test-unit test-integration test-e2e test-py build

build:
	cd server && go build ./...

test: test-unit test-integration test-e2e test-py

test-unit:
	cd server && go test $$(go list ./... | grep -vE '/integration|/e2e')

test-integration:
	cd server && go test -tags=integration ./integration/...

test-e2e:
	cd server && go test -tags=e2e ./e2e/...

test-py:
	cd worker && python -m pytest tests/ -q

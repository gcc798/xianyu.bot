.PHONY: build build-server build-web

build: build-web build-server

build-web:
	cd agent-web && npm ci --prefer-offline --no-audit --no-fund && npm run build

build-server:
	cd server && go build ./cmd/server

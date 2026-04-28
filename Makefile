.PHONY: dev
dev:
	$(MAKE) -j 2 dev-server dev-client

.PHONY: dev-server
dev-server:
	cd web-server && WALLET_NOTE_TABLE=wallet-note-development go run ./cmd/wallet-note

.PHONY: dev-client
dev-client:
	cd web-client && pnpm dev

.PHONY: lint
lint:
	cd web-server && go vet ./...

.PHONY: tf-plan-dev
tf-plan-dev:
	cd infrastructure/environments/development && terraform plan

.PHONY: tf-apply-dev
tf-apply-dev:
	cd infrastructure/environments/development && terraform apply

.PHONY: deploy-server-dev
deploy-server-dev:
	./scripts/deploy-server-dev.sh

.PHONY: deploy-client-dev
deploy-client-dev:
	./scripts/deploy-client-dev.sh

.PHONY: test
test:
	docker compose up -d dynamodb-local
	cd web-server && go test ./...

.PHONY: test-down
test-down:
	docker compose down

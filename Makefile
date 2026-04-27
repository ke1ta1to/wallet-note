.PHONY: dev-server
dev-server:
	cd web-server && go run ./cmd/wallet-note

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

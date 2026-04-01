.DEFAULT_GOAL := help
SHELL := /bin/bash
APP = keycloak-broker
COMMIT_SHA = $(shell git rev-parse --short HEAD)

.PHONY: help
## help: prints this help message
help:
	@echo "Usage:"
	@sed -n 's/^## //p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: dev
## dev: runs main.go with the golang race detector
dev:
	source _fixtures/env; source .env_private; go run -race main.go

.PHONY: run
## run: runs via air hot-reloader
run:
	source _fixtures/env; source .env_private; air

.PHONY: build
## build: builds the application
build: clean
	@echo "Building binary ..."
	@mise trust --all || true
	go build -o ${APP}

.PHONY: clean
## clean: cleans up binary files
clean:
	@echo "Cleaning up ..."
	@mise trust --all || true
	@go clean

.PHONY: test
## test: runs go test with the race detector
test:
	@mise trust --all || true
	GOARCH=amd64 GOOS=linux go test -v -race ./...

.PHONY: init
## init: sets up go modules
init:
	@echo "Setting up modules ..."
	@go mod init 2>/dev/null; go mod tidy && go mod vendor

.PHONY: install-air
## install-air: installs air hot-reloader
install-air:
	go install github.com/air-verse/air@v1.64.5
	#go install github.com/air-verse/air@latest

.PHONY: cleanup
cleanup: docker-cleanup
.PHONY: docker-cleanup
## docker-cleanup: cleans up local docker images and volumes
docker-cleanup:
	docker system prune --volumes -a

#=======================================================================================================================
.PHONY: setup
## setup: setup keycloak for development
setup: start-keycloak keycloak-client-create

.PHONY: start-keycloak
## start-keycloak: start dev keycloak
start-keycloak:
	@echo "Starting keycloak ..."
	./scripts/start_keycloak.sh

.PHONY: delete-keycloak
## delete-keycloak: delete keycloak dev server
delete-keycloak: keycloak-client-delete
	docker kill keycloak-dev || true
	docker rm keycloak-dev || true

.PHONY: keycloak-client-create
## keycloak-client-create: create OIDC client directly
keycloak-client-create:
	./scripts/create_client.sh

.PHONY: keycloak-client-get
## keycloak-client-get: query OIDC client directly
keycloak-client-get:
	./scripts/get_client.sh

.PHONY: keycloak-client-delete
## keycloak-client-delete: delete OIDC client directly
keycloak-client-delete:
	./scripts/delete_client.sh

.PHONY: cleanup
cleanup: docker-cleanup
.PHONY: docker-cleanup
## docker-cleanup: cleans up local docker images and volumes
docker-cleanup:
	docker system prune --volumes -a

#=======================================================================================================================
.PHONY: metrics-check
## metrics-check: check broker metrics endpoint
metrics-check:
	curl -v http://localhost:9999/metrics

.PHONY: health-check
## health-check: check broker health endpoint
health-check:
	curl -v http://localhost:9999/health

.PHONY: provision
## provision: creates an example service instance
provision:
	curl -v http://disco:dingo@localhost:9999/v2/service_instances/fe5556b9-8478-409b-ab2b-3c95ba06c5fc \
		-X PUT -H "Content-Type: application/json" \
		-d '{ "service_id":"fff5b36a-da19-4dc2-bd28-3dd331146290", "plan_id":"40627d0f-dedd-4d68-8111-2ebae510ba1b", "parameters": { "redirect_uris": ["https://myapp.example.com/callback"] } }' \
		| jq .

.PHONY: fetch-instance
## fetch-instance: queries example service instance
fetch-instance:
	curl -v http://disco:dingo@localhost:9999/v2/service_instances/fe5556b9-8478-409b-ab2b-3c95ba06c5fc \
		-X GET | jq .

.PHONY: deprovision
## deprovision: deletes example service instance
deprovision:
	curl -v http://disco:dingo@localhost:9999/v2/service_instances/fe5556b9-8478-409b-ab2b-3c95ba06c5fc \
		-X DELETE

.PHONY: bind-instance
## bind-instance: creates an example service instance binding
bind-instance:
	curl -v http://disco:dingo@localhost:9999/v2/service_instances/fe5556b9-8478-409b-ab2b-3c95ba06c5fc/service_bindings/db59931a-70a6-43c1-8885-b0c6b1c194d4 \
		-X PUT -H "Content-Type: application/json" | jq .

.PHONY: fetch-binding
## fetch-binding: queries example service instance binding
fetch-binding:
	curl -v http://disco:dingo@localhost:9999/v2/service_instances/fe5556b9-8478-409b-ab2b-3c95ba06c5fc/service_bindings/db59931a-70a6-43c1-8885-b0c6b1c194d4 \
		-X GET | jq .

.PHONY: delete-binding
## delete-binding: deletes example service instance binding
delete-binding:
	curl -v http://disco:dingo@localhost:9999/v2/service_instances/fe5556b9-8478-409b-ab2b-3c95ba06c5fc/service_bindings/db59931a-70a6-43c1-8885-b0c6b1c194d4  \
		-X DELETE

#=======================================================================================================================
#================== registry related stuff
.PHONY: image-login
## image-login: login to registry
image-login:
	@export PATH="$$HOME/bin:$$PATH"
	@echo $$CONTAINER_REGISTRY_PASSWORD | docker login -u $$CONTAINER_REGISTRY_USERNAME --password-stdin $$CONTAINER_REGISTRY_HOSTNAME

.PHONY: image-build
## image-build: build registry image
image-build: build
	@export PATH="$$HOME/bin:$$PATH"
	docker build -t $$CONTAINER_REGISTRY_HOSTNAME/swisscom/${APP}:${COMMIT_SHA} .

.PHONY: image-publish
## image-publish: build and publish registry image
image-publish:
	@export PATH="$$HOME/bin:$$PATH"
	docker push $$CONTAINER_REGISTRY_HOSTNAME/swisscom/${APP}:${COMMIT_SHA}
	docker tag $$CONTAINER_REGISTRY_HOSTNAME/swisscom/${APP}:${COMMIT_SHA} $$CONTAINER_REGISTRY_HOSTNAME/swisscom/${APP}:$$HELM_CHART_VERSION
	docker tag $$CONTAINER_REGISTRY_HOSTNAME/swisscom/${APP}:${COMMIT_SHA} $$CONTAINER_REGISTRY_HOSTNAME/swisscom/${APP}:latest
	docker push $$CONTAINER_REGISTRY_HOSTNAME/swisscom/${APP}:$$HELM_CHART_VERSION
	docker push $$CONTAINER_REGISTRY_HOSTNAME/swisscom/${APP}:latest

.PHONY: image-run
## image-run: run registry image
image-run:
	@export PATH="$$HOME/bin:$$PATH"
	docker run --rm -p 8080:8080 $$CONTAINER_REGISTRY_HOSTNAME/swisscom/${APP}:${COMMIT_SHA}

.PHONY: image-clean
## image-clean: cleans up local registry image
image-clean:
	docker rm -f ${APP} || true
	docker rm -f $$CONTAINER_REGISTRY_HOSTNAME/swisscom/${APP} || true

.PHONY: image-test
## image-test: runs tests for image
image-test: test image-build
	docker images | grep swisscom/${APP}

########################################################################################################################
####### helm related stuff, needed for concourse #######################################################################
########################################################################################################################
.PHONY: helm-test
## helm-test: runs testing
helm-test: test image-test helm-render
	# TODO: implement helm chart testsuite, see https://helm.sh/docs/topics/chart_tests/
	# helm test keycloak-broker

.PHONY: helm-update-chart
## helm-update-chart: update image(s) and versions in helm chart
helm-update-chart:
	sed -i "s|repository:.*|repository: $$CONTAINER_REGISTRY_HOSTNAME/swisscom/keycloak-broker|g" kubernetes/values.yaml
	sed -i "s|registry:.*|registry: $$CONTAINER_REGISTRY_HOSTNAME|g" kubernetes/values.yaml
	sed -i "s|tag:.*|tag: $$HELM_CHART_VERSION|g" kubernetes/values.yaml
	sed -i "s|version:.*|version: $$HELM_CHART_VERSION|g" kubernetes/Chart.yaml
	sed -i "s|appVersion:.*|appVersion: $$HELM_CHART_VERSION|g" kubernetes/Chart.yaml

.PHONY: helm-render
## helm-render: renders the helm chart templates
helm-render:
	helm template keycloak-broker "./kubernetes" -n kube-system --atomic --timeout 10m -f "kubernetes/values.yaml"

.PHONY: helm-deploy
## helm-deploy: deploys the helm chart
helm-deploy:
	helm upgrade --install --cleanup-on-fail --atomic --wait --timeout "10m" \
		-f "values.yaml" -n kube-system keycloak-broker "./kubernetes"

.PHONY: helm-package
## helm-package: packages the helm chart
helm-package: helm-update-chart
	helm package "./kubernetes" --version $$HELM_CHART_VERSION

.PHONY: helm-images
## helm-images: list images used by helm chart
helm-images:
	@helm template "./kubernetes" | yq '..|.image? | select(.)' | sort -u | grep -v '\---'

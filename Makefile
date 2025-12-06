include .env
export

.PHONY: help
help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


compose-up: ### Run docker-compose
	docker compose up --build -d && docker compose logs -f
.PHONY: compose-up

compose-down: ### Down docker-compose
	docker compose down --remove-orphans
.PHONY: compose-down

docker-rm-volume: ### Remove docker volume
	docker volume rm pg-data
.PHONY: docker-rm-volume


migrate-create: ### Create new migration
	migrate create -ext sql -dir migrations 'support_bot'
.PHONY: migrate-create

# migrate-up: ### Migration up
# 	migrate -path migrations -database '$(PG_URL_LOCALHOST)?sslmode=disable' up
# .PHONY: migrate-up

# migrate-down: ### Migration down
# 	echo "y" | migrate -path migrations -database '$(PG_URL_LOCALHOST)?sslmode=disable' down
# .PHONY: migrate-down

test: ### Run tests
	go test -v ./...

cover-html: ### run test with coverage and open html report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out
.PHONY: coverage-html

cover: ### run test with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	rm coverage.out
.PHONY: coverage



generate-deploy-key: ### generate rsa keys for deploy
	ssh-keygen -q -t rsa -N '' -f !deploy/deploy_rsa
.PHONY: generate-deploy-key

build:
	docker --log-level=info build --pull --file=./Dockerfile --tag=${REGISTRY}/feedback-bot:${IMAGE_TAG} .
.PHONY: build

deploy:
	ssh -o StrictHostKeyChecking=no -i ./!deploy/deploy_rsa deploy@${HOST} -p ${PORT} 'docker network create --driver=overlay traefik-public || true'
	ssh -o StrictHostKeyChecking=no -i ./!deploy/deploy_rsa deploy@${HOST} -p ${PORT} 'rm -rf support_bot_${BUILD_NUMBER} && mkdir support_bot_${BUILD_NUMBER}'

	envsubst < docker-compose-production.yaml > docker-compose-production-env.yml
	scp -o StrictHostKeyChecking=no -i ./!deploy/deploy_rsa -P ${PORT} docker-compose-production-env.yml deploy@${HOST}:support_bot_${BUILD_NUMBER}/docker-compose.yml
	scp -o StrictHostKeyChecking=no -i ./!deploy/deploy_rsa -P ${PORT} ./.env.prod deploy@${HOST}:support_bot_${BUILD_NUMBER}/.env
	scp -o StrictHostKeyChecking=no -i ./!deploy/deploy_rsa -P ${PORT} -r ./loki deploy@${HOST}:support_bot_${BUILD_NUMBER}/loki
	scp -o StrictHostKeyChecking=no -i ./!deploy/deploy_rsa -P ${PORT} -r ./grafana deploy@${HOST}:support_bot_${BUILD_NUMBER}/grafana
	rm -f docker-compose-production-env.yml

	ssh -o StrictHostKeyChecking=no -i ./!deploy/deploy_rsa deploy@${HOST} -p ${PORT} 'cd support_bot_${BUILD_NUMBER} && docker stack deploy --compose-file ./docker-compose.yml support_bot --with-registry-auth --prune'

deploy-clean:
	rm -f docker-compose-production-env.yml



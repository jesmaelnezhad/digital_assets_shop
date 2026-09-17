# Build all services
build:
	docker compose build

# Push all images to local registry
push:
	@echo "Tagging and pushing images to localhost:5000..."
	@services="identity-service product-service commerce-service community-service review-service payment-service admin-service media-service shell-mfe shop-mfe product-mfe community-mfe account-mfe checkout-mfe auth-mfe admin-mfe"; \
	for svc in $$services; do \
		docker tag pawradise-$$svc:latest localhost:5000/pawradise-$$svc:latest; \
		docker push localhost:5000/pawradise-$$svc:latest; \
	done

# Deploy to k3s
deploy:
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/configmap.yaml
	kubectl apply -f k8s/secrets.yaml
	kubectl apply -f k8s/postgres.yaml
	kubectl apply -f k8s/redis.yaml
	kubectl apply -f k8s/registry.yaml
	kubectl apply -f k8s/identity-service-deployment.yaml
	kubectl apply -f k8s/product-service-deployment.yaml
	kubectl apply -f k8s/commerce-service-deployment.yaml
	kubectl apply -f k8s/community-service-deployment.yaml
	kubectl apply -f k8s/review-service-deployment.yaml
	kubectl apply -f k8s/payment-service-deployment.yaml
	kubectl apply -f k8s/admin-service-deployment.yaml
	kubectl apply -f k8s/media-service-deployment.yaml
	kubectl apply -f k8s/shell-mfe-deployment.yaml
	kubectl apply -f k8s/shop-mfe-deployment.yaml
	kubectl apply -f k8s/product-mfe-deployment.yaml
	kubectl apply -f k8s/community-mfe-deployment.yaml
	kubectl apply -f k8s/account-mfe-deployment.yaml
	kubectl apply -f k8s/checkout-mfe-deployment.yaml
	kubectl apply -f k8s/auth-mfe-deployment.yaml
	kubectl apply -f k8s/admin-mfe-deployment.yaml
	kubectl apply -f k8s/ingress.yaml

# Delete all resources from k3s
undeploy:
	kubectl delete -f k8s/ --recursive

# View logs for all services
logs:
	kubectl logs -n pawradise -l app --all-containers=true --tail=100 -f

# View logs for a specific service (usage: make logs-svc svc=identity-service)
logs-svc:
	kubectl logs -n pawradise -l app=$(svc) -f --tail=100

# Check cluster status
status:
	kubectl get all -n pawradise
	kubectl get all -n database
	kubectl get all -n registry

# Run E2E tests
test:
	cd e2e && ./customer-journeys.sh

# Run integration tests
test-integration:
	cd tests/integration && go test ./...

# Run unit tests
test-unit:
	cd tests/unit && go test ./...

# Port-forward to access services locally
port-forward:
	kubectl port-forward -n pawradise svc/shell-mfe 30080:80 &
	kubectl port-forward -n pawradise svc/identity-service 8081:8081 &
	kubectl port-forward -n pawradise svc/product-service 8082:8082 &
	kubectl port-forward -n pawradise svc/postgres 5432:5432 &
	@echo "Services available at:"
	@echo "  Shell MFE:          http://localhost:30080"
	@echo "  Identity Service:   http://localhost:8081"
	@echo "  Product Service:    http://localhost:8082"
	@echo "  PostgreSQL:         localhost:5432"

# Stop port-forwarding
stop-port-forward:
	pkill -f "kubectl port-forward"

# Start local development environment
up:
	docker compose up -d

# Stop local development environment
down:
	docker compose down

# Restart a single service (usage: make restart svc=identity-service)
restart:
	docker compose restart $(svc)

# Clean everything
clean:
	docker compose down -v
	docker system prune -f

# Help
help:
	@echo "Pawradise Makefile targets:"
	@echo ""
	@echo "Development:"
	@echo "  make up               - Start local dev environment"
	@echo "  make down             - Stop local dev environment"
	@echo "  make build            - Build all Docker images"
	@echo "  make restart svc=X    - Restart a specific service"
	@echo ""
	@echo "Deployment:"
	@echo "  make push             - Push images to local registry"
	@echo "  make deploy           - Deploy to k3s"
	@echo "  make undeploy         - Remove from k3s"
	@echo "  make port-forward     - Access services locally"
	@echo ""
	@echo "Testing:"
	@echo "  make test             - Run E2E tests"
	@echo "  make test-unit        - Run unit tests"
	@echo "  make test-integration - Run integration tests"
	@echo ""
	@echo "Utilities:"
	@echo "  make logs             - Tail all service logs"
	@echo "  make logs-svc svc=X   - Tail specific service logs"
	@echo "  make status           - Show cluster status"
	@echo "  make clean            - Clean up everything"

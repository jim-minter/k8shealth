all: certs login server webhook

certs:
	go run ./cmd/certs webhook.default.svc
	kubectl delete secret webhook --ignore-not-found
	kubectl create secret tls webhook --cert=certs/webhook.default.svc.crt --key=certs/webhook.default.svc.key

login:
	@docker login -u 00000000-0000-0000-0000-000000000000 -p $(shell az acr login -n $(USER) --expose-token --query accessToken -o tsv 2>/dev/null) $(USER).azurecr.io

server:
	CGO_ENABLED=0 go build ./cmd/server
	docker build -t $(USER).azurecr.io/server:latest -f Dockerfile.server .
	docker push $(USER).azurecr.io/server:latest
	envsubst < server.yaml | kubectl apply -f -
	kubectl rollout restart deployment/server

webhook:
	CGO_ENABLED=0 go build ./cmd/webhook
	docker build -t $(USER).azurecr.io/webhook:latest -f Dockerfile.webhook .
	docker push $(USER).azurecr.io/webhook:latest
	CA_BUNDLE=$(shell base64 -w0 certs/@ca.crt) envsubst < webhook.yaml | kubectl apply -f -
	kubectl rollout restart deployment/webhook

.PHONY: all certs login server webhook

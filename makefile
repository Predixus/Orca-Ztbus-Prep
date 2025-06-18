.PHONY: data

data:
	docker compose up -d
	sqlc vet -f sqlc.yaml
	sqlc generate -f sqlc.yaml


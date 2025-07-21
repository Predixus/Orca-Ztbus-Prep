.PHONY: data

data:
	sqlc vet -f sqlc.yaml
	sqlc generate -f sqlc.yaml


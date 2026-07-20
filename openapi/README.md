# Tradera REST contract

`tradera-v4.json` is a pinned copy of Tradera's canonical OpenAPI document:

- Source: https://api.tradera.com/v4/swagger/v4/swagger.json
- API version: `v4-beta`
- OpenAPI version: `3.0.1`

`client-operations.json` assigns stable public method names and service clients because the upstream contract does not provide `operationId` values. Generation fails unless this metadata covers every REST v4 operation exactly once.

Run `go run ./internal/updatecontract && go generate ./...` to intentionally refresh the upstream contract and regenerate the clients. Review the contract, metadata, and generated diff before committing an update.
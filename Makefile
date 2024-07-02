docker:
	docker-compose up

server:
	go run cmd/main.go

watch: 
	air

mock-gen:
	mockgen -source ./internal/cache/cache.repository.go -destination ./mocks/cache/cache.repository.go
	mockgen -source ./internal/pin/pin.service.go -destination ./mocks/pin/pin.service.go
	mockgen -source ./internal/stamp/stamp.repository.go -destination ./mocks/stamp/stamp.repository.go
	mockgen -source ./internal/stamp/stamp.service.go -destination ./mocks/stamp/stamp.service.go
	mockgen -source ./internal/selection/selection.repository.go -destination ./mocks/selection/selection.repository.go
	mockgen -source ./internal/selection/selection.service.go -destination ./mocks/selection/selection.service.go

test:
	go vet ./...
	go test  -v -coverpkg ./internal/... -coverprofile coverage.out -covermode count ./internal/...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

proto:
	go get github.com/isd-sgcu/rpkm67-go-proto@latest

model:
	go get github.com/isd-sgcu/rpkm67-model@latest
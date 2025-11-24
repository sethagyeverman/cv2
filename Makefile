.PHONY: api
api:
	goctl api go --api api/service.api --dir . --style goZero

.PHONY: swagger
swagger:
	goctl api plugin -plugin goctl-swagger="swagger -filename cv2.json -host localhost:8888 -basepath /" -api api/service.api -dir ./docs

.PHONY: generate
generate:
	go generate ./...

.PHONY: run
run:
	go run .

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  api            - Generate go-zero code from API files"
	@echo "  swagger        - Generate Swagger/OpenAPI documentation"
	@echo "  generate       - Generate go code from API files"
	@echo "  run            - Run the application"
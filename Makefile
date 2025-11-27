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

.PHONY: build
build:
	go build -o bin/cv2 .

.PHONY: git
git:
	git add .
	git commit -m "$(m)"
	git push origin main

.PHONY: git-stats
git-stats:
	@echo "=== 本周提交 ==="
	@git log --oneline --since="1 week ago" | wc -l | xargs echo "commits:"
	@echo "\n=== 开发者提交情况 ==="
	@git log --since="1 week ago" --numstat --pretty="%aN" | awk 'NF==1 {author=$$0} NF==3 {plus[author]+=$$1; minus[author]+=$$2} END {for(a in plus) printf "%s: +%d / -%d\n", a, plus[a], minus[a]}'

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  api            - Generate go-zero code from API files"
	@echo "  swagger        - Generate Swagger/OpenAPI documentation"
	@echo "  generate       - Generate go code from API files"
	@echo "  run            - Run the application"
APP_URL := http://localhost:8080
PORT := 8080
export APP_URL
export PORT

.PHONY: build
build:
	go build -o injecture-web .

.PHONY: dev
dev: build
	./injecture-web

.PHONY: watch
watch:
	PORT=8081 gin -a 8081 -p 8080

.PHONY: docker
docker:
	git archive HEAD | docker build -t injecture:dev -

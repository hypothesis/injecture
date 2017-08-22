HEROKU_URL := http://localhost:8080
PORT := 8080
export HEROKU_URL
export PORT

.PHONY: build
build:
	go build .

.PHONY: dev
dev: build
	./injecture

.PHONY: watch
watch:
	PORT=8081 gin -a 8081 -p 8080

.PHONY: docker
docker:
	git archive HEAD | docker build -t injecture:dev -

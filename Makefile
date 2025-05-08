APP_NAME=bot

build:
	go build -o $(APP_NAME) .

run:
	go run main.go

docker-build:
	docker build -t $(APP_NAME):latest .

docker-run:
	docker run --rm --env-file .env $(APP_NAME):latest

fly-deploy:
	flyctl deploy

fly-logs:
	flyctl logs

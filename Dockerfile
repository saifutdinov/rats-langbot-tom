# syntax=docker/dockerfile:1

FROM golang:1.21-alpine

WORKDIR /app

# Копируем всё в контейнер
COPY . .

# Загружаем зависимости и компилируем
RUN go mod tidy
RUN go build -o bot .

# Устанавливаем tini — помогает корректно завершать процессы
RUN apk add --no-cache tini

# Устанавливаем entrypoint через tini
ENTRYPOINT ["/sbin/tini", "--"]

CMD ["./bot"]

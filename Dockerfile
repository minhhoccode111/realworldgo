FROM golang:1.25-alpine AS build

WORKDIR /app

# to run migrations
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main cmd/api/main.go

FROM alpine:3.20.1 AS prod
WORKDIR /app

# Install necessary packages for running goose
RUN apk add --no-cache make

# to run migrations
COPY --from=build /go/bin/goose /usr/local/bin/goose
COPY --from=build /app/internal/migrations /app/migrations
COPY --from=build /app/main /app/main
COPY --from=build /app/Makefile /app/Makefile
COPY --from=build /app/.env /app/.env

EXPOSE ${SERVER_PORT}

# Create a startup script that runs migrations then starts the app
RUN echo '#!/bin/sh' > /app/start.sh && \
    echo 'echo "Running migrations..."' >> /app/start.sh && \
    echo 'DB_URL="postgres://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}&search_path=${DB_SCHEMA}"' >> /app/start.sh && \
    echo 'goose -dir migrations postgres "$DB_URL" up' >> /app/start.sh && \
    echo 'echo "Starting application..."' >> /app/start.sh && \
    echo 'exec ./main' >> /app/start.sh && \
    chmod +x /app/start.sh

CMD ["/app/start.sh"]

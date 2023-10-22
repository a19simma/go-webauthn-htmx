FROM golang:1.21-alpine AS build
RUN apk update \
    && apk add --no-cache git \
    && apk add --no-cache ca-certificates \
    && apk add --update gcc musl-dev \
    && update-ca-certificates

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -o /main

FROM build AS test
RUN go test -v ./...

FROM alpine
COPY --from=build --chmod=777 /main /main
COPY --from=build /etc/passwd /etc/passwd
EXPOSE 4200
ENTRYPOINT ["./main"]


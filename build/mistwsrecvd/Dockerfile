# Stage 1
FROM golang:1.22-alpine3.20 AS gobuilder

WORKDIR /app/mist-to-tsdb/
COPY . .
ENV CGO_ENABLED 1
ENV GOFLAGS -mod=vendor
ENV GOOS=linux
ENV GOARCH=amd64
RUN apk -U add ca-certificates
RUN apk update && apk add pkgconf git bash build-base sudo
RUN go mod download
RUN go build -tags musl -o mistwsrecvd ./cmd/mistwsrecvd/main.go

# Stage 2
FROM alpine:3.20
WORKDIR /app
COPY --from=gobuilder /app/mist-to-tsdb/mistwsrecvd .
USER 1001
CMD [ "/app/mistwsrecvd" ]

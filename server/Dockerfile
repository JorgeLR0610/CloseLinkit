FROM golang:1.27.0 AS builder

ENV CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o server ./cmd/CloseLinkit

FROM alpine:3.24.1

RUN adduser -D appuser

WORKDIR /

COPY --from=builder /build/server ./

USER appuser

EXPOSE 8080

CMD [ "./server" ]

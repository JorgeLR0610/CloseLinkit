FROM golang:1.26.8 AS builder

ENV CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o server ./cmd/CloseLinkit

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=builder /build/server /server

EXPOSE 8080

ENTRYPOINT [ "/server" ]

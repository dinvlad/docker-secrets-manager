FROM golang:alpine AS builder

RUN go install filippo.io/age/cmd/age@latest && \
    go install github.com/foxboron/age-plugin-tpm/cmd/age-plugin-tpm@latest

WORKDIR /app

COPY go.* .
RUN go mod download

COPY *.go .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -o secrets-manager

FROM cgr.dev/chainguard/static

COPY --from=builder /go/bin/ /app/secrets-manager /
CMD ["/secrets-manager"]

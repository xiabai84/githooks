FROM golang:1.22-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG VERSION=dev
ARG GIT_COMMIT=unknown

RUN CGO_ENABLED=0 go build -ldflags "-s -w \
    -X github.com/xiabai84/githooks/buildinfo.version=${VERSION} \
    -X github.com/xiabai84/githooks/buildinfo.gitCommit=${GIT_COMMIT}" \
    -o githooks .

FROM alpine:3.20

RUN apk add --no-cache git bash python3

COPY --from=builder /build/githooks /usr/local/bin/githooks
COPY scripts/bump-version.py /usr/local/bin/bump-version.py
COPY scripts/commit-msg /usr/local/bin/commit-msg

ENTRYPOINT ["githooks"]

ARG GO_VERSION=1.22
ARG ALPINE_VERSION=3.19

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS build

WORKDIR /src

COPY go.sum go.mod ./

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build/ \
    go mod download -x && \
    go install github.com/a-h/templ/cmd/templ@$(grep 'github.com/a-h/templ' go.sum | head -1 | awk '{print $2}')

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build/ \
    go generate ./... && \
    GCO_ENABLED=0 go build -o /bin/server .

FROM alpine:${ALPINE_VERSION}

ARG UID=10001
RUN adduser --disabled-password --gecos "" --home /nonexistent --shell "/sbin/nologin" \
    --no-create-home --uid "${UID}" user
USER user

COPY --from=build /bin/server /bin/

EXPOSE 3000

ENTRYPOINT [ "/bin/server" ]

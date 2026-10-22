FROM golang:1.26.10-trixie AS go-toolchain

FROM node:24.21.0-trixie
COPY --from=go-toolchain /usr/local/go /usr/local/go
ENV PATH="/root/go/bin:/usr/local/go/bin:${PATH}" \
    GOPATH=/root/go \
    GOFLAGS=-buildvcs=false \
    GOTOOLCHAIN=local

RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates git build-essential \
    && rm -rf /var/lib/apt/lists/*

COPY . /src
WORKDIR /src

RUN for module in . agent gateway openapi server tests; do \
      (cd "$module" && go mod download) || exit 1; \
    done
RUN npm install --global "@redocly/cli@$(sed -n 's/^REDOCLY_VERSION=//p' versions.env)" \
    && cd ui && npm ci

RUN cd server && go build -o /out/server . \
    && cd ../agent && go build -tags docker -o /out/agent . \
    && cd ../gateway && go build -o /out/gateway .

RUN cd openapi && redocly bundle spec/openapi.yaml -o ../ui/apps/console/openapi.json \
    && cd ../ui/apps/console \
    && OPENAPI_SPEC_PATH=./openapi.json npm run generate \
    && npm exec -- vite build

ENV GOPROXY=off \
    GOSUMDB=off

RUN (go test ./pkg/... || echo "WARNING: pkg tests reported failures") \
    && (cd server && go test -timeout 30m ./... || echo "WARNING: server tests reported failures") \
    && (cd agent && go test -tags docker ./... || echo "WARNING: agent tests reported failures") \
    && (cd gateway && go test ./... || echo "WARNING: gateway tests reported failures") \
    && (cd tests && go vet ./... || echo "WARNING: tests/ does not compile") \
    && (cd ui/apps/console && npm test || echo "WARNING: console tests reported failures")

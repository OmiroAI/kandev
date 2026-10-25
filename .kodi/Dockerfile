# Build dependencies once with networking; execute QA offline against committed source.
FROM ghcr.io/kdlbs/kandev-ci@sha256:020cab12058c2c1913e6aef3e66f655652fbf6040ac7a4f294f6b907e6a3ba44
USER root
ENV GOTOOLCHAIN=local GOFLAGS="-trimpath -tags=fts5 -p=1" GOMAXPROCS=2 GOGC=50 GOMEMLIMIT=1536MiB NODE_OPTIONS=--max-old-space-size=6144
WORKDIR /opt/kandev/source
COPY . .
RUN cd apps && pnpm install --frozen-lockfile
RUN cd apps/backend && go mod download
RUN GOCACHE=/opt/kandev/go-build sh -c 'cd apps/backend && go test -run="^$" ./internal/task/models ./internal/task/dto ./internal/task/repository/sqlite ./internal/task/service ./internal/task/handlers ./internal/task/usage ./internal/analytics/...'
RUN python3 .kodi/prepare-locks.py
RUN mkdir -p /go/pkg && ln -s /root/go/pkg/mod /go/pkg/mod && \
    chmod a+rx /root /root/go && \
    cp -a /root/.cache/node/corepack /opt/kandev/corepack && \
    chmod -R a+rX /opt/kandev/corepack
ENV COREPACK_HOME=/opt/kandev/corepack GOPATH=/go GOMODCACHE=/go/pkg/mod
WORKDIR /workspace

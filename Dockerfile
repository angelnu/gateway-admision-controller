FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.25@sha256:bc45dfd319e982dffe4de14428c77defe5b938e29d9bc6edfbc0b9a1fc171cb3 AS build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
ENV GO111MODULE=on

COPY *.go go.mod *.sum ./

# Download
RUN go mod download

# Bulild
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH}  go build -o app -ldflags '-w -extldflags "-static"' ./cmd

#Test
RUN  CCGO_ENABLED=0 go test -v ./...

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
# debug tag adds a shell (not recommended for prod)
FROM --platform=${BUILDPLATFORM} gcr.io/distroless/static:nonroot@sha256:cba10d7abd3e203428e86f5b2d7fd5eb7d8987c387864ae4996cf97191b33764
WORKDIR /
COPY --from=build /workspace/app /app/app
USER nonroot:nonroot

ENTRYPOINT ["/app/app"]

ARG IMAGE_SOURCE
LABEL org.opencontainers.image.source=$IMAGE_SOURCE 

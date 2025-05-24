FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.24@sha256:4c0a1814a7c6c65ece28b3bfea14ee3cf83b5e80b81418453f0e9d5255a5d7b8 AS build

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
FROM gcr.io/distroless/static:nonroot@sha256:188ddfb9e497f861177352057cb21913d840ecae6c843d39e00d44fa64daa51c
WORKDIR /
COPY --from=build /workspace/app /app/app
USER nonroot:nonroot

ENTRYPOINT ["/app/app"]

ARG IMAGE_SOURCE
LABEL org.opencontainers.image.source=$IMAGE_SOURCE 

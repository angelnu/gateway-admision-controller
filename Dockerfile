FROM golang:1.23@sha256:8c10f21bec412f08f73aa7b97ca5ac5f28a39d8a88030ad8a339fd0a781d72b4 AS build

WORKDIR /workspace
ENV GO111MODULE=on

COPY *.go go.mod *.sum ./

# Download
RUN go mod download

# Bulild
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 go build -o app -ldflags '-w -extldflags "-static"' ./cmd

#Test
RUN  CCGO_ENABLED=0 go test -v ./...

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
# debug tag adds a shell (not recommended for prod)
FROM gcr.io/distroless/static:nonroot@sha256:6ec5aa99dc335666e79dc64e4a6c8b89c33a543a1967f20d360922a80dd21f02
WORKDIR /
COPY --from=build /workspace/app /app/app
USER nonroot:nonroot

ENTRYPOINT ["/app/app"]

ARG IMAGE_SOURCE
#https://github.com/k8s-at-home/template-container-image
LABEL org.opencontainers.image.source $IMAGE_SOURCE 

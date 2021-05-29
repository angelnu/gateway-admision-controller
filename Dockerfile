FROM golang:1.16 AS build

WORKDIR /workspace
ENV GO111MODULE=on

COPY *.go go.mod *.sum ./

# Build
RUN go mod download

RUN CGO_ENABLED=0 go build -o app -ldflags '-w -extldflags "-static"' .

#Test
RUN  CCGO_ENABLED=0 go test -v .

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
# debug tag adds a shell (not recommended for prod)
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=build /workspace/app /app/app
USER nonroot:nonroot

ENTRYPOINT ["/app/app"]

ARG IMAGE_SOURCE
#https://github.com/k8s-at-home/template-container-image
LABEL org.opencontainers.image.source $IMAGE_SOURCE 

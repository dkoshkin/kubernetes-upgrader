# syntax=docker/dockerfile:1

# Use distroless/static:nonroot image for a base.
FROM --platform=linux/amd64 gcr.io/distroless/static as linux-amd64
FROM --platform=linux/arm64 gcr.io/distroless/static as linux-arm64

FROM --platform=linux/${TARGETARCH} linux-${TARGETARCH}

# Run as nonroot user using numeric ID for compatibllity.
USER 65532

COPY kubernetes-upgrader /manager

ENTRYPOINT ["/manager"]

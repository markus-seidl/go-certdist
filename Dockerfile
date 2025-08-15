# Use a multi-platform-aware base image
FROM alpine:latest

# TARGETARCH is automatically provided by buildx
ARG TARGETPLATFORM

# Add ca-certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the pre-compiled binary from the build context
# The workflow will place the binaries in dist/<arch>/go-certdist
COPY dist/${TARGETPLATFORM}/certdist /certdist

# Expose the default server port
EXPOSE 8080

# The command to run the application
# The user will need to provide the command-line arguments (e.g., "server", "config.yml")
ENTRYPOINT ["/certdist"]

# Use the official Golang image as base image
FROM golang:latest as BUILDER
LABEL MAINTAINERS="spilett n/t IT team"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Download and install any required dependencies
RUN go get -d -v ./...

# Build the Go app
RUN go build -o backup .

FROM debian:stable-slim as RUNNER

# Non root user parameters
ARG UID
ARG GID
# Define timezone as env to be able to change at runtime
ENV TZ="Europe/Berlin"

USER root

# Install PostgreSQL client tools
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    postgresql-client-15 tzdata \
    && rm -rf /var/lib/apt/lists/*

RUN addgroup --gid $GID nonroot && \
    adduser --uid $UID --gid $GID --disabled-password --gecos "" nonroot

WORKDIR /app    

COPY --chown=$UID:$GID --from=BUILDER /app/backup /app/backup

RUN mkdir -p /backups
RUN chown -R $UID:$GID /backups

USER nonroot

CMD [ "./backup" ]
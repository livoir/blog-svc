#!/bin/sh

set -eu

# Validate required
if [ -z "${GITHUB_ACCESS_TOKEN:-}" ] || [ -z "${GITHUB_USER:-}" ]; then
    echo "GITHUB_ACCESS_TOKEN and GITHUB_USER must be set"
    exit 1
fi

# Variables
IMAGE_NAME="livoir-blog"
IMAGE_TAG="dev"
REGISTRY="ghcr.io/livoir"
FULL_IMAGE_NAME="${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"

# Build the Docker image
echo "Building Docker image ${FULL_IMAGE_NAME}"
docker build --platform linux/amd64 --build-arg BUILDKIT_INLINE_CACHE=1 --cache-from $FULL_IMAGE_NAME -t $FULL_IMAGE_NAME .

# Login to the GitHub Container Registry
echo "Logging in to the GitHub Container Registry"
echo $GITHUB_ACCESS_TOKEN | docker login ghcr.io -u $GITHUB_USER --password-stdin || { echo "Failed to authenticate with GHCR"; exit 1; }
trap 'docker logout ghcr.io' EXIT
# Push the Docker image to the registry
max_attempts=3
attempt=1
until docker push $FULL_IMAGE_NAME; do
    if [ $attempt -eq $max_attempts ]; then
        echo "Failed to push Docker image ${FULL_IMAGE_NAME} after $max_attempts attempts"
        exit 1 
    fi
    echo "Failed to push Docker image ${FULL_IMAGE_NAME}, attempt $attempt of $max_attempts. Retrying..."
    attempt=$((attempt + 1))
    sleep 5
done
echo "Successfully pushed image to ${FULL_IMAGE_NAME}"
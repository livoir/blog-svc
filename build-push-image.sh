#!/bin/sh

set -eu

# Validate required
if [ -z "${GITHUB_ACCESS_TOKEN:-}" ] || [ -z "${GITHUB_USER:-}" ]; then
    echo "Error: GITHUB_ACCESS_TOKEN and GITHUB_USER must be set"
    exit 1
fi

# Validate Docker installation
if ! command -v docker > /dev/null 2>&1; then
    echo "Error: Docker is not installed or not in PATH"
    exit 1
fi

# Validate docker daemon is running
if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker daemon is not running"
    exit 1
fi

# Variables
IMAGE_NAME="livoir-blog"
IMAGE_TAG="${IMAGE_TAG:-dev}"
REGISTRY="ghcr.io/livoir"
FULL_IMAGE_NAME="${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"


# Build the Docker image
echo "Building Docker image ${FULL_IMAGE_NAME}"
if ! docker build --platform linux/amd64 --build-arg BUILDKIT_INLINE_CACHE=1 --cache-from ${FULL_IMAGE_NAME} -t ${FULL_IMAGE_NAME} .; then
    echo "Error: Failed to build Docker image ${FULL_IMAGE_NAME}"
    exit 1
fi

# Login to the GitHub Container Registry
echo "Logging in to the GitHub Container Registry"
if ! timeout 60s sh -c 'docker login ghcr.io -u "${GITHUB_USER}" --password-stdin < /dev/fd/3' 3<<< "${GITHUB_ACCESS_TOKEN}"; then
    echo "Error: Failed to authenticate with GHCR"
    exit 1
fi
trap 'echo "Logging out from GHCR"; docker logout ghcr.io' EXIT
# Push the Docker image to the registry
max_attempts=3
attempt=1
until docker push "${FULL_IMAGE_NAME}"; do
    if [ $attempt -eq $max_attempts ]; then
        echo "Failed to push Docker image ${FULL_IMAGE_NAME} after $max_attempts attempts"
        exit 1 
    fi
    echo "Failed to push Docker image ${FULL_IMAGE_NAME}, attempt $attempt of $max_attempts. Retrying..."
    attempt=$((attempt + 1))
    case $attempt in
        1) sleep 1 ;;
        2) sleep 2 ;;
        3) sleep 4 ;;
    esac
done
echo "Successfully pushed image to ${FULL_IMAGE_NAME}"

# Verify the pushed image
echo "Verifying the pushed image"
if ! timeout 60s docker pull "${FULL_IMAGE_NAME}" > /dev/null 2>&1; then
    echo "Error: Failed to verify the pushed image ${FULL_IMAGE_NAME}. Pull test failed"
    exit 1
fi
echo "Successfully verified the pushed image ${FULL_IMAGE_NAME}"   

# Cleanup verification image
docker rmi "${FULL_IMAGE_NAME}" > /dev/null 2>&1 || true
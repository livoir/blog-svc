#!/bin/sh

# Variables
IMAGE_NAME="livoir-blog"
IMAGE_TAG="dev"
REGISTRY="ghcr.io/livoir"
FULL_IMAGE_NAME="${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"

# Build the Docker image
echo "Building Docker image ${FULL_IMAGE_NAME}"
docker build -t $FULL_IMAGE_NAME .

# Login to the GitHub Container Registry
echo "Logging in to the GitHub Container Registry"
echo $GITHUB_ACCESS_TOKEN | docker login ghcr.io -u $GITHUB_USER --password-stdin

# Push the Docker image to the registry
docker push $FULL_IMAGE_NAME
echo "Docker image pushed to ${FULL_IMAGE_NAME}"
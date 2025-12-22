#!/usr/bin/env bash
#
# Darts Web Deployment Script
# Usage: ./deploy.sh [TAG]
#   ./deploy.sh          Deploy latest git tag
#   ./deploy.sh v1.0.15  Deploy specific tag
#

set -euo pipefail

# Configuration
HELM_RELEASE="darts"
HELM_CHART="charts/darts-web"
VALUES_FILE="charts/darts-web/values.yaml"

# Parse arguments
if [[ $# -eq 1 ]]; then
    if [[ "$1" == "--help" ]]; then
        echo "Usage: $0 [TAG]"
        echo ""
        echo "Examples:"
        echo "  $0          # Deploy latest git tag"
        echo "  $0 v1.0.15  # Deploy specific tag"
        exit 0
    fi
    VERSION="$1"
else
    VERSION=""
fi

# Git pull and fetch tags FIRST
echo "→ Pulling latest changes..."
git pull --rebase origin main
git fetch --tags

# Determine version
if [[ -z "$VERSION" ]]; then
    VERSION=$(git describe --tags --abbrev=0)
    if [[ -z "$VERSION" ]]; then
        echo "Error: No git tags found. Create a tag first:"
        echo "  git tag v1.0.15 -m 'Release 1.0.15'"
        echo "  git push origin v1.0.15"
        exit 1
    fi
    echo "Using latest tag: $VERSION"
else
    # Validate that tag exists
    if ! git rev-parse "$VERSION" >/dev/null 2>&1; then
        echo "Error: Tag $VERSION does not exist"
        exit 1
    fi
    echo "Using specified tag: $VERSION"
fi

# Remove 'v' prefix for Docker tag if present
DOCKER_VERSION="${VERSION#v}"

echo ""
echo "Deployment: $VERSION (Docker: $DOCKER_VERSION)"
read -p "Continue? (y/n) " -n 1 -r
echo
[[ ! $REPLY =~ ^[Yy]$ ]] && exit 0

# Build Docker image
echo "→ Building Docker image..."
docker build --platform linux/amd64 \
	-t darts-app:${DOCKER_VERSION} \
	-t darts-app:latest .
echo "✓ Image tagged as: darts-app:${DOCKER_VERSION} and darts-app:latest"

# Export to tar
echo "→ Exporting image..."
docker save "darts-app:${DOCKER_VERSION}" -o "darts-app-${DOCKER_VERSION}.tar"

# Load into Minikube
echo "→ Loading into Minikube..."
minikube image load "darts-app-${DOCKER_VERSION}.tar"

# Update Helm values
echo "→ Updating Helm values..."
if command -v yq &> /dev/null; then
    yq eval ".image.tag = \"$DOCKER_VERSION\"" -i "$VALUES_FILE"
else
    sed -i.bak "s/tag: \".*\"/tag: \"$DOCKER_VERSION\"/" "$VALUES_FILE"
    rm -f "${VALUES_FILE}.bak"
fi

# Deploy with Helm
echo "→ Deploying with Helm..."
if helm status "$HELM_RELEASE" &> /dev/null; then
    helm upgrade "$HELM_RELEASE" "$HELM_CHART" --wait --timeout 5m
else
    helm install "$HELM_RELEASE" "$HELM_CHART" --wait --timeout 5m
fi

# Verify pods
echo "→ Checking pod status..."
sleep 3
kubectl get pods -l app.kubernetes.io/name=darts-web

# Cleanup tar file
echo "→ Cleaning up..."
rm -f "darts-app-${DOCKER_VERSION}.tar"

# Commit Helm values changes
echo "→ Committing Helm values update..."
git add "$VALUES_FILE"
git commit -m "chore: deploy version $VERSION" || echo "No changes to commit"

# Push to remote
echo "→ Pushing to remote..."
git push origin main

echo ""
echo "✓ Deployment completed: $VERSION"
echo "  Access: http://mini-pc/darts"

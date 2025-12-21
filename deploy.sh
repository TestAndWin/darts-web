#!/usr/bin/env bash
#
# Darts Web Deployment Script
# Usage: ./deploy.sh [--version X.Y.Z]
#

set -euo pipefail

# Configuration
HELM_RELEASE="darts-web"
HELM_CHART="charts/darts-web"
VALUES_FILE="charts/darts-web/values.yaml"

# Parse arguments
MANUAL_VERSION=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --version) MANUAL_VERSION="$2"; shift 2 ;;
        --help)
            echo "Usage: $0 [--version X.Y.Z]"
            exit 0
            ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

# Get current version and bump patch
CURRENT_VERSION=$(cat VERSION | tr -d '[:space:]')

if [[ -n "$MANUAL_VERSION" ]]; then
    NEW_VERSION="$MANUAL_VERSION"
else
    IFS='.' read -r major minor patch <<< "$CURRENT_VERSION"
    NEW_VERSION="${major}.${minor}.$((patch + 1))"
fi

echo "Deployment: $CURRENT_VERSION → $NEW_VERSION"
read -p "Continue? (y/n) " -n 1 -r
echo
[[ ! $REPLY =~ ^[Yy]$ ]] && exit 0

# Update VERSION file
echo "$NEW_VERSION" > VERSION

# Git pull
echo "→ Pulling latest changes..."
git pull --rebase origin main

# Build Docker image
echo "→ Building Docker image..."
make build

# Export to tar
echo "→ Exporting image..."
docker save "darts-app:${NEW_VERSION}" -o "darts-app-${NEW_VERSION}.tar"

# Load into Minikube
echo "→ Loading into Minikube..."
minikube image load "darts-app-${NEW_VERSION}.tar"

# Update Helm values
echo "→ Updating Helm values..."
if command -v yq &> /dev/null; then
    yq eval ".image.tag = \"$NEW_VERSION\"" -i "$VALUES_FILE"
else
    sed -i.bak "s/tag: \".*\"/tag: \"$NEW_VERSION\"/" "$VALUES_FILE"
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
rm -f "darts-app-${NEW_VERSION}.tar"

# Commit changes
echo "→ Committing version changes..."
git add VERSION "$VALUES_FILE"
git commit -m "chore: bump version to $NEW_VERSION"
git tag -a "v$NEW_VERSION" -m "Release $NEW_VERSION"

# Push to remote
echo "→ Pushing to remote..."
git push origin main
git push origin "v$NEW_VERSION"

echo ""
echo "✓ Deployment completed: v$NEW_VERSION"
echo "  Access: http://mini-pc/darts"

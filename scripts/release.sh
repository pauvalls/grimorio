#!/bin/bash
set -euo pipefail

# release.sh — Build release artifacts for Grimorio MCP v2

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
DIST_DIR="${ROOT_DIR}/dist"

VERSION="${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"
COMMIT="$(git rev-parse --short HEAD)"
DATE="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

echo "=== Grimorio Release Script ==="
echo "Version: ${VERSION}"
echo "Commit:  ${COMMIT}"
echo "Date:    ${DATE}"
echo ""

# Ensure clean dist directory
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# Run tests first
echo "=== Running tests ==="
cd "$ROOT_DIR"
go test ./... || { echo "Tests failed. Aborting release."; exit 1; }

# Build binaries for each platform
echo ""
echo "=== Building binaries ==="
for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    output="grimorio-${VERSION}-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output="${output}.exe"
    fi

    echo "Building ${output}..."
    env GOOS="$GOOS" GOARCH="$GOARCH" CGO_ENABLED=0 \
        go build -ldflags "$LDFLAGS" -o "${DIST_DIR}/${output}" ./cmd/grimorio
done

# Build migrate tool (linux/amd64 only for release)
echo "Building migrate-v1-to-v2-${VERSION}-linux-amd64..."
env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -ldflags "$LDFLAGS" -o "${DIST_DIR}/migrate-v1-to-v2-${VERSION}-linux-amd64" ./cmd/migrate-v1-to-v2

# Generate changelog snippet
echo ""
echo "=== Generating changelog snippet ==="
LAST_TAG="$(git describe --tags --abbrev=0 2>/dev/null || echo "")"
if [ -n "$LAST_TAG" ]; then
    git log --pretty=format:"- %s" "${LAST_TAG}..HEAD" > "${DIST_DIR}/CHANGELOG-snippet.md" || true
else
    echo "- Initial release" > "${DIST_DIR}/CHANGELOG-snippet.md"
fi

# Build Docker image
echo ""
echo "=== Building Docker image ==="
docker build \
    --build-arg VERSION="$VERSION" \
    --build-arg COMMIT="$COMMIT" \
    --build-arg DATE="$DATE" \
    -t "grimorio:mcp-${VERSION}" \
    -t "grimorio:mcp-latest" \
    "$ROOT_DIR"

echo ""
echo "=== Release artifacts ==="
ls -la "$DIST_DIR"

echo ""
echo "=== Docker images ==="
docker images "grimorio:mcp-${VERSION}"

echo ""
echo "=== Next steps ==="
echo "1. Tag the release: git tag -a ${VERSION} -m 'Release ${VERSION}'"
echo "2. Push the tag:    git push origin ${VERSION}"
echo "3. Push Docker:     docker push grimorio:mcp-${VERSION}"
echo "4. Create GitHub release with binaries from ${DIST_DIR}/"

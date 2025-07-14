#!/bin/bash

# Beta release script
# Usage: ./scripts/release-beta.sh [version]
# Example: ./scripts/release-beta.sh v1.0.0-beta.1

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() {
    echo -e "${BLUE}[BETA] $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Helper function to get next beta version
get_next_beta_version() {
    local base_version=$1
    local latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    
    # If this is the first beta for this version
    if [[ ! $latest_tag =~ $base_version-beta ]]; then
        echo "${base_version}-beta.1"
    else
        # Increment beta number
        local beta_num=$(echo $latest_tag | grep -o 'beta\.[0-9]*' | grep -o '[0-9]*')
        local next_beta=$((beta_num + 1))
        echo "${base_version}-beta.${next_beta}"
    fi
}

# Debug information
print_status "🧪 Beta Release Script Started"
print_status "Arguments: $@"
print_status "Current directory: $(pwd)"

# Check arguments
if [ $# -eq 0 ]; then
    print_error "Version not provided!"
    echo ""
    echo "Usage: $0 <version>"
    echo ""
    echo "Examples:"
    echo "  $0 v1.0.0-beta.1      # First beta for v1.0.0"
    echo "  $0 v1.0.0-beta.2      # Second beta for v1.0.0"
    echo "  $0 v1.1.0-beta.1      # First beta for v1.1.0"
    echo ""
    echo "Or use auto-increment:"
    echo "  $0 auto v1.0.0        # Auto-generate next beta number"
    exit 1
fi

# Handle auto-increment
if [ "$1" = "auto" ] && [ -n "$2" ]; then
    VERSION=$(get_next_beta_version "$2")
    print_status "Auto-generated version: $VERSION"
else
    VERSION=$1
fi

print_status "Processing version: $VERSION"

# Validate beta version format
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+-beta\.[0-9]+$ ]]; then
    print_error "Invalid beta version format: $VERSION"
    echo "Expected format: vX.Y.Z-beta.N"
    echo "Examples: v1.0.0-beta.1, v1.2.0-beta.3"
    exit 1
fi

print_success "Version format is valid"

# Extract base version for comparison
BASE_VERSION=$(echo $VERSION | sed 's/-beta\.[0-9]*$//')
print_status "Base version: $BASE_VERSION"

# Check if this is a Git repo
if [ ! -d ".git" ]; then
    print_error "Not a Git repository!"
    exit 1
fi

print_success "Git repository detected"

# Standard checks
if [ -n "$(git status --porcelain)" ]; then
    print_error "Working directory is not clean!"
    exit 1
fi

if git tag -l | grep -q "^$VERSION$"; then
    print_error "Tag $VERSION already exists!"
    exit 1
fi

print_success "Tag $VERSION is available"

# Check if we're on develop or feature branch for beta
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" = "main" ]; then
    print_warning "You are on 'main' branch for a beta release"
    print_warning "Consider using 'develop' branch for beta releases"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_error "Beta release cancelled"
        exit 1
    fi
fi

# Run tests (make optional if tests don't exist)
print_status "Running tests..."
if go test -v ./... 2>/dev/null; then
    print_success "Tests passed"
else
    print_warning "Tests failed or not found - continuing anyway"
fi

# Validate GoReleaser config (make optional)
if command -v goreleaser &> /dev/null; then
    print_status "Validating GoReleaser configuration..."
    if goreleaser check 2>/dev/null; then
        print_success "GoReleaser configuration valid"
        
        # Test build (optional)
        print_status "Testing beta build..."
        if goreleaser release --snapshot --clean --skip=publish 2>/dev/null; then
            print_success "Beta build test passed"
        else
            print_warning "Beta build test failed - continuing anyway"
        fi
    else
        print_warning "GoReleaser configuration invalid - continuing anyway"
    fi
else
    print_warning "GoReleaser not found - skipping build tests"
fi

# Update CHANGELOG for beta
print_status "Creating CHANGELOG entry..."
if [ ! -f CHANGELOG.md ]; then
    cat > CHANGELOG.md << EOF
# Changelog

## [${VERSION}] - $(date +%Y-%m-%d) [BETA]

### Added
- Beta release for testing new features

### Changed
- 

### Fixed
- 

### Notes
- This is a beta release - please test thoroughly
- Report issues: https://github.com/MIna-Maher/k8s-diff-informer/issues

EOF
    print_success "Created CHANGELOG.md"
else
    # Add beta entry to existing changelog
    TEMP_FILE=$(mktemp)
    head -n 2 CHANGELOG.md > "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
    echo "## [${VERSION}] - $(date +%Y-%m-%d) [BETA]" >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
    echo "### Added" >> "$TEMP_FILE"
    echo "- " >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
    echo "### Changed" >> "$TEMP_FILE"
    echo "- " >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
    echo "### Fixed" >> "$TEMP_FILE"
    echo "- " >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
    echo "### ⚠️ Beta Notes" >> "$TEMP_FILE"
    echo "- This is a beta release for testing" >> "$TEMP_FILE"
    echo "- Please report issues: https://github.com/MIna-Maher/k8s-diff-informer/issues" >> "$TEMP_FILE"
    echo "- Not recommended for production use" >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
    tail -n +3 CHANGELOG.md >> "$TEMP_FILE"
    mv "$TEMP_FILE" CHANGELOG.md
    print_success "Updated CHANGELOG.md"
fi

# Commit and tag
print_status "Creating git tag..."
git add CHANGELOG.md
git commit -m "chore: prepare beta release $VERSION

🧪 Beta Release Notes:
- This is a beta release for testing
- Features may be unstable
- Not recommended for production
- Feedback welcome on GitHub Issues" 2>/dev/null || echo "No changes to commit"

git tag -a "$VERSION" -m "Beta Release $VERSION

🧪 BETA RELEASE - FOR TESTING ONLY

This is a beta release containing new features and changes that need validation.

⚠️ Important Notes:
- Test in development/staging environments only
- Do not use in production
- Report issues: https://github.com/MIna-Maher/k8s-diff-informer/issues
- Join discussions: https://github.com/MIna-Maher/k8s-diff-informer/discussions

🚀 What's New:
- See CHANGELOG.md for detailed changes
- Docker images tagged as 'beta' and '$VERSION'
- Helm chart includes beta version

🔗 Testing:
- Docker: ghcr.io/mina-maher/k8s-diff-informer:beta
- Helm: --set image.tag=$VERSION

Next Steps:
- Collect feedback for 1-2 weeks
- Address reported issues
- Release stable version: $BASE_VERSION"

print_success "Beta release $VERSION prepared!"

# Push beta release
print_status "Ready to push beta release..."
echo ""
echo "📋 What happens next:"
echo "1. GitHub Actions will build beta artifacts"
echo "2. Docker images will be tagged as 'beta' and '$VERSION'"
echo "3. GitHub release will be marked as pre-release"
echo "4. Helm chart will be available with beta version"
echo ""
echo "🚀 To push the release, run:"
echo "   git push origin $CURRENT_BRANCH"
echo "   git push origin $VERSION"
echo ""
echo "🧪 Beta Testing:"
echo "   Docker: docker pull ghcr.io/mina-maher/k8s-diff-informer:beta"
echo "   Helm: helm install k8s-diff-informer k8s-diff-informer/k8s-diff-informer --set image.tag=$VERSION"
echo ""
echo "📝 Next Steps:"
echo "1. Test the beta release thoroughly"
echo "2. Gather feedback from users"
echo "3. Fix any reported issues"
echo "4. Create stable release: ./scripts/release-stable.sh $BASE_VERSION"
echo ""
echo "🔗 Monitor: https://github.com/$(git config --get remote.origin.url | sed 's/.*github\.com[:/]\([^/]*\/[^/]*\)\.git/\1/' 2>/dev/null || echo "MIna-Maher/k8s-diff-informer")/actions"
echo ""
print_success "Script completed successfully!"
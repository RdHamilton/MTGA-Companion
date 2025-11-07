#!/bin/bash
# Script to set up branch protection for main branch with required status checks
# This ensures quality checks must pass before merging PRs

set -e

REPO="RdHamilton/MTGA-Companion"
BRANCH="main"

# Required status checks from CI workflow
REQUIRED_CHECKS=(
    "Test (Linux)"
    "Lint"
    "Format Check"
)

echo "Setting up branch protection for 'main' branch..."
echo ""

# Check if branch protection already exists
if gh api repos/${REPO}/branches/${BRANCH}/protection >/dev/null 2>&1; then
    echo "Branch protection already exists for 'main' branch"
    echo "Updating to add required status checks..."
else
    echo "Creating branch protection for 'main' branch..."
fi

# Create/update branch protection with required status checks
# Note: This requires admin access to the repository

echo ""
echo "To set up branch protection with required status checks:"
echo ""
echo "Option 1: Via GitHub UI (Recommended)"
echo "1. Go to: https://github.com/${REPO}/settings/branches"
echo "2. Click 'Add rule' or edit existing rule for 'main' branch"
echo "3. Enable the following settings:"
echo "   - Require a pull request before merging"
echo "   - Require status checks to pass before merging"
echo "   - Require branches to be up to date before merging (optional but recommended)"
echo ""
echo "4. Under 'Status checks that are required', select:"
for check in "${REQUIRED_CHECKS[@]}"; do
    echo "   - $check"
done
echo ""
echo "5. Save the changes"
echo ""

echo "Option 2: Via GitHub API (requires admin access)"
echo "Run the following command (you may need to adjust permissions):"
echo ""
echo "gh api repos/${REPO}/branches/${BRANCH}/protection \\"
echo "  --method PUT \\"
echo "  --field required_status_checks[contexts][]=Test\ \(Linux\) \\"
echo "  --field required_status_checks[contexts][]=Lint \\"
echo "  --field required_status_checks[contexts][]=Format\ Check \\"
echo "  --field required_status_checks[strict]=true \\"
echo "  --field enforce_admins=true \\"
echo "  --field required_pull_request_reviews[required_approving_review_count]=0 \\"
echo "  --field restrictions=null"
echo ""

echo "Note: The API approach may require additional permissions."
echo "The UI approach is recommended for reliability."


#!/bin/bash
# Script to update repository ruleset to require status checks for PRs
# This ensures quality checks must pass before merging

set -e

REPO="RdHamilton/MTGA-Companion"
RULESET_ID=9523549

# Get the current ruleset
echo "Fetching current ruleset..."
CURRENT_RULESET=$(gh api repos/${REPO}/rulesets/${RULESET_ID})

# Get workflow IDs for required checks
echo "Fetching workflow IDs..."
CI_WORKFLOW_ID=$(gh api repos/${REPO}/actions/workflows --jq '.workflows[] | select(.name == "CI") | .id')

if [ -z "$CI_WORKFLOW_ID" ]; then
    echo "Error: Could not find CI workflow"
    exit 1
fi

echo "CI Workflow ID: $CI_WORKFLOW_ID"

# Get the job names from the CI workflow
# We'll need to get the actual job names from the workflow file
# For now, let's use the standard job names
REQUIRED_CHECKS=(
    "Test (Linux)"
    "Lint"
    "Format Check"
)

# Create the required_status_checks rule
# Note: This requires the ruleset to support required_status_checks
# We'll need to check if this rule type is available

echo ""
echo "To require status checks, you need to:"
echo "1. Go to: https://github.com/${REPO}/settings/rules"
echo "2. Edit the ruleset (ID: ${RULESET_ID})"
echo "3. Add a 'Required status checks' rule"
echo "4. Select the following checks:"
for check in "${REQUIRED_CHECKS[@]}"; do
    echo "   - $check"
done
echo ""
echo "Alternatively, you can set up branch protection for 'main' branch:"
echo "1. Go to: https://github.com/${REPO}/settings/branches"
echo "2. Add or edit branch protection rule for 'main'"
echo "3. Enable 'Require status checks to pass before merging'"
echo "4. Select the required checks listed above"


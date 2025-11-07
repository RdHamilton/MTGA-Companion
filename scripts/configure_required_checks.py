#!/usr/bin/env python3
"""
Script to configure required status checks for pull requests.
This updates the repository ruleset to require quality checks to pass before merging.
"""

import json
import subprocess
import sys

REPO = "RdHamilton/MTGA-Companion"
RULESET_ID = 9523549

# Required status checks from CI workflow
REQUIRED_CHECKS = [
    "Test (Linux)",
    "Lint",
    "Format Check",
]


def run_gh_command(cmd):
    """Run a gh CLI command and return the JSON result."""
    try:
        result = subprocess.run(
            ["gh", "api"] + cmd.split(),
            capture_output=True,
            text=True,
            check=True,
        )
        return json.loads(result.stdout)
    except subprocess.CalledProcessError as e:
        print(f"Error running command: {cmd}")
        print(f"Error: {e.stderr}")
        return None
    except json.JSONDecodeError as e:
        print(f"Error parsing JSON: {e}")
        return None


def get_current_ruleset():
    """Get the current ruleset configuration."""
    cmd = f"repos/{REPO}/rulesets/{RULESET_ID}"
    return run_gh_command(cmd)


def update_ruleset_with_required_checks():
    """Update the ruleset to include required status checks."""
    current = get_current_ruleset()
    if not current:
        print("Error: Could not fetch current ruleset")
        return False

    print(f"Current ruleset: {current.get('name')}")
    print(f"Current rules: {[r.get('type') for r in current.get('rules', [])]}")

    # Check if required_status_checks rule already exists
    has_required_checks = any(
        r.get("type") == "required_status_checks" for r in current.get("rules", [])
    )

    if has_required_checks:
        print("Required status checks rule already exists in ruleset")
        return True

    print("\nTo add required status checks, you have two options:")
    print("\nOption 1: Update the ruleset via GitHub UI")
    print(f"1. Go to: https://github.com/{REPO}/settings/rules")
    print(f"2. Click on ruleset ID {RULESET_ID}")
    print("3. Click 'Add rule' and select 'Required status checks'")
    print("4. Add the following checks:")
    for check in REQUIRED_CHECKS:
        print(f"   - {check}")
    print("5. Save the changes")

    print("\nOption 2: Set up branch protection (Recommended)")
    print(f"1. Go to: https://github.com/{REPO}/settings/branches")
    print("2. Click 'Add rule' or edit existing rule for 'main' branch")
    print("3. Enable 'Require status checks to pass before merging'")
    print("4. Select the following required checks:")
    for check in REQUIRED_CHECKS:
        print(f"   - {check}")
    print("5. Optionally enable 'Require branches to be up to date before merging'")
    print("6. Save the changes")

    print("\nNote: Branch protection is more flexible and easier to manage.")
    print("Rulesets are newer but may have limitations.")

    return False


if __name__ == "__main__":
    success = update_ruleset_with_required_checks()
    sys.exit(0 if success else 1)


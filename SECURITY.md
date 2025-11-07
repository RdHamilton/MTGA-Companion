# Security Policy

## Supported Versions

We take security seriously and provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| Latest  | :white_check_mark: |
| < Latest| :x:                |

**Note**: As a small homegrown project, we focus on maintaining the latest version. Older versions may not receive security updates.

## Reporting a Vulnerability

If you discover a security vulnerability, please **do not** open a public issue. Instead, please report it privately using one of the following methods:

### Preferred Method: GitHub Security Advisory

1. Go to the [Security tab](https://github.com/ramonehamilton/MTGA-Companion/security) in this repository
2. Click on "Report a vulnerability"
3. Fill out the security advisory form with details about the vulnerability

### Alternative: Email

If you prefer, you can email security concerns directly to the repository owner.

### What to Include

When reporting a vulnerability, please include:

- **Description** - Clear description of the security issue
- **Steps to reproduce** - How to trigger the vulnerability
- **Impact** - What could be compromised or affected
- **Suggested fix** - If you have ideas for how to fix it (optional)
- **Affected versions** - Which versions are affected

### What to Expect

- **Acknowledgment**: You'll receive an acknowledgment within 48 hours
- **Initial assessment**: We'll provide an initial assessment within 7 days
- **Updates**: We'll keep you informed of our progress
- **Resolution**: We'll work to address the issue as quickly as possible

### Disclosure Policy

- We will work with you to understand and resolve the issue quickly
- We will credit you for the discovery (unless you prefer to remain anonymous)
- We will not disclose the vulnerability publicly until a fix is available
- We will coordinate with you on the disclosure timeline

## Security Best Practices

### For Users

- **Keep updated**: Always use the latest version of MTGA-Companion
- **Review permissions**: Be aware of what file system access the application requires
- **Log files**: The application only reads MTGA log files; it does not modify them
- **No network access**: The application does not connect to external servers (unless explicitly configured)

### For Developers

- **Dependencies**: Keep dependencies up to date
- **Input validation**: Always validate and sanitize input
- **File operations**: Be careful with file system operations
- **Error handling**: Don't expose sensitive information in error messages
- **Testing**: Test security-critical code paths

## Known Security Considerations

### Current Security Posture

- **Local-only operation**: The application operates locally and reads log files
- **No network communication**: By default, no external network connections are made
- **Read-only access**: The application only reads MTGA log files; it does not modify game files
- **Go standard library**: Uses Go's standard library for file operations

### Security Vulnerabilities

We track known security vulnerabilities in our dependencies and the Go standard library. See our [GitHub Security Advisories](https://github.com/ramonehamilton/MTGA-Companion/security/advisories) for current issues.

**Current Known Issues:**
- GO-2025-4010 (net/url): Fixed in Go 1.24.8, but we're using Go 1.23.12 (latest stable). See [Issue #31](https://github.com/ramonehamilton/MTGA-Companion/issues/31) for tracking.

## Security Updates

Security updates will be released as:
- **Patch versions** (e.g., 1.0.1 → 1.0.2) for security fixes
- **Minor versions** (e.g., 1.0.0 → 1.1.0) may include security improvements
- **Major versions** (e.g., 1.0.0 → 2.0.0) may include breaking security changes

## Thank You

Thank you for helping keep MTGA-Companion secure! Your responsible disclosure helps protect all users of the application.


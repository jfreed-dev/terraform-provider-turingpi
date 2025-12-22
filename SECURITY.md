# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability, please report it responsibly:

1. **Do not** open a public issue
2. Email the maintainer directly or use GitHub's private vulnerability reporting
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

## Security Considerations

This provider handles sensitive data:

- **Credentials**: BMC username and password are used for authentication
- **Network Access**: Communicates with Turing Pi BMC over HTTPS
- **Firmware Files**: Handles firmware images for flashing

### Best Practices

1. **Use environment variables** for credentials instead of hardcoding in `.tf` files
2. **Never commit** `terraform.tfstate` files containing credentials to version control
3. **Use HTTPS** endpoints (the default) for BMC communication
4. **Verify firmware images** before flashing to nodes

## Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Release**: Depends on severity and complexity

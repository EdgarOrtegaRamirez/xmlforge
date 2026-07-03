# Security Policy

## Overview

XmlForge is a library for processing XML documents. This document describes security considerations when using this tool.

## XML Security Considerations

### XXE (XML External Entity) Attacks

XmlForge's parser does **not** resolve external entities by default. The parser only handles:
- Standard entity references (`&amp;`, `&lt;`, `&gt;`, `&apos;`, `&quot;`)
- Internal DTD subsets (not loaded)

This provides protection against XXE attacks where malicious XML references external files or URLs.

### Billion Laughs Attack (XML Bomb)

The parser does not process DTD definitions or expand entity references beyond the five standard XML entities. This provides protection against the billion laughs attack (also known as XML bomb or exponential entity expansion).

### Denial of Service

- **Max Depth**: The parser supports a configurable maximum depth to prevent stack overflow from deeply nested XML
- **Input Size**: The parser streams input from `io.Reader`, so it can handle large documents without loading everything into memory
- **Entity Expansion**: Only standard entities are supported, preventing entity expansion attacks

## Reporting Security Issues

If you discover a security vulnerability in XmlForge, please report it responsibly:

1. Do NOT open a public GitHub issue
2. Email security@xmlforge.example.com (or contact the maintainer directly)
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

We will acknowledge receipt within 48 hours and provide a timeline for resolution.

## Security Best Practices

When using XmlForge:

1. **Validate input**: Use the validator package to check XML before processing
2. **Limit depth**: Set maximum depth limits when parsing untrusted XML
3. **Set timeouts**: For network-sourced XML, implement read timeouts
4. **Sanitize output**: When converting XML to HTML/JSON, be aware of injection risks
5. **Use lenient mode carefully**: Lenient mode may accept malformed XML that could cause issues downstream

## Dependencies

XmlForge has **zero external dependencies** — it uses only the Go standard library. This minimizes the attack surface from third-party code.

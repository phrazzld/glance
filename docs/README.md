# Glance Documentation

Welcome to the Glance project documentation. This directory contains comprehensive guides, development philosophy, and technical references.

## 📚 Documentation Structure

### 🎯 [Development Philosophy](../DEVELOPMENT_PHILOSOPHY.md)
Core principles, coding standards, and mandatory practices that guide all development work.

### 🔧 [Development Guides](./guides/)
Practical guides for working with the project:
- [GitHub Actions](./guides/github-actions.md) - CI/CD pipeline configuration
- [Security Scanning](./guides/security-scanning.md) - Vulnerability scanning and security policies
- [Linting](./guides/linting.md) - Code quality and linting setup
- [Pre-commit Hooks](./guides/precommit.md) - Git hooks configuration
- [Pre-commit Testing](./guides/precommit-testing.md) - Testing pre-commit hooks
- [Security Suppressions](./guides/security-suppressions.md) - Security scanning configuration
- [Mocking Approach](./guides/mocking-approach.md) - Testing and mocking guidelines

### 🏗️ [Development Patterns](./leyline/)
Detailed development philosophy implementation:
- [Core Tenets](./leyline/tenets/) - Fundamental principles
- [Pattern Bindings](./leyline/bindings/) - Language-specific implementations
- [Go Patterns](./leyline/bindings/categories/go/) - Go-specific best practices

## 🚀 Quick Start

1. Read the [Development Philosophy](../DEVELOPMENT_PHILOSOPHY.md) first
2. Check the [Go Appendix](../DEVELOPMENT_PHILOSOPHY_APPENDIX_GO.md) for language-specific rules
3. Review relevant guides in the [guides/](./guides/) directory
4. Follow the patterns in [leyline/](./leyline/) for implementation details

## 📝 Contributing to Documentation

- Keep documentation close to code when possible
- Use clear, actionable language
- Include examples and rationale
- Follow the established directory structure
- Update this index when adding new documentation

## 🔍 Need Help?

- Check the main [README](../README.md) for project overview
- Review [CLAUDE.md](../CLAUDE.md) for AI coding assistant guidelines
- Look for package-specific documentation in respective directories

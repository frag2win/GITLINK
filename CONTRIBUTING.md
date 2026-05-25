# Contributing to GITLINK

First off, thank you for considering contributing to GITLINK! It's people like you that make GITLINK such a great tool for privacy-preserving, peer-to-peer Git hosting.

## Architecture Overview
GITLINK is built with a hybrid Go + Rust architecture:
- **Rust (`services/git-server`)**: Handles all raw Git binary data and filesystem operations.
- **Go (`services/api-server`)**: The API Gateway, providing an HTTP interface and bridging Go ↔ Rust.
- **Go (`services/libp2p-node`)**: The networking edge handling P2P, NAT traversal, and DHT discovery.
- **React (`ui`)**: The frontend SPA.

## Getting Started

1. **Fork the repository** on GitHub.
2. **Clone your fork** locally.
3. Ensure you have the following prerequisites installed:
   - Go 1.22+
   - Rust 1.77+
   - Node.js 20+
   - Docker and Docker Compose
4. **Set up your environment**:
   ```bash
   make setup
   ```
5. **Run the development environment**:
   ```bash
   make dev
   ```

## Development Workflow

- Create a descriptive branch for your work: `git checkout -b feature/your-feature-name`
- Make sure your code is well-tested:
  - Run Go tests: `make test-api` or `make test-libp2p`
  - Run Rust tests: `make test-git`
  - Run all tests: `make test`
- Keep your commits atomic and write meaningful commit messages.

## Pull Requests

1. **Update documentation** if your changes modify any architecture, API, or setup instructions.
2. Ensure your code passes all linting and formatting checks (`make lint`, `make fmt`).
3. Fill out the **Pull Request Template** completely.
4. Your PR will be reviewed by maintainers, and you may be asked to make some adjustments.

## Reporting Bugs and Requesting Features

Please use the provided issue templates when reporting bugs or requesting new features. Provide as much context as possible!

Thank you for contributing!

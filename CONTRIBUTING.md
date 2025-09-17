# Contributing to Sloth Runner 🦥

First off, thank you for considering contributing! Every contribution helps make Sloth Runner better.

This document provides guidelines for contributing to the project.

## How to Contribute

There are many ways to contribute, from writing documentation to submitting bug reports and feature requests or writing code.

-   **🐛 Reporting Bugs:** If you find a bug, please open an issue on GitHub. Include as much detail as possible: your operating system, the command you ran, the Lua script you used, the full output, and what you expected to happen.
-   **💡 Suggesting Enhancements:** If you have an idea for a new feature or an improvement, open an issue to discuss it. This allows us to coordinate efforts and ensure the feature aligns with the project's goals.
-   ** Pull Requests:** Pull requests are the best way to propose changes to the codebase.

## Development Setup

To get started with development, you'll need Go (version 1.21+) installed.

1.  **Fork the repository:** Click the "Fork" button on the top right of the GitHub repository page.
2.  **Clone your fork:**
    ```bash
    git clone https://github.com/YOUR_USERNAME/sloth-runner.git
    cd sloth-runner
    ```
3.  **Build the project:**
    ```bash
    go build ./cmd/sloth-runner
    ```
4.  **Run the tests:**
    ```bash
    go test -v ./...
    ```

## Pull Request Process

1.  Ensure any install or build dependencies are removed before the end of the layer when doing a build.
2.  Update the `README.md` with details of changes to the interface, this includes new environment variables, exposed ports, useful file locations and container parameters.
3.  Create a new branch for your changes (`git checkout -b feature/my-awesome-feature`).
4.  Make your changes and add tests for them.
5.  Ensure the test suite passes (`go test -v ./...`).
6.  Commit your changes (`git commit -m "feat: Add some awesome feature"`).
7.  Push to the branch (`git push origin feature/my-awesome-feature`).
8.  Open a pull request to the `master` branch of the main repository.

Thank you for your contribution!

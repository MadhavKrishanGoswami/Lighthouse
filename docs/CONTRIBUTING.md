# Contributing to Lighthouse

Thank you for considering contributing to **Lighthouse**! It's contributions from passionate developers like you that make open-source projects thrive. Whether you're reporting bugs, suggesting features, improving documentation, or writing code, your efforts are greatly appreciated.

This document outlines how you can contribute. Feel free to propose improvements by submitting a pull request.

---

## 📜 Code of Conduct

This project and all participants are governed by the [Lighthouse Code of Conduct](https://github.com/MadhavKrishanGoswami/Lighthouse/blob/main/CODE_OF_CONDUCT.md). Please read and adhere to it. We strive for a welcoming, inclusive, and collaborative environment, and appreciate your cooperation.

---

## 💡 How Can I Contribute?

There are many ways you can help improve Lighthouse:

- 🐞 Report bugs and unexpected behavior  
- 🚀 Suggest new features or enhancements  
- 📖 Improve or expand the documentation  
- ✍️ Write tutorials, blog posts, or guides  
- 🔧 Submit code improvements or new features

Every contribution helps, whether big or small!

---

## 🐛 Reporting Bugs

Before reporting an issue:

1. Check the existing issues on [GitHub Issues](https://github.com/MadhavKrishanGoswami/Lighthouse/issues) to see if it’s already reported.
2. If it’s not, open a new issue and include:
   - A descriptive title  
   - A clear explanation of the problem  
   - Steps to reproduce the issue  
   - Relevant code samples or configuration files  
   - Expected vs actual behavior

This will help maintainers understand and resolve the issue quickly.

---

## ✨ Suggesting Enhancements

Enhancements and new features are always welcome! Please open a new issue and include:

- A summary of the enhancement  
- The problem it solves or why it’s useful  
- Any implementation ideas or mockups

This helps us prioritize and plan future improvements.

---

## 📂 Pull Requests

Pull requests are a fantastic way to contribute!

1. Fork the repository and create your feature branch from `main`.  
2. Follow the existing code style and lint your code before submission.  
3. Write clear and descriptive commit messages.  
4. Reference any related issues in your PR.

Example workflow:

```bash
git clone https://github.com/MadhavKrishanGoswami/Lighthouse.git
cd Lighthouse
git checkout -b my-feature-branch
````

Once your work is ready, submit the pull request and describe the changes you’ve made.

---

## 🛠 Development Setup

Follow these steps to set up Lighthouse for development. This guide assumes you have a working Go environment and Docker installed.

### 1. Prerequisites

Install the following tools on your local machine:

* [Go (v1.24+)](https://golang.org/dl/)
* [Docker & Docker Compose](https://www.docker.com/products/docker-desktop)
* [protoc (Protobuf compiler)](https://grpc.io/docs/protoc-installation/)
* [golang-migrate/migrate CLI](https://github.com/golang-migrate/migrate)
* [sqlc](https://sqlc.dev/)

---

### 2. Clone the Repository

```bash
git clone https://github.com/MadhavKrishanGoswami/Lighthouse.git
cd Lighthouse
```

---

### 3. Generate Code from Definitions

Lighthouse uses gRPC and `sqlc` for code generation. Run the following commands to generate Go code:

```bash
# Generate Go code from .proto files
make protos

# Generate Go code from SQL queries
make sqlc/orchestrator
```

---

### 4. Set Up the Database

You can easily set up PostgreSQL using Docker with this command:

```bash
make make-db
```

This will start a PostgreSQL container, create the `lighthouse` database, and apply migrations.

Alternatively, manage the database manually using:

```bash
make postgres     # Start PostgreSQL container
make createdb    # Create the database
make dropdb      # Drop the database
make migrate-up  # Apply migrations
make migrate-down # Roll back migrations
```

---

### 5. Run Lighthouse Services

Open multiple terminal windows and run the services as follows:

```bash
# Terminal 1: Run the Registry Monitor
make run-registry-monitor

# Terminal 2: Run the Orchestrator
make run-orchestrator

# Terminal 3: Run the Host Agent
make run-host-agent

# Terminal 4: Run the TUI
make run-tui
```

You now have Lighthouse fully running locally and can start contributing!

---

## 🙏 Thank You!

Every contribution, whether it’s reporting a bug or writing new features, helps make Lighthouse better. We appreciate your time, effort, and enthusiasm.

Let’s build something amazing together! 🚀📦

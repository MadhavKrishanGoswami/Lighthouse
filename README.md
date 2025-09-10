<p align="center">
<img src="https://github.com/user-attachments/assets/1c7c23c0-da2b-4972-818a-9eacf527d31d" alt="Lighthouse Logo" width="400">
</p>
<h1 align="center">Lighthouse</h1>
<p align="center">
  <strong>A microservices-oriented Docker orchestrator that manages and updates your entire container infrastructure with zero hassle.</strong>
  <br /><br />
  <a href="https://github.com/MadhavKrishanGoswami/Lighthouse/issues">Report a Bug</a>
  ·
  <a href="https://github.com/MadhavKrishanGoswami/Lighthouse/issues">Request a Feature</a>
</p>

---

## 📋 Table of Contents
- [✨ Key Features](#-key-features)
- [🚀 Getting Started](#-getting-started)
- [📸 Screenshots](#-screenshots)
- [⚙️ How It Works](#-how-it-works)
- [🤝 Contributing](#-contributing)
- [📜 License](#-license)
- [📞 Contact](#-contact)

---

## ✨ Key Features

✅ **Fleet Update Management**  
Manage and update containers across multiple hosts from one centralized place, with minimal setup and maximum control.

✅ **Sleek Text-Based UI (TUI)**  
Monitor and control your containers in real-time through an intuitive and lightweight terminal interface.

✅ **Automated Rollouts**  
Lighthouse watches your containers and automatically pulls new images, gracefully shuts down running containers, and restarts them using the same configuration as before.

✅ **Lightweight Host Agents**  
A small agent runs on each host, handling updates locally while communicating securely with the orchestrator via gRPC.

✅ **Decentralized Execution with Central Control**  
The orchestrator maintains state and schedules updates, while containers are updated locally on each host without performance bottlenecks.

✅ **Simple Setup**  
Get started quickly with minimal configuration and step-by-step deployment instructions.

✅ **Built with Go & gRPC**  
A fast, reliable, and scalable architecture ensures smooth operation across any number of hosts.

---

## 🚀 Getting Started

Deploying Lighthouse is quick and straightforward. Our setup guide walks you through installation, configuration, and deployment in both development and production environments.

➡️ Read the full guide in [DEPLOY.md](https://github.com/MadhavKrishanGoswami/Lighthouse/docs/DEPLOYMENT.md).

---

## 📸 Screenshots

Experience Lighthouse’s sleek TUI in action, giving you an overview of your fleet, container statuses, and updates—all in real-time.

<p align="center">
  <img src="https://placehold.co/800x400/0f172a/ffffff?text=Lighthouse+TUI+Screenshot" alt="Lighthouse TUI Screenshot">
</p>

---

## ⚙️ How It Works

1. **Orchestrator**  
   Keeps track of all hosts and containers using a PostgreSQL database.

2. **Registry Monitor**  
   Periodically scans container registries for new image versions.

3. **Host Agents**  
   On detecting an update, the orchestrator instructs the host agents to pull the new image, gracefully stop the container, and restart it with its original configuration.

4. **Real-time Visibility**  
   All updates, state changes, and logs are streamed back to the TUI, giving you full insight into your container fleet’s health and operations.

➡️ For technical details, see [ARCHITECTURE.md](https://github.com/MadhavKrishanGoswami/Lighthouse/docs/ARCHITECTURE.md).

---

## 🤝 Contributing

Lighthouse is an open-source project powered by developers like you! Whether you're fixing bugs, enhancing features, or sharing ideas, every contribution makes this tool better.

Please fork the repository and submit a pull request or open an issue labeled “enhancement” to share your thoughts.

➡️ Learn more in [CONTRIBUTING.md](https://github.com/MadhavKrishanGoswami/Lighthouse/docs/CONTRIBUTING.md).

---

## 📜 License

Distributed under the MIT License. See the [LICENSE](https://github.com/MadhavKrishanGoswami/Lighthouse/blob/main/LICENSE) file for details.

---

## 📞 Contact

**Madhav Krishan Goswami**  
Twitter: [@Goswamimadhav24](https://twitter.com/Goswamimadhav24)  
Email: goswamimadhavkrishan@gmail.com

🔗 [Project on GitHub](https://github.com/MadhavKrishanGoswami)

---

Keep your containers up-to-date, effortlessly and elegantly—with Lighthouse. 🚀📦

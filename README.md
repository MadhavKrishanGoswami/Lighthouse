<p align="center">
  <img src="https://github.com/user-attachments/assets/1c7c23c0-da2b-4972-818a-9eacf527d31d" alt="Lighthouse Logo" width="400">
</p>

<h1 align="center">Lighthouse</h1>

<p align="center">
  <strong>A microservices-oriented Docker orchestrator that seamlessly manages and updates your entire container infrastructure.</strong>
  <br /><br />
  <a href="https://github.com/MadhavKrishanGoswami/Lighthouse/issues">Report a Bug</a>
  ·
  <a href="https://github.com/MadhavKrishanGoswami/Lighthouse/issues">Request a Feature</a>
</p>

---

## ✨ Features

**Fleet Update Management**  
Manage and update containers across multiple hosts from one centralized control plane, minimizing setup and maximizing efficiency.

**Sleek Terminal Interface**  
Monitor and manage your containers in real time through an intuitive and lightweight Text-based UI (TUI), designed for clarity and ease of use.

**Automated, Graceful Rollouts**  
Lighthouse watches your containers, pulls new images when available, gracefully shuts down running containers, and restarts them using the exact configuration as they were deployed.

**Lightweight Host Agents**  
Deploy small, efficient agents on each host that communicate securely with the orchestrator via gRPC while executing updates locally for faster and more reliable performance.

**Centralized State, Local Execution**  
The orchestrator handles scheduling and state management, while updates and deployments occur locally on each host, reducing network overhead and improving resilience.

**Simple Setup**  
Quick to deploy with minimal configuration. Step-by-step guides help you get Lighthouse running in minutes.

**Built with Go & gRPC**  
A modern tech stack ensures performance, scalability, and reliability across distributed environments.

---
## 🤝 Contributing

We welcome contributions from the community! Whether you’re fixing bugs, suggesting improvements, or adding new features, your input helps make Lighthouse better for everyone.

Please fork the repository, submit a pull request, or open an issue labeled “enhancement.”

➡️ See how you can contribute in [CONTRIBUTING.md](https://github.com/MadhavKrishanGoswami/Lighthouse/blob/main/docs/CONTRIBUTING.md).

---

## 🚀 Getting Started

Setting up Lighthouse is fast and hassle-free. Follow our comprehensive guide to install, configure, and deploy Lighthouse in both development and production environments.

➡️ See the full instructions in [DEPLOYMENT.md](https://github.com/MadhavKrishanGoswami/Lighthouse/blob/main/docs/DEPLOYMENT.md).

---

## 📹 Watch Lighthouse in Action

Experience how Lighthouse manages container updates across a fleet of hosts with ease. Watch this demo showcasing its sleek interface and automated rollouts.

<p align="center">


https://github.com/user-attachments/assets/76a66687-d465-419e-b20a-8357b972c83c

  
</p>

---

## ⚙️ How It Works

**Orchestrator**  
Keeps track of all hosts and containers using a PostgreSQL database, ensuring consistency and coordination across your infrastructure.

**Registry Monitor**  
Periodically checks container registries for new image versions and triggers updates when changes are detected.

**Host Agents**  
Upon receiving update instructions, agents pull the new image, gracefully stop the container, and restart it with its original configuration, ensuring minimal downtime.

**Real-time Monitoring**  
State changes, logs, and update processes are streamed back to the TUI for instant visibility, helping you stay in control at all times.

➡️ Learn more about the architecture in [ARCHITECTURE.md](https://github.com/MadhavKrishanGoswami/Lighthouse/blob/main/docs/ARCHITECTURE.md).

---

## 📜 License

Lighthouse is open source and distributed under the MIT License. See the [LICENSE](https://github.com/MadhavKrishanGoswami/Lighthouse/blob/main/LICENSE) file for details.

---

## 📞 Contact

**Madhav Krishan Goswami**  
Twitter: [@Goswamimadhav24](https://twitter.com/Goswamimadhav24)  
Email: goswamimadhavkrishan@gmail.com

🔗 [Explore the project on GitHub](https://github.com/MadhavKrishanGoswami/Lighthouse)

---

Keep your container infrastructure updated, resilient, and effortless—with Lighthouse. 🚀📦

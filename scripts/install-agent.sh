#!/bin/bash

# --- This script must be run with sudo ---
if [ "$EUID" -ne 0 ]; then
  echo "Please run this installer with sudo."
  exit
fi

echo "--- Lighthouse Host Agent Installer ---"

# --- STEP 1: Set the orchestrator value by asking the user ---
read -p "Enter the Orchestrator URL (e.g., 192.168.1.100:50051): " ORCHESTRATOR_URL

if [ -z "$ORCHESTRATOR_URL" ]; then
  echo "Orchestrator URL cannot be empty. Aborting."
  exit 1
fi

echo "Configuration received. Installing..."

# --- STEP 2: Create the configuration file ---
CONFIG_DIR="/etc/lighthouse-host-agent"
CONFIG_FILE="$CONFIG_DIR/config.yaml"
mkdir -p $CONFIG_DIR
cat >$CONFIG_FILE <<EOL
# Configuration for the Lighthouse Host Agent
orchestrator_addr: "$ORCHESTRATOR_URL"
EOL

# --- STEP 3: Install the binary (the "image" you mentioned) ---
# Assumes 'host-agent-linux' is in the same folder as this script.
# We rename it to 'lighthouse-agent' for cleaner commands.
install host-agent-linux /usr/local/bin/lighthouse-agent

# --- STEP 4: "Open" the agent by installing and starting it as a service ---
# We are now calling the program we just installed.
echo "Registering and starting the system service..."
/usr/local/bin/lighthouse-agent install
/usr/local/bin/lighthouse-agent start

echo ""
echo "✅ Lighthouse Host Agent installation complete!"
echo "The service is now running in the background."

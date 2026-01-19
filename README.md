# Istio Config Manager

**A lightweight Go CLI tool to automate Istio L4/L7 routing rule configurations using Kubernetes API and Istio CRDs.**

---

## 🚀 Overview

**Istio Config Manager** is a standalone Golang-based tool designed to help DevOps teams and platform engineers automate the management of traffic routing in a Kubernetes cluster using **Istio**. It reads all configuration from environment variables and programmatically applies **VirtualService** and **DestinationRule** configurations to the target cluster.

This project demonstrates how automation can simplify mesh management and integrates well into CI/CD workflows.

---

## 🔧 Features

- Configure **L4/L7 routing** for services with full control over hosts, ports, routes, and weights
- Supports **Istio Sidecar** and **Ambient** modes
- Works out-of-the-box with **environment variable-driven** configuration
- Uses Kubernetes and Istio Go clients for direct CRD management
- Ideal for **GitOps**, CI/CD integration, and custom mesh policy automation
- Safe, declarative, and idempotent operations

---

## 🏗️ Architecture


```
+-------------------------+
| Environment Variables |
+-------------------------+
↓
+-------------------------+
| Config Parser |
+-------------------------+
↓
+-------------------------+
| Kubernetes API (Go SDK) |
+-------------------------+
↓
+-------------------------+
| Istio CRD Applier |
+-------------------------+
↓
+-------------------------+
| Cluster VirtualService |
| & DestinationRule CRDs |
+-------------------------+
```

---

## ⚙️ Environment Variables

You can configure your routing setup using the following environment variables:

| Variable Name             | Description                                      |
|--------------------------|--------------------------------------------------|
| `NAMESPACE`              | Target Kubernetes namespace                      |
| `SERVICE_NAME`           | Name of the Kubernetes service                   |
| `HOST`                   | Fully-qualified domain or hostname               |
| `DESTINATION_HOST`       | Service destination (e.g. `user-service.default`)|
| `DESTINATION_PORT`       | Port to route traffic to                         |
| `ROUTE_WEIGHT`           | Traffic weight percentage (e.g., 100)           |
| `MATCH_PREFIX`           | Path prefix for routing (e.g., `/api/`)         |

> 💡 You can extend this list based on your use case—e.g., adding subsets, retries, timeouts, etc.

---

## 🧪 Example

```bash
export NAMESPACE=default
export SERVICE_NAME=user-service
export HOST=user.example.com
export DESTINATION_HOST=user-service.default.svc.cluster.local
export DESTINATION_PORT=8080
export ROUTE_WEIGHT=100
export MATCH_PREFIX=/api/

./istio-config-manager apply
---

## Installation
git clone https://github.com/syedomair/istio-config-manager.git
cd istio-config-manager
go build -o istio-config-manager main.go

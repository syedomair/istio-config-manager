# Istio Config Manager

**Istio Config Manager** is a production-oriented Go-based CLI tool that programmatically manages Istio traffic policies by directly interacting with Kubernetes and Istio CRDs.  
It enables automated, idempotent configuration of **L4/L7 routing**, **traffic shifting**, **fault injection**, **timeouts**, **retries**, **mirroring**, and **circuit breaking** — all driven by environment variables and suitable for CI/CD and GitOps workflows.

---

## 🚀 Why Istio Config Manager?

Managing Istio resources manually using YAML can quickly become error-prone, repetitive, and difficult to automate at scale.  
Istio Config Manager solves this by offering:

- **Code-driven mesh policy management**
- **Declarative, idempotent updates**
- **Native Go integration with Kubernetes & Istio APIs**
- **Seamless CI/CD and GitOps compatibility**

This project demonstrates how platform teams can treat **service mesh configuration as code**, not static manifests.

---

## ✨ Key Features

- ✅ Programmatic creation and updates of **VirtualService** and **DestinationRule**
- ✅ L4 / L7 traffic routing
- ✅ Canary & A/B traffic splitting
- ✅ Header-based routing with URI rewrites
- ✅ Fault injection (delay)
- ✅ Traffic mirroring (shadow traffic)
- ✅ Request timeouts and retry policies
- ✅ Circuit breaker configuration
- ✅ Safe, idempotent operations (no duplicate rules)
- ✅ Works with **Istio Sidecar and Ambient Mesh**
- ✅ Environment-variable driven (CI/CD friendly)

---

## 🏗️ High-Level Architecture
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

 💡 You can extend this list based on your use case—e.g., adding subsets, retries, timeouts, etc.

```bash

## 🧪 Example

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





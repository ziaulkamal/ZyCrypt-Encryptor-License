## About

**ZyCrypt** (Ziya Encryptor) is a lightweight, self-hosted software license management system built with Go — distributed as a single binary with no runtime dependencies.

Designed to protect and distribute **Laravel + Vue + Inertia.js** applications through domain-bound license keys, real-time validation, and a built-in HTTP API server.

### Why ZyCrypt?

Most license managers are either too complex, require SaaS subscriptions, or depend on heavy server stacks. ZyCrypt runs as a single Go binary — deploy it anywhere with `zycrypt serve`.

### Key Features

- 🔑 **License Key Generation** — Auto-generated keys in `ZYC-XXXX-XXXX-XXXX-XXXX` format with HMAC checksum
- 🌐 **Domain Binding** — Licenses are locked to specific domains, preventing unauthorized redistribution
- ⚡ **Real-time Validation API** — `POST /api/v1/validate` with AES-256 encrypted responses
- 🛡️ **Dual-layer Protection** — `zycrypt-laravel` (Composer) + `zycrypt-vue` (NPM)
- 📋 **Audit Logging** — Every license event is recorded with IP, domain, and timestamp
- 🖥️ **CLI-first** — Full management via terminal: plans, licenses, domains, keys
- 📦 **Zero Dependencies** — Single binary, no PHP, no Node, no web server required

### Tech Stack

`Go 1.22` · `Cobra` · `Chi` · `GORM` · `PostgreSQL` · `Viper` · `HMAC-SHA256` · `AES-256`

### Scope

```
v1.0.0  ✅  Core license engine — CLI + API + domain binding
v1.1.0  🔜  Theme delivery — encrypted bundle + one-time token
v1.2.0  🔜  Hardening — structured logging, Prometheus metrics
```

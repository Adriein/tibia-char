# Home Server - Ansible Deployment

Automated deployment of **tibia-char** to a home server via Ansible.

## Prerequisites

1. Create `ansible_vault_password.txt` (gitignored) with your vault password.

## Usage

```bash
# First-time setup on a fresh server (prompts for SSH + sudo passwords)
make bootstrap

# Full deployment (build + deploy + migrate + health check)
make run

# Build artifact only (localhost)
make build-only

# Deploy without rebuilding (uses existing artifact)
make deploy-no-build

# Update env vars and restart containers only
make config-only

# Run pre-flight checks only
make preflight

# Install Neovim + config (server tooling)
make setup
```

Or with tags directly:
```bash
ansible-playbook tibia-char-playbook.yml -i inventory.yml \
  --key-file ~/.ssh/ansible \
  --ask-become-pass --vault-password-file ansible_vault_password.txt \
  --tags "deploy,health"
```

## Variables

Configuration lives in `group_vars/all/` (auto-loaded by Ansible):

| File | Contents |
|------|----------|
| `vars.yml` | Non-sensitive vars (`server_port`, `api_protocol`, `api_url`) |
| `vault.yml` | Secrets encrypted with `ansible-vault` |

### Vault

To encrypt or re-encrypt the vault file:
```bash
ansible-vault encrypt group_vars/all/vault.yml
```

## Architecture

| Role | Runs On | Purpose |
|------|---------|---------|
| `pre-flight` | Remote | Connectivity, disk space, DB reachability |
| `zip` | Localhost | Create deployment artifact |
| `docker-install` | Remote | Install Docker Engine + plugins |
| `build` | Remote | Extract artifact, backup old deployment |
| `env` | Remote | Generate `.env` (mode 0600) |
| `run` | Remote | Docker compose down/up |
| `migrations` | Remote | Run PostgreSQL migrations (idempotent) |
| `health-check` | Remote | Verify containers + app health |
| `cleanup` | Remote | Remove temp files |
| `bootstrap` | Both | Generate SSH key + deploy to server |
| `setup-nvim` | Remote | Install Neovim + write `init.lua` config |

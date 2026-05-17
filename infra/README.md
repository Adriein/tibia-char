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
```

Or with tags directly:
```bash
ansible-playbook tibia-char-playbook.yml -i inventory.yml \
  -e @ansible_vault.enc --key-file ~/.ssh/ansible \
  --ask-become-pass --vault-password-file ansible_vault_password.txt \
  --tags "deploy,health"
```

## Vault

To encrypt or re-encrypt the vault file:
```bash
ansible-vault encrypt ansible_vault.enc
```

## Architecture

| Role | Runs On | Purpose |
|------|---------|---------|
| `pre-flight` | Remote | Connectivity, disk space, DB reachability |
| `tibia-char-zip` | Localhost | Create deployment artifact |
| `docker-install` | Remote | Install Docker Engine + plugins |
| `tibia-char-build` | Remote | Extract artifact, backup old deployment |
| `tibia-char-set-env-vars` | Remote | Generate `.env` (mode 0600) |
| `tibia-char-run` | Remote | Docker compose down/up |
| `tibia-char-run-migrations` | Remote | Run PostgreSQL migrations (idempotent) |
| `health-check` | Remote | Verify containers + app health |
| `cleanup` | Remote | Remove temp files |
| `bootstrap` | Both | Generate SSH key + deploy to server |

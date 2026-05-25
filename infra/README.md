# Tibia Char - Ansible Deployment

Automated deployment of **tibia-char**.

## Prerequisites

1. Create `ansible_vault_password.txt` (gitignored) with your vault password.
2. Create a ssh key and copy the .pub to the remote
```bash
ssh-keygen -t ed25519 -f ~/.ssh/ansible -C "ansible"
ssh-copy-id -i ~/.ssh/ansible.pub aclaret@192.168.1.56
```
3. Update sudoers removing the need for a password to become sudo
```bash
sudo visudo -f /etc/sudoers.d/ansible
aclaret ALL=(ALL) NOPASSWD:ALL
```

## Usage

```bash
# First-time setup on a fresh server
make bootstrap

# Full deployment to stg
make deploy-stg

# Build artifact only (localhost)
make build-only
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
To decrypt the actual vault file:
```bash
ansible-vault decrypt group_vars/all/vault.yml
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

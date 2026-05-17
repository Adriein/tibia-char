# Makefile for executing Ansible playbook

PLAYBOOK = tibia-char-playbook.yml
INVENTORY = inventory.yml
KEY_FILE = ~/.ssh/ansible
ANSIBLE_VAULT_ENC = @ansible_vault.enc

bootstrap:
	ansible-playbook $(PLAYBOOK) -i $(INVENTORY) --key-file $(KEY_FILE) --ask-pass --ask-become-pass --tags bootstrap

run:
	ansible-playbook $(PLAYBOOK) -i $(INVENTORY) -e $(ANSIBLE_VAULT_ENC) --key-file $(KEY_FILE) --ask-become-pass --vault-password-file ansible_vault_password.txt

build-only:
	ansible-playbook $(PLAYBOOK) -i $(INVENTORY) -e $(ANSIBLE_VAULT_ENC) --key-file $(KEY_FILE) --ask-become-pass --vault-password-file ansible_vault_password.txt --tags build

config-only:
	ansible-playbook $(PLAYBOOK) -i $(INVENTORY) -e $(ANSIBLE_VAULT_ENC) --key-file $(KEY_FILE) --ask-become-pass --vault-password-file ansible_vault_password.txt --tags config

deploy-no-build:
	ansible-playbook $(PLAYBOOK) -i $(INVENTORY) -e $(ANSIBLE_VAULT_ENC) --key-file $(KEY_FILE) --ask-become-pass --vault-password-file ansible_vault_password.txt --tags deploy

preflight:
	ansible-playbook $(PLAYBOOK) -i $(INVENTORY) --key-file $(KEY_FILE) --ask-become-pass --vault-password-file ansible_vault_password.txt --tags always

.PHONY: bootstrap run build-only config-only deploy-no-build preflight clean

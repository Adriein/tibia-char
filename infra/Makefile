# Makefile for executing Ansible playbook

CD_PLAYBOOK = deploy-playbook.yml
CI_PLAYBOOK = build-playbook.yml
INVENTORY = inventory.yml
KEY_FILE = ~/.ssh/ansible

deploy:
	ansible-playbook $(CD_PLAYBOOK) -i $(INVENTORY) --key-file $(KEY_FILE) --ask-become-pass --vault-password-file ansible_vault_password.txt

build:
	ansible-playbook ${CI_PLAYBOOK} --vault-password-file ansible_vault_password.txt

.PHONY: build deploy

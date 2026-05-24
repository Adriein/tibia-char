# Makefile for executing Ansible playbook

PLAYBOOK = deploy-playbook.yml
INVENTORY = inventory.yml
KEY_FILE = ~/.ssh/ansible

bootstrap:
	ansible-playbook $(PLAYBOOK) -i $(INVENTORY) -t bootstrap --key-file $(KEY_FILE) --ask-become-pass --vault-password-file ansible_vault_password.txt

deploy-stg:
	ansible-playbook $(PLAYBOOK) -i $(INVENTORY) -t full --key-file $(KEY_FILE) --ask-become-pass --vault-password-file ansible_vault_password.txt

build:
	ansible-playbook ${PLAYBOOK} -t build --vault-password-file ansible_vault_password.txt

.PHONY: build deploy bootstrap

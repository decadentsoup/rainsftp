#!/bin/bash -e
# Command-line script to bcrypt a password using Apache HTTPd tooling
# This is intended to be used with JSON_USERS auth
# Bash is required for the -s option to hide the password
read -sp 'Password: ' password
htpasswd -bnBC 10 "" "$password" | tr -d :

ensure_vault_login() {
  export VAULT_TOKEN=$(vault print token)

  if ! vault login "$VAULT_TOKEN" > /dev/null 2>&1; then
    echo "Enter your LDAP password:"
    if ! vault login -method=ldap 2> /dev/null; then
      echo "LDAP login failed. Please check your credentials or connection."
    fi
  fi
}

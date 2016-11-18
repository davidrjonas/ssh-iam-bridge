#!/bin/bash -e

USER=ssh-iam-bridge
CMD=/usr/local/bin/ssh-iam-bridge
AUTH_KEYS_CMD=/usr/sbin/ssh-iam-bridge-public-keys

# ssh is picky about AuthorizedKeysCommand, see man sshd_config
cat <<EOF > $AUTH_KEYS_CMD
#!/bin/sh
exec $CMD authorized_keys "$@"
EOF
chown root:root $AUTH_KEYS_CMD
chmod 0755 $AUTH_KEYS_CMD

# Create a special user for running the AuthorizedKeysCommand
useradd --comment "ssh-iam-bridge authorized keys lookup" --shell /usr/sbin/nologin "$USER"

# Comment out existing directives
cat <<EOF | sed -i -f - /etc/ssh/sshd_config
s/^AuthorizedKeysCommand /#AuthorizedKeysCommand /
s/^AuthorizedKeysCommandUser /#AuthorizedKeysCommandUser /
s/^ChallengeResponseAuthentication /#ChallengeResponseAuthentication /
s/^AuthenticationMethods /#AuthenticationMethods /
EOF

cat <<EOF >> /etc/ssh/sshd_config
AuthorizedKeysCommand $AUTH_KEYS_CMD
AuthorizedKeysCommandUser $USER
ChallengeResponseAuthentication yes
AuthenticationMethods publickey keyboard-interactive:pam,publickey
EOF

# Verify that sshd_config is still valid
sshd -t

PAM_EXEC="auth requisite pam_exec.so stdout $CMD pam_create_user"

grep -q -F "$PAM_EXEC" /etc/pam.d/sshd || sed -i "1 a$PAM_EXEC" /etc/pam.d/sshd

# Ensure groups are updated
echo "*/10 * * * * root /usr/bin/ssh-iam-bridge sync_groups" > /etc/cron.d/ssh-iam-bridge

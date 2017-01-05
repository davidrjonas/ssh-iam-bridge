AWS IAM/SSH Bridge
==================

[![Build Status](https://travis-ci.org/davidrjonas/ssh-iam-bridge.svg)](https://travis-ci.org/davidrjonas/ssh-iam-bridge)
[![Go Report Card](https://goreportcard.com/badge/github.com/davidrjonas/ssh-iam-bridge)](https://goreportcard.com/report/github.com/davidrjonas/ssh-iam-bridge)

ssh-iam-bridge lets you use the SSH public keys stored in AWS IAM to
authenticate users on linux hosts.

Inspired by and nearly a direct port of
[Keymaker](https://github.com/kislyuk/keymaker) from Python to Go.

Theory of Operation
-------------------

When a client connects to a host via SSH the sshd daemon may look to an
external command to find the list of authorized keys for that user. Those keys
can be pulled from IAM on demand. Assuming we trust IAM, at that point the user
is considered "known" and good.
[Pam](https://en.wikipedia.org/wiki/Pluggable_authentication_module) can be
configured to trust ssh and add the user to the system. The local system groups
are synchronized from the IAM groups by looking for ones with a given prefix.
This allows group management to be done from IAM alone.

Resources
---------

- http://man.openbsd.org/cgi-bin/man.cgi/OpenBSD-current/man5/sshd_config.5
- http://www.linux-pam.org/Linux-PAM-html/sag-pam_exec.html
- https://github.com/hashicorp/vault-ssh-helper

Usage
-----

Create groups in AWS IAM with the prefix "system-" and "system-&lt;role&gt;-". These
groups will be created on your servers. For instance, the IAM group
"system-wheel" will be created as the "wheel" group on the system.

When launching EC2 instances give them an IAM Role (instance profile) that
includes read access to IAM. There is a predefined policy named
`IAMReadOnlyAccess` that works well. Or, since this program uses the official
[AWS SDK](https://aws.amazon.com/sdk-for-go/), it will search out credentials
in the [usual places](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-config-files).

Run `ssh-iam-bridge install` on your linux host. This does a few things: create
a script for sshd AuthorizedKeysCommand to run, create a user under which the
script is run, modify sshd_config to run the script, modify pam to create the
iam user locally during ssh, and install a cronjob to synchronize the groups.


```
usage: ssh-iam-bridge [<flags>] <command> [<args> ...]

Flags:
  --help     Show context-sensitive help (also try --help-long and --help-man).
  --version  Show application version.

Commands:
  help [<command>...]
    Show help.

  install [<flags>] [<user>]
    Install this program to authenticate SSH connections and create users

    Flags:
      --no-pam   Don't install to PAM (no autocreate user on login, create users on sync)

  authorized_keys <user>
    Get the authorized_keys from IAM for user

  sync
    Sync the IAM users and groups with the local system

  sync_groups
    Sync only the IAM groups with the local system groups

  pam_create_user
    Create a user from the env during the sshd pam phase
```

Warranty
--------

I'm not a security expert and I don't program in Go very often. Use at your own
risk. Pull requests will be received with immense gratitude.

TODO
----

- [ ] Sanitize usernames. IAM is more permissive than linux. (use the ARN in
  the comment to get iam user)
- [ ] Test with 2FA also enabled (like Duo Security or
  libpam-google-authenticator)

Similar Projects
----------------
- https://github.com/kislyuk/keymaker
- https://github.com/widdix/aws-ec2-ssh


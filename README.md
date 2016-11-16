AWS IAM/SSH Bridge Thing
========================

That's not a good name.

Resources
---------
- https://github.com/kislyuk/keymaker/blob/master/keymaker/__init__.py
- https://golang.org/pkg/os/user/
- https://github.com/aws/aws-sdk-go/blob/master/service/iam/iamiface/interface.go

Tasks
-----
- install itself
- get authorized_keys
- ensure an iam user exists
- generate uid and gids consistently
- create users
- sync groups
  - create groups, add/remove users from groups

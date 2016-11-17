AWS IAM/SSH Bridge
==================

Inspired by and nearly a direct port of https://github.com/kislyuk/keymaker from Python to Go.

Resources
---------
- https://github.com/kislyuk/keymaker
- https://golang.org/pkg/os/user/
- https://github.com/aws/aws-sdk-go/blob/master/service/iam/iamiface/interface.go

Tasks
-----
- [ ] install itself
- [X] get authorized_keys
- [ ] ensure an iam user exists
- [X] generate uid and gids consistently
- [ ] create users
- [X] sync groups
  - [X] create groups, add/remove users from groups

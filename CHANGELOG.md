Change Log
==========

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

## [1.0.4] - 2017-01-05
### Fixed
- Properly match group prefixes

### Added
- Tests for strarray
- `sync` command
- `--no-pam` flag to `install` command

### Changed
- AWS SDK update
- strarray.Filter() and Unique() to return empty slice

## [1.0.3] - 2016-11-30
### Fixed
- pam_exec.so line fix

### Changed
- Minor messages changes

## [1.0.2] - 2016-11-30
### Fixed
- Fix strarray.ReadFile() return

### Added
- Add get-latest.sh

## [1.0.1] - 2016-11-21
### Changed
- Changed PATH in pam to match `getconf PATH`
- Replaced panic()s with better error messages
- Added this change log

## 1.0.0 - 2016-11-21
- Initial release with everything "working"

[Unreleased]: https://github.com/davidrjonas/ssh-iam-bridge/compare/1.0.4...HEAD
[1.0.4]: https://github.com/davidrjonas/ssh-iam-bridge/compare/1.0.3...1.0.4
[1.0.3]: https://github.com/davidrjonas/ssh-iam-bridge/compare/1.0.2...1.0.3
[1.0.2]: https://github.com/davidrjonas/ssh-iam-bridge/compare/1.0.1...1.0.2
[1.0.1]: https://github.com/davidrjonas/ssh-iam-bridge/compare/1.0.0...1.0.1


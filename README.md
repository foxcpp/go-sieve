go-sieve
====================

Sieve email filtering language ([RFC 5228]) interpreter implementation in Go.

## Features

* Binary representation for faster load/execute cycles (`Script.Save`, `sieve.RestoreSaved`).
* Integration tests harness for MTA integration testing (see `tests/execute.go`).

## Supported extensions

- envelope ([RFC 5228])
- fileinto ([RFC 5228])
- encoded-character ([RFC 5228])
- imap4flags ([RFC 5232])
- variables ([RFC 5229])
- relational ([RFC 5231])
- copy ([RFC 3894])
- reject/ereject ([RFC 5429])

## Example

See ./cmd/sieve-run.

## Known issues

- Some invalid scripts are accepted as valid (see tests/compile_test.go)
- Comments in addresses are not ignored when testing equality, etc.
- Source routes in addresses are not ignored when testing equality, etc.

[RFC 5228]: https://datatracker.ietf.org/doc/html/rfc5228
[RFC 5229]: https://datatracker.ietf.org/doc/html/rfc5229
[RFC 5232]: https://datatracker.ietf.org/doc/html/rfc5232
[RFC 5231]: https://datatracker.ietf.org/doc/html/rfc5231
[RFC 3894]: https://datatracker.ietf.org/doc/html/rfc3894
[RFC 5429]: https://datatracker.ietf.org/doc/html/rfc5429

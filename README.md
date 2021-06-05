Ayd? FTP Probe
==============

FTP and FTPS check plugin for [Ayd?](https://github.com/macrat/ayd) status monitoring service.


## Install

1. Download binary from [release page](https://github.com/macrat/ayd-ftp-probe/releases).

2. Save downloaded binary as `ayd-ftp-probe` or `ayd-ftps-probe` to somewhere directory that registered to PATH.


## Usage

``` shell
# test only connection
$ ayd ftp://example.com

# test connection and login
$ ayd ftp://user:pass@example.com
```

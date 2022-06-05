# envcd

environment configurations detector/discovery/dictionary

[![license card](https://img.shields.io/badge/License-Apache%202.0-brightgreen.svg?label=license)](https://github.com/openingo/envcd/blob/main/LICENSE)
[![go version](https://img.shields.io/github/go-mod/go-version/openingo/envcd)](#)
[![go report](https://goreportcard.com/badge/github.com/openingo/envcd)](https://goreportcard.com/report/github.com/openingo/envcd)
[![codecov report](https://codecov.io/gh/openingo/envcd/branch/main/graph/badge.svg)](https://codecov.io/gh/openingo/envcd)
[![workflow](https://github.com/openingo/envcd/actions/workflows/go.yml/badge.svg?event=push)](#)
[![lasted release](https://img.shields.io/github/v/release/openingo/envcd?label=lasted)](https://github.com/openingo/envcd/releases)

![Envcd Architecture](envcd.png)

## features
- user & data permission
  - user & admin console
  - application data sync permission
- openapi support
- data version control
- sync mode
  - sync to env
  - sync to application
  - based on redis or etcd
- multi store
  - mysql
  - redis
  - etcd
  - ...
- sidecar
- 
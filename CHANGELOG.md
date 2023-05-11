# Changelog

All notable changes to this project will be documented in this file.

## Table of Contents

- [2.0.2](#202)
- [2.0.1](#201)
- [2.0.0](#200)
- [1.0.0](#100)

---

## `2.0.2`

- integrate e2e tests with network policies
- fix a bug in udp testing

## `2.0.1`

- fix release naming

## `2.0.0`

- complete rewrite of the tool in Go, with unit and integration tests
- leverages the ephemeral container support in Kubernetes > v1.25
- test case(s) are written in YAML
- support for Pods, StatefulSets, DaemonSets and Deployments which are directly referred through their names in the test suites
- artifacts are available for download

## `1.0.0`

- initial release
- no artifacts available
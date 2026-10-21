# Changelog

## [0.1.3](https://github.com/SpechtLabs/gibson/compare/v0.1.2...v0.1.3) (2026-10-21)


### Features

* an 8-bit chiptune soundtrack, streamed over SSH and played on your speakers ([#19](https://github.com/SpechtLabs/gibson/issues/19)) ([ac9a581](https://github.com/SpechtLabs/gibson/commit/ac9a581b02bfddbde5e4fb0ca07441bca7ad8b88))


### Bug Fixes

* **deps:** update gcr.io/distroless/static-debian13:nonroot docker digest to 5167571 ([#17](https://github.com/SpechtLabs/gibson/issues/17)) ([7fcc5c5](https://github.com/SpechtLabs/gibson/commit/7fcc5c5fe2db98f2b9624d4219c0e705ae0be4ad))
* **deps:** update go toolchain directive to v1.27.3 ([#23](https://github.com/SpechtLabs/gibson/issues/23)) ([d29e5ac](https://github.com/SpechtLabs/gibson/commit/d29e5ac5d353d12e8f190cc8b325dc23cb7f6058))

## [0.1.2](https://github.com/SpechtLabs/gibson/compare/v0.1.1...v0.1.2) (2026-10-16)


### Bug Fixes

* make /etc/gibson in the image traversable by the nonroot user ([#9](https://github.com/SpechtLabs/gibson/issues/9)) ([4f201b8](https://github.com/SpechtLabs/gibson/commit/4f201b898226320f666fc0c7b383981e16ad1857))

## [0.1.1](https://github.com/SpechtLabs/gibson/compare/v0.1.0...v0.1.1) (2026-10-16)


### Features

* run behind a PROXY-protocol load balancer ([#7](https://github.com/SpechtLabs/gibson/issues/7)) ([157a6be](https://github.com/SpechtLabs/gibson/commit/157a6bec6e2cd6c1fb2580913451942f1fe33ed7))

## 0.1.0 (2026-10-16)


### Features

* gibson, a retro neon SpechtLabs intro served over SSH ([73397ac](https://github.com/SpechtLabs/gibson/commit/73397ac877c5a30f954e7cd3bac746835d51e979))
* more lines for turned-away viewers, from movies and the industry ([0f38515](https://github.com/SpechtLabs/gibson/commit/0f385151a729820db4c55a2c3a561ca7d52482ce))
* turn viewers away at capacity with a Hackers-style ACCESS DENIED ([94a631c](https://github.com/SpechtLabs/gibson/commit/94a631cffa8164f65e305b4a82b73424697b63dc))


### Bug Fixes

* log the loaded config file only when serving ([a7d7052](https://github.com/SpechtLabs/gibson/commit/a7d7052cfe627e15b7e7dab12aecc784bd8ee0a5))

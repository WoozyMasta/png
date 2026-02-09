# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog][],
and this project adheres to [Semantic Versioning][].

<!--
## Unreleased

### Added
### Changed
### Removed
-->

## [1.0.0][] - 2026-02-10

### Added

* Encoder: `HuffmanOnly` compression level and numeric zlib levels 1-9.
* Encoder: `BufferPool` and `BufferSize` options.
* Decoder type with optional `BufferPool` for row buffers;
  `Decoder.Decode()` method.

### Changed

* `compress/zlib` replaced with `klauspost/compress/zlib`
  (faster encode/decode, fewer allocs).
* Decoder: Paeth filter fast paths for 1/3/4 bytes per pixel.
* Struct field order tuned with betteralign (reduced padding, smaller structs).
* Style: some conditionals rewritten as switch; empty branches removed (linter)

[1.0.0]: https://github.com/WoozyMasta/png/tree/v1.0.0

<!--links-->
[Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html

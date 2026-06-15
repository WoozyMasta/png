<!-- markdownlint-disable MD024 -->
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

## Unreleased

### Added

* `purego` build tag compile-time
  and `PNG_PUREGO=1` runtime environment variable
  to force the pure-Go paths (no assembly);

### Changed

* Encoder: concrete-type fast paths for 16-bit images
  (`Gray16`/`RGBA64`/`NRGBA64`) remove ~one allocation per pixel
  (307k -> 25 allocs/op on a 640x480 image) and cut encode memory 35-68%.
* Decoder: 16-bit truecolor / grayscale-alpha and non-transparent `Gray16`
  rows are copied directly into the pixel buffer.
* `writeChunk` reuses `crc32.IEEETable` instead of allocating a hash per chunk.
* SIMD (amd64, avo-generated): SSE2 Up-filter add/SAD and SSSE3 `PSHUFB`
  RGB<->RGBA conversion, with a pure-Go fallback elsewhere;
  geomean -30% time vs 1.0.0.

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

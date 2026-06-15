# png

Drop-in accelerated PNG encoder/decoder, API-compatible with `image/png`.
Same format, fewer allocations, faster.

## What's faster

### Zlib

`compress/zlib` replaced with [klauspost/compress/zlib]. Same stream format.

* Time: geomean -33%; decode -7...-21% encode -42...-70%
  (e.g. DecodeRGB -21%, EncodeGray -59%, EncodeGrayWithBufferPool -70%).
* Allocs: geomean -19% (e.g. DecodeGray 102->69, DecodeRGB 157->110).

### Paeth filter

Paeth filter fast paths for 1/3/4 bytes per pixel
(single-pass, no strided inner loop).

* Time: DecodeGray -4%, DecodeRGB -3.5%, DecodeInterlacing -3.7%.

### 16-bit color types

Encoder fast paths for `Gray16`/`RGBA64`/`NRGBA64` use the concrete pixel
types instead of the per-pixel `m.At()`/`Convert` path (one alloc per pixel);
decode copies 16-bit rows straight into the pixel buffer.

* Time: encode -40...-65%, decode -15...-61%.
* Allocs: 16-bit encode 307225->25; encode B/op -35...-68%.

### SIMD (amd64)

avo-generated kernels, with a pure-Go fallback on other architectures
and under the `purego` build tag.

* Up filter (SSE2): add for decode reconstruction,
  sum-of-absolute-differences for the encoder filter heuristic.
* RGB<->RGBA (SSSE3 `PSHUFB`): 8-bit truecolor conversion, decode and encode.
* Time: EncodeRGBOpaque -43%, EncodeNRGBA -37%, EncodeRGBA -28%.

## Encoder and decoder options

### Encoder

* CompressionLevel: added `HuffmanOnly`
  (Huffman-only, no LZ; fastest encode, larger files)
  and numeric zlib level 1-9 (1=fast, 9=best). Rest as in standard library.
* BufferPool (`EncoderBufferPool`): reuse encoder internal buffers
  across multiple `Encode` calls.
  Cuts allocations when encoding many images in a row.
* BufferSize: size in bytes of the `bufio.Writer` used when writing IDAT chunks
  (default 32KB). Lets you tune for very large images or higher throughput;
  zero means default.

### Decoder

* Decoder + BufferPool (`DecoderBufferPool`): optional pool for row buffers
  (current/previous line) used in `readImagePass`.
  Reuse buffers when decoding many PNGs in sequence;
  no benefit for a single image.

[klauspost/compress/zlib]: https://github.com/klauspost/compress

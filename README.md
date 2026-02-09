# png

Drop-in accelerated PNG encoder/decoder, API-compatible with `image/png`.
Same format, fewer allocations, faster.

## What's faster

### Zlib

`compress/zlib` replaced with [klauspost/compress/zlib]. Same stream format.

* Time: geomean -33%; decode -7...-21% encode -42...-70%
  (e.g. DecodeRGB -21%, EncodeGray -59%, EncodeGrayWithBufferPool -70%).
* Allocs: geomean -19% (e.g. DecodeGray 102->69, DecodeRGB 157->110).

### Decoder

Paeth filter fast paths for 1/3/4 bytes per pixel
(single-pass, no strided inner loop).

* Time: DecodeGray -4%, DecodeRGB -3.5%, DecodeInterlacing -3.7%.

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

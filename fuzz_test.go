// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package png

import (
	"bytes"
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func FuzzDecode(f *testing.F) {
	if testing.Short() {
		f.Skip("Skipping in short mode")
	}

	testdata, err := os.ReadDir("testdata/common")
	if err != nil {
		f.Fatalf("failed to read testdata directory: %s", err)
	}
	for _, de := range testdata {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".png") {
			continue
		}
		b, err := os.ReadFile(filepath.Join("testdata/common", de.Name()))
		if err != nil {
			f.Fatalf("failed to read testdata: %s", err)
		}
		f.Add(b)
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		cfg, _, err := image.DecodeConfig(bytes.NewReader(b))
		if err != nil {
			return
		}
		if cfg.Width*cfg.Height > 1e6 {
			return
		}
		img, typ, err := image.Decode(bytes.NewReader(b))
		if err != nil || typ != "png" {
			return
		}
		// Canonical pixels of the source, so the roundtrip is checked against
		// the actual image data, not merely its bounds.
		wantPix := toNRGBA64(img).Pix

		levels := []CompressionLevel{
			DefaultCompression,
			NoCompression,
			BestSpeed,
			BestCompression,
			HuffmanOnly,
			1, // numeric zlib level (fastest)
			9, // numeric zlib level (best)
		}
		for _, l := range levels {
			var w bytes.Buffer
			e := &Encoder{CompressionLevel: l}
			err = e.Encode(&w, img)
			if err != nil {
				t.Errorf("level %d: failed to encode valid image: %s", l, err)
				continue
			}
			img1, err := Decode(&w)
			if err != nil {
				t.Errorf("level %d: failed to decode roundtripped image: %s", l, err)
				continue
			}
			if got, want := img1.Bounds(), img.Bounds(); got != want {
				t.Errorf("level %d: roundtripped image bounds have changed, got: %s, want: %s", l, got, want)
				continue
			}
			if !bytes.Equal(toNRGBA64(img1).Pix, wantPix) {
				t.Errorf("level %d: roundtripped image pixels have changed", l)
			}
		}
	})
}

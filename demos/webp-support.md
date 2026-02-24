# WebP Image Support

*2026-02-24T15:35:09Z by Showboat dev*
<!-- showboat-id: 1727d21c-a77f-408a-9252-8865de2200bf -->

This walkthrough demonstrates that showboat now supports WebP images. Previously, `showboat image` rejected .webp files with 'unrecognized image format: .webp'.

Create a valid 1x1 WebP image from raw bytes (no external tools needed).

```bash
python3 -c "
import sys
data = bytes([
    0x52,0x49,0x46,0x46,0x36,0x00,0x00,0x00,0x57,0x45,0x42,0x50,0x56,0x50,0x38,0x20,
    0x2a,0x00,0x00,0x00,0xf0,0x01,0x00,0x9d,0x01,0x2a,0x01,0x00,0x01,0x00,0x02,0x00,
    0x34,0x25,0xa0,0x02,0x74,0xba,0x01,0xf8,0x00,0x05,0xf4,0x00,0x00,0x9b,0xff,0xcb,
    0x3d,0xe6,0x37,0x7b,0xa6,0xff,0xe2,0xce,0xee,0x96,0x75,0xd0,0x00,0x00
])
sys.stdout.buffer.write(data)
" > /tmp/test-webp.webp && file /tmp/test-webp.webp
```

```output
/tmp/test-webp.webp: RIFF (little-endian) data, Web/P image, VP8 encoding, 1x1, Scaling: [none]x[none], YUV color, decoders should clamp
```


```bash
go test ./exec/ -run TestCopyImageWebP -v
```

```output
=== RUN   TestCopyImageWebP
--- PASS: TestCopyImageWebP (0.00s)
PASS
ok  	github.com/simonw/showboat/exec	(cached)
```

All passing. The fix adds `.webp` to the `validImageExts` map in `exec/image.go`.

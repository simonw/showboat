# Dev Server Startup

*2026-02-14T18:23:19Z by Showboat dev*

This demo shows the new `{server}` block feature. Server blocks let documents include long-running server processes that subsequent code blocks can interact with. A free port is auto-assigned and made available as `$PORT`.

## Setting up

Create a directory with a simple web page for our server to serve:

```bash
rm -rf /tmp/server-demo && mkdir -p /tmp/server-demo && echo "<h1>Hello from showboat</h1>" > /tmp/server-demo/index.html && cat /tmp/server-demo/index.html
```

```output
<h1>Hello from showboat</h1>
```

Now create a demo document with a `{server}` block using Python to avoid markdown escaping issues:

```python3

import pathlib
bt = chr(96) * 3
doc = f'''# Server Demo

*2026-02-14T00:00:00Z*

{bt}bash {{server}}
cd /tmp/server-demo && python3 -m http.server $PORT
{bt}

{bt}bash
curl -s http://localhost:$PORT/index.html
{bt}

{bt}output
<h1>Hello from showboat</h1>
{bt}
'''
pathlib.Path('/tmp/server-demo/demo.md').write_text(doc)
print('Created /tmp/server-demo/demo.md')

```

```output
Created /tmp/server-demo/demo.md
```

## Verifying with server blocks

`showboat verify` automatically starts the server, assigns a free port via `$PORT`, and runs subsequent code blocks against it:

```bash
/tmp/showboat verify /tmp/server-demo/demo.md && echo "Verification passed\!"
```

```output
Serving HTTP on 0.0.0.0 port 36559 (http://0.0.0.0:36559/) ...
127.0.0.1 - - [14/Feb/2026 18:27:49] "GET /index.html HTTP/1.1" 200 -
Verification passed\!
```

## Extract emits server commands

`showboat extract` correctly emits `showboat server` for `{server}` blocks:

```bash
/tmp/showboat extract /tmp/server-demo/demo.md
```

```output
showboat init /tmp/server-demo/demo.md 'Server Demo'
showboat server /tmp/server-demo/demo.md bash 'cd /tmp/server-demo && python3 -m http.server $PORT'
showboat exec /tmp/server-demo/demo.md bash 'curl -s http://localhost:$PORT/index.html'
```

## Drift detection still works

If we change the served content, verify detects the drift:

```bash
echo "<h1>Changed content</h1>" > /tmp/server-demo/index.html
/tmp/showboat verify /tmp/server-demo/demo.md
echo "Exit code: $?"
```

```output
Serving HTTP on 0.0.0.0 port 63336 (http://0.0.0.0:63336/) ...
127.0.0.1 - - [14/Feb/2026 18:27:50] "GET /index.html HTTP/1.1" 200 -
block 2:
  expected: <h1>Hello from showboat</h1>
  actual:   <h1>Changed content</h1>
Exit code: 1
```

The server was started, the curl ran against it, and verify correctly detected that the output changed from the expected `<h1>Hello from showboat</h1>` to `<h1>Changed content</h1>`.

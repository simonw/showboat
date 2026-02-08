# Shell Output and Pop

*2026-02-08T13:52:45Z*

This demo shows two new showboat features: build run now prints shell output to stdout and reflects the exit code, and the new pop command removes the most recent entry from a document.

## Build run output

When running a shell command, the output is now printed to stdout so the agent can see the result:

```bash
echo 'Hello from the shell\!'
```

```output
Hello from the shell\!
```

The exit code from the shell is reflected. A successful command exits 0:

```bash
echo 'This succeeds' && python3 -c "print(2 + 2)"
```

```output
This succeeds
4
```

## The pop command

The pop command removes the most recent entry from a document. Let's say an agent runs a command that produces an error:

```bash
ls /tmp | head -5
```

```output
145.0.7632.46
claude-0
claude-code-318714088.diag.log
claude-code.log
claude-command
```

The failed entry was removed and replaced with the correct one. The final document only shows the successful attempt.

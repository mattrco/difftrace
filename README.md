difftrace
=========

`difftrace` is a small tool for manipulating `strace` output to produce more meaningful diffs.

Normally when you run strace, memory addresses, timestamps and other things will change between runs. `difftrace` replaces some of these things with placeholders. Here's an example:

```
cat strace_run.out | difftrace
```

Currently `difftrace` doesn't handle all possible outputs, but it does handle the simple cases. A lexer and parser have been implemented so that extending what it does is easier.

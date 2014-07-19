Synesthesia
===========

Synesthesia takes regular expressions and uses them to colorize stdin.
Moreover, every identical matches are colored with the same color, making it a
good tool to visualize large files or streaming data.

```
Usage: synesthesia ([flags]|[patterns])... [-- [patterns]...]
```

Where patterns are any valid Python regular expressions and flags are any of
the following flags. Multiple patterns are or'd together in the order they are
provided. Multiple flags controlling the same parameter may also be specified;
the last one specified for that parameter will be used.

```
Flags:
-b --background     Turn on or off background colorization. (default off)
   --no-background
-f --foreground     Turn on or off foreground colorization. (default on)
   --no-foreground
-h --help           Print usage information.
```

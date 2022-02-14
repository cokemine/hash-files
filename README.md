# hash-files

Self-used hash generator.

Support MD5, SHA1, SHA256, QuickXorHash

Options:

```
Usage
  -algo string
        The hash algorithm to use, multiple algorithms can be specified by comma separated (default "md5")
  -dir string
        The directory to be hashed (default "$(pwd)")
  -parallel int
        The number of parallel workers (default runtime.NumCPU())
  -verbose
        Verbose output log
```

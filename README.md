# hashfiles

Self-used hash generator.

Support MD5, SHA1, SHA256, QuickXorHash

Usage:

```
NAME:
   hashfiles - Recursively generate checksum of all files in a directory

USAGE:
   hashfiles [global options] command [command options] [arguments...]

COMMANDS:
   hash     Hash files
   verify   Verify files
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --dir value, -d value       The directory to be hashed (default: "$(pwd)")
   --algo value, -a value      The hash algorithm to use, multiple algorithms can be specified by comma separated (default: "md5")
   --parallel value, -n value  The number of parallel workers (default: 16)
   --verbose                   Verbose output log (default: false)
   --help, -h                  show help (default: false)
```

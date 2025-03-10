# Simple service wrapper to run an executable as windows service

```sh
# Remember there is an alias i pwsh for sc
sc.exe create <service_name> binPath= "<path_to_wrapper_exe> <path_to_wrapped_exe>"

sc.exe delete <service_name>
```
## Info

stdout and stderr will be written to a file under c:\logs the filename will contain the wrapped executable's name.

file size is limited to 50MB and 10 files are kept on disk.

This should be developed and built on Windows

Uses: Lumberjack, windows/svc

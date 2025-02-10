# Simple service wrapper to run an executable as windows service

```sh
sc create <service_name> binPath= "<path_to_wrapper_exe> <path_to_wrapped_exe>"
```
## Info

stdout and stderr will be written to a file under c:\logs the filename will contain the wrapped executable's name.

file size is limited to 50MB and 10 files are kept on disk.

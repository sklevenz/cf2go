```
# get sources
go get github.com/sklevenz/cf2go
# builds and installs binary in $GOPATH/bin directory
go install github.com/sklevenz/cf2go
```

```
usage: cf2go [<flags>] <command> [<args> ...]

A command line toolset to manage CF landscapes

Flags:
  --help     Show context-sensitive help (also try --help-long and --help-man).
  --url="https://raw.githubusercontent.com/sklevenz/cf2go/master/landscape.json"  
             URL from which the landscape configuration is requested
  --version  Show application version.

Commands:
  help [<command>...]
    Show help.

  list
    List landscapes configuration

  jump <landscape-id>
    SSH to a jumpbox system

```

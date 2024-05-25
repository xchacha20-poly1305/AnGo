# go-install-update

Update all of GOBIN.

AnGo provides an easy way to manage your GOBIN. You can install module with `-trimpath` and `-ldflags` flags, and automatically update them.

# Usage

Install by go:

```shell
go install -v -trimpath -ldflags "-w -s -buildid=" github.com/xchacha20-poly1305/ango/cmd/ango@latest
```

Build:

```shell
go build -v -trimpath -ldflags "-w -s -buildid=" ./cmd/ango/
```

command: 

```
Usage of ango:
  -V	Print version.
  -d	Dry run. Just check update.
  -ldflags string
    	 (default "-s -w")
  -r	Re-install all binaries.
  -trimpath
    	 (default true)
  -v	Show verbose info. And append -v flag to go install

```

# Credits

* [Gelio/go-global-update](https://github.com/Gelio/go-global-update)
* [rsc/goversion](https://github.com/rsc/goversion)
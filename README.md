# go-install-update

Update all of GOBIN.

# Usage

Install by go:

```shell
go install -v -trimpath -ldflags "-w -s -buildid=" github.com/xchacha20-poly1305/go-install-update/cmd/go-install-update@latest
```

Build:

```shell
go build -v -trimpath -ldflags "-w -s -buildid=" ./cmd/go-install-update/
```

command: 

```shell
Usage of go-install-update:
  -V	
  -d	Dry run. Just check update.
  -ldflags string
    	 (default "-s -w")
  -r	Re-install all binaries.
  -trimpath
    	 (default true)
  -v	go install -v

```

# Credits

* [Gelio/go-global-update](https://github.com/Gelio/go-global-update)
* [rsc/goversion](https://github.com/rsc/goversion)
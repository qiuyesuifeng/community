## community

community is a tool for collecting github informations, including repos, stargazers, forkers.

## How to build

```
make deps && sh deps.sh (optional, install golang dependent packages)
make build
```

## How to use

```
Usage of community:
  -L string
      log level: debug, info, warn, error, fatal (default "info")
  -config string
      Config file
  -o string
      github owner
  -r string
      github repo
  -s string
      print service information [ repos | stargazers ]
  -t string
      github access token
```

## Example

```
./community -o pingcap -r tidb -t {your-token} -s repos
```

Or use config file.

```
./community --config=config.toml -s stargazers
```

## License
Apache 2.0 license. See the [LICENSE](./LICENSE) file for details.

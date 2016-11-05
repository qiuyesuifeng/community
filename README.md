## community

community is a tool for collecting github informations, including repos, stargazers, forkers, etc.

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
  -end string
      end date
  -i string
      input file
  -o string
      github owner
  -r string
      github repo
  -s string
      print service information [ repos | stargazers | stargazer-ids | users ]
  -start string
      start date
  -t string
      github access token
```

## Example

```
./community -o pingcap -r tidb -t {your-token} -s {service}
```

Or use config file.

```
./community --config=config.toml -s {service}
```

## Service


### repos
List github public repos of owner.

### stargazers
List github stargazers of {owner}/{repo}.

*you can choose start and end as starred time filter for stargazers collection, like `-start 2016-11-01 -end 2016-11-06`*

### stargazer-ids
List github stargazer ids of {owner}/{repo}.

*you can choose start and end as starred time filter for stargazers collection, like `-start 2016-11-01 -end 2016-11-06`*

### users
List github users.

*you can choose input file as specified user ids for users collection，like `-i input`*

## License
Apache 2.0 license. See the [LICENSE](./LICENSE) file for details.

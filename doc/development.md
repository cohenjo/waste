
# Development

## Pre-reqs
We need a local redis to keep the artifact to cluster binding - just because I hate chef.
`docker run -n some-redis -P redis`

## current flow
go run go/cmds/wastedirect/waste.go --config ./conf/waste.conf

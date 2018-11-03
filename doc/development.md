
# Development

## Pre-reqs
We need a local redis to keep the artifact to cluster binding - just because I hate chef.  
`docker run -n some-redis -P redis`

## current flow
create some configuration `./conf/waste.conf` - you can see the sample and just put correct stuff there.
run:  
`go run go/cmds/wastedirect/waste.go --config ./conf/waste.conf`

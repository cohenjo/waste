
# Development

## Pre-reqs
gh-ost interduced a change:
../pkg/mod/github.com/github/gh-ost@v1.0.48/go/base/context.go:718:2: undefined: gcfg.RelaxedParserMode
../pkg/mod/github.com/github/gh-ost@v1.0.48/go/base/context.go:719:2: undefined: scanner.RelaxedScannerMode

we need to drop that as we don't have that internal change... - we don't use config file anyway...

### Permissions on the DB
for dev as we reach from a different network please create a user, e.g.:
`CREATE USER 'dbschema'@'192.168.%' IDENTIFIED WITH 'mysql_native_password' AS '*password' REQUIRE NONE PASSWORD EXPIRE DEFAULT ACCOUNT UNLOCK;`
`GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, INDEX, ALTER, LOCK TABLES, REPLICATION SLAVE, REPLICATION CLIENT, TRIGGER ON *.* TO 'dbschema'@'192.168.%';`

## current flow
create some configuration `./conf/waste.conf` - you can see the sample and just put correct stuff there.
run:  
`go run go/cmds/wastedirect/waste.go --config ./conf/waste.conf`

after completion we need to delete the remote and local branch.
remote can be easily deleted via the pull request
the local can be cleared via:
`git pull -p`
or explicitly using:
`git branch -d the_local_branch`

check if this is useful:
https://github.com/GuiaBolso/darwin
https://github.com/gobuffalo/pop
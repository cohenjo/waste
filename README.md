# waste

[![build status](https://travis-ci.org/cohenjo/waste.svg)](https://travis-ci.org/cohenjo/waste) [![downloads](https://img.shields.io/github/downloads/cohenjo/waste/total.svg)](https://github.com/cohenjo/waste/releases) [![release](https://img.shields.io/github/release/cohenjo/waste.svg)](https://github.com/cohenjo/waste/releases)

#### What artifact schema transformations etc for MySQL <img src="doc/images/waste-logo-light-160.png" align="right">

 `waste` is a schema transformation tool for micro-services , it integrates systems like Chef, Orchestrator & gh-ost to match between artifact, his DB server and the gh-ost utility to make the actual changes (plus some changes done directly...)
 
`waste` produces a light workload on the master throughout the migration, decoupled from the existing workload on the migrated table.

It has been designed based on years of experience with existing solutions, taking the worst off each.


## How?

does it matter?
if you really want to know - just read the code.

Hint: it uses gh-ost for the heavy lifting - small wrapping and connections done by me.

## Why?

Because we can!

## Highlights

- get's changes off some github repo - approved Pull-requests and such...
- can do stuff
- most likely to fail only 47% of changes

Please refer to the [docs](doc) for more information. No, really, read the [docs](doc).

## Usage

The [cheatsheet](doc/cheatsheet.md) has it all. You may be interested in invoking `waste` in various modes:

- a _noop_ migration (merely testing that the migration is valid and good to go)
- a real migration, utilizing a replica (the migration runs on the master; `waste` figures out identities of servers involved. Required mode if your master uses Statement Based Replication)
- a real migration, run directly on the master (but `waste` prefers the former)
- a real migration on a replica (master untouched)
- a test migration on a replica, the way for you to build trust with `waste`'s operation.

Our tips:

- [Testing above all](doc/testing-on-replica.md), try out `--test-on-replica` first few times. Better yet, make it continuous. We have multiple replicas where we iterate our entire fleet of production tables, migrating them one by one, checksumming the results, verifying migration is good.
- For each master migration, first issue a _noop_
- Then issue the real thing via `--execute`.

More tips:

- Use `--exact-rowcount` for accurate progress indication
- Use `--postpone-cut-over-flag-file` to gain control over cut-over timing
- Get familiar with the [interactive commands](doc/interactive-commands.md)

gh-ost requires an account with these privileges:
  ALTER, CREATE, DELETE, DROP, INDEX, INSERT, LOCK TABLES, SELECT, TRIGGER, UPDATE on the migrated database 

Also see:

- [requirements and limitations](doc/requirements-and-limitations.md)
- [common questions](doc/questions.md)
- [what if?](doc/what-if.md)
- [the fine print](doc/the-fine-print.md)
- [Community questions](https://github.com/github/waste/issues?q=label%3Aquestion)
- [Using `waste` on AWS RDS](doc/rds.md)

## What's in a name?

A couple of whiskey shots, Vodka, 2 bottles of wine and a six pack.

## License

`waste` is licensed under the [MIT license](https://github.com/github/waste/blob/master/LICENSE)

`waste` uses 3rd party libraries, each with their own license. These are found [here](https://github.com/github/waste/tree/master/vendor).

## Community

`waste` is released at a stable state, but with mileage to go. We are [open to pull requests](https://github.com/github/waste/blob/master/.github/CONTRIBUTING.md). Please first discuss your intentions via [Issues](https://github.com/github/waste/issues).

I developed `waste` to learn Go-lang. honestly if you find use for this feel free to suggest - I'd be happy to make changes.
Please see [Coding waste](doc/coding-waste.md) for a guide to getting started developing with waste.

## Download/binaries/source

`waste` is never GA and never stable.

`waste` is a Go project; it is built with Go `1.8` (though `1.7` should work as well). To build on your own, use either:
- [script/build](https://github.com/github/waste/blob/master/script/build) - this is the same build script used by CI hence the authoritative; artifact is `./bin/waste` binary.
- [build.sh](https://github.com/github/waste/blob/master/build.sh) for building `tar.gz` artifacts in `/tmp/waste`

## Authors

`waste` is designed, authored, reviewed and tested by me:
- [@cohenjo](https://github.com/cohenjo)

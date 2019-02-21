package main

import (
	"github.com/cohenjo/waste/go/config"
	"github.com/cohenjo/waste/go/http"
	"github.com/outbrain/golib/log"
)

// add column i int
// add key owner_idx(owner_id)
// add unique key owner_name_idx(owner_id, name) - though you need to make sure to not write conflicting rows while this migration runs
// drop key name_uidx - primary key is shared between the tables
// drop primary key, add primary key(owner_id, loc_id) - name_uidx is shared between the tables and is used for migration
// change id bigint unsigned - the 'primary key is used. The change of type still makes the primary key workable.
// drop primary key, drop key name_uidx, create primary key(name), create unique key id_uidx(id) - swapping the two keys. gh-ost is still happy because id is still unique in both tables. So is name.

func main() {

	log.SetLevel(log.INFO)

	log.Infof("Hello, world.\n")
	clio := config.CLIOptions{}
	clio.ReadArgs()

	config.Config = clio
	http.Serve()

	// gh-ost \
	// --max-load=Threads_running=25 \
	// --critical-load=Threads_running=1000 \
	// --chunk-size=1000 \
	// --throttle-control-replicas="myreplica.1.com,myreplica.2.com" \
	// --max-lag-millis=1500 \
	// --user="gh-ost" \
	// --password="123456" \
	// --host=replica.with.rbr.com \
	// --database="my_schema" \
	// --table="my_table" \
	// --verbose \
	// --alter="engine=innodb" \
	// --switch-to-rbr \
	// --allow-master-master \
	// --cut-over=default \
	// --exact-rowcount \
	// --concurrent-rowcount \
	// --default-retries=120 \
	// --panic-flag-file=/tmp/ghost.panic.flag \
	// --postpone-cut-over-flag-file=/tmp/ghost.postpone.flag \
	// [--execute]

	log.Info("# Done\n")

}

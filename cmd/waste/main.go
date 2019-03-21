package main

import (
	"github.com/cohenjo/waste/go/config"
	"github.com/cohenjo/waste/go/http"
	"github.com/cohenjo/waste/go/logic"
	"github.com/cohenjo/waste/go/scheduler"
	"github.com/rs/zerolog/log"
)

//go:generate swagger generate spec

// add column i int
// add key owner_idx(owner_id)
// add unique key owner_name_idx(owner_id, name) - though you need to make sure to not write conflicting rows while this migration runs
// drop key name_uidx - primary key is shared between the tables
// drop primary key, add primary key(owner_id, loc_id) - name_uidx is shared between the tables and is used for migration
// change id bigint unsigned - the 'primary key is used. The change of type still makes the primary key workable.
// drop primary key, drop key name_uidx, create primary key(name), create unique key id_uidx(id) - swapping the two keys. gh-ost is still happy because id is still unique in both tables. So is name.

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {

	log.Info().Msgf("Hello, world.\n")
	config.Config = config.LoadConfiguration()
	scheduler.WS = scheduler.SetupScheduler()
	logic.CM = logic.SetupChangeManager()
	scheduler.WS.Start()
	http.Serve()
	log.Info().Msgf("# Done")

}

package scheduler

type scheduler struct {
}

// root@localhost(db-mysql-others-local0a.42):[information_schema]> select TABLE_SCHEMA,TABLE_NAME  from information_schema.TABLES where TABLE_NAME like '__waste_%';
// +--------------+----------------------------+
// | TABLE_SCHEMA | TABLE_NAME                 |
// +--------------+----------------------------+
// | greyhound_db | __waste_2019_2_28_jonyTest |
// +--------------+----------------------------+
// 1 row in set (0.00 sec)

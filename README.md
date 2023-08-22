# velaux-database-migrator

Database migration command-line tool for VelaUX.

## Introduction

velaux-database-migrator is a command line tool for migrating the database of [velaux](https://github.com/kubevela/velaux) across different database drivers.

It's an easy, fast and reliable way of database migration for velaux.

## Prerequisites

- Ensure the source and target database servers are running.

## Quickstart

### Installation

### Setup

For migrating the database you need a config file for the configuration of source and targte databases and other options.

example config file :
```yaml
source:
  URL: "<user>:<password>d@<host>/<database>"
  Type: "mysql"
  Database: "kubevela"

target:
  URL: "mongodb+srv://<username>:<password>@<host>"
  Type: "mongodb"
  Database: "kubevela"

actionOnDup: "update" // you can use "skip" also. And if nothing is provided then it will throw an error
tables: // this is for specifying the table names which are to be migrated. By default it will migrate all the tables
  - "vela_user"
  - "vela_application"
```

- In the config file the `actionOnDup` tag represents the action that will be taken in case of duplicate entry. By default it will throw error.
- `tables` represents the database tables that needs to be migrated. If no table is provided then it will migrate the whole database tables.

After setting up the config file just run - 

``` shell
velamg migrate -c ./config.yaml
```

There you go ! After the migration is successful all the database tables from source database is migrated to the target database according to the given config.

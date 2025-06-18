# Orca ZTBus Data Prep

This repository contains a utility executable to port the ZTBus dataset into a database.

Current supported databases are:

- PostgreSQL

## How to Use this Tool

This tool is focused on porting the ZTBus dataset to a database. Follow these instructions to do this:

### 1. Download the ZTBus dataset

The ZTBus dataset can be found here:
https://www.research-collection.ethz.ch/handle/20.500.11850/626723

Download the data and export the raw .csv files to a directory (including metaData.csv). E.g. `data/raw/`.

### 2. Build the Utility

The utility is written in go. Build it with the following command:

```go
go build -ldflags "-s -w"
```

### 3. Start a Database Instance

Start a database Instance. There is a `docker-compose.yml` file at the root of this repository that
will start a postgresql instance locally:

```bash
docker compose up -d
```

### 4. Start the Data Migration

Run the CLI tool to execute the migration:

```bash
./orca-ztbus-prep --connStr "postgresql://ztbus:ztbus@localhost:5437/ztbus?sslmode=disable" --platfo
rm postgresql --migrate --dataDir "./data/raw/"
```

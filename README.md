# logsql

## Name

*logsql* - logs DNS queries to a SQL database.

## Description

The *logsql* plugin logs DNS queries and their associated domain names to a SQL database. It captures all
domains present in DNS responses and stores them in a database.

Domains are "upserted" into the database, meaning that if a domain already exists, only the `updated_at`
timestamp is updated for the existing database row. This allows you to track both when domains were first
queried and when they were last seen.

Currently, the plugin supports PostgreSQL and SQLite3 databases.

## Syntax

```txt
logsql DIALECT DSN
```

- **DIALECT** is the database dialect to use (`postgres` or `sqlite3`)
- **DSN** is the Data Source Name (connection string) for your database

## Examples

Log all DNS queries to a PostgreSQL database:

```corefile
. {
    logsql postgres postgres://user:password@localhost/dns_logs
    forward . 8.8.8.8
}
```

Log all DNS queries to a SQLite database:

```corefile
. {
    logsql sqlite3 ./dns_logs.db
    forward . 8.8.8.8
}
```

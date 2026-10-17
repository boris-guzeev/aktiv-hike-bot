# AktivHike Telegram Bots

## AktivHike is a modular Telegram bot system built in Go, designed to manage hiking events and bookings

## The system consists of two independent bots

* Admin Bot — manages hikes and booking workflow
* Client Bot — allows users to browse hikes and create bookings

The project follows a domain-based modular architecture with clear separation of responsibilities.

## Entry points for bots and utility commands

```
cmd/
 ├── admin-bot/
 │   └── main.go
 │
 ├── client-bot/
 │   └── main.go
 │
 ├── getchatid/
 └── seeds/
```

## Internal Structure

```
internal/
 ├── adminbot/
 ├── clientbot/
 ├── config/
 ├── db/
 └── logger/
 ```

## Admin Bot Structure

```
internal/adminbot/
 ├── booking/
 │   ├── handler/
 │   ├── repository/
 │   └── service/
 │
 ├── hike/
 │   ├── fsm/
 │   ├── handler/
 │   ├── parser/
 │   ├── repository/
 │   └── service/
 │
 ├── user/
 │   ├── repository/
 │   └── service/
 │
 ├── ui/
 │   ├── booking/
 │   ├── hike/
 │   └── common/
 │
 └── router.go
```

### Responsibilities

__booking__ - Handles admin booking workflow and status updates<br>
__hike__ - Create, edit, publish hikes and manage FSM creation flow<br>
__user__ - Admin Telegram users management<br>
__ui__ - Telegram message formatting and keyboards

## Client Bot Structure

```
internal/clientbot/
 ├── admin/
 │   ├── repository/
 │   └── service/
 │
 ├── booking/
 │   ├── handler/
 │   ├── repository/
 │   └── service/
 │
 ├── hike/
 │   ├── handler/
 │   ├── repository/
 │   └── service/
 │
 ├── user/
 │   ├── repository/
 │   └── service/
 │
 ├── ui/
 │   ├── booking/
 │   ├── hike/
 │   └── common/
 │
 └── router.go
```

### Responsibilities

* __admin__ - Ensures that admin exists when booking goes
* __booking__ - Creates bookings and handles client callbacks
* __hike__ - Displays hikes and booking buttons<br>
* __user__ - Client Telegram users<br>
* __ui__ - Telegram UI components and message builders

## Shared Infrastructure

```
internal/
 ├── app/
 ├── db/
 └── logger/
```

__app__ — application config and i18n<br>
__db__ — PostgreSQL connection and sqlc queries<br>
__logger__ — structured logging

## Architecture

Each domain follows a layered structure:

```
domain/
 ├── handler/
 ├── service/
 └── repository/
```

__handler__ — Telegram updates<br>
__service__ — business logic<br>
__repository__ — database access

## Tech Stack

* Go
* PostgreSQL
* sqlc
* Telegram Bot API
* Docker
* Structured logging

## Features

### Admin Bot

* Create hikes
* Edit hikes
* Publish hikes
* Manage bookings
* Admin workflow

### Client Bot

* Browse hikes
* Book hikes
* Notify admins
* Client notifications

## Database Backup and Restore

The backup and restore commands require Docker Compose, a running `db` service,
and valid PostgreSQL settings in `.env`.

### Create a Backup

```sh
make backup-db
```

The command creates a PostgreSQL custom-format dump in the `backups/`
directory. The file name contains the database name and a timestamp, for
example:

```text
backups/aktivhike_20261017_143000.dump
```

The dump is first written to a temporary file and validated with `pg_restore`
before it is moved to its final path. The `backups/` directory is excluded from
Git.

To use another output directory, set `BACKUP_DIR`:

```sh
make backup-db BACKUP_DIR=/path/to/backups
```

### Restore a Backup

> [!WARNING]
> Restoring a backup completely replaces the current database.

Specify the dump file and explicitly confirm the operation:

```sh
make restore-db \
  BACKUP_FILE=backups/aktivhike_20261017_143000.dump \
  CONFIRM_RESTORE=yes
```

Before replacing the database, the command validates the dump and stops the
running admin and client bot services. After a successful restore, it restarts
only the bot services that were running beforehand. If the restore fails, the
bot services remain stopped so that the incomplete database is not used.

Database dumps do not contain files from the `hike_storage` Docker volume.
Back up that volume separately when a full application backup is required.

## Design Goals

* Clean architecture
* Modular structure
* Easy scaling
* Maintainable codebase
* Clear separation of domains

## License

This project is licensed under the [MIT License](LICENSE).

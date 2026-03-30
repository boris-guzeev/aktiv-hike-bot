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
 ├── app/
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

## Design Goals

* Clean architecture
* Modular structure
* Easy scaling
* Maintainable codebase
* Clear separation of domains

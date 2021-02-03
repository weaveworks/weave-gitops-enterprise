# MCCP Event Writer Component

The event-writer subscribes to the NATS messaging server, converts, and stores
received messages to the database.

To build the binary on MacOS and Ubuntu run:

```bash
go build -o event-writer main.go
```

Note: in the wks build process, the binary will be built in the Dockerfile with linker flags set, that are needed as it is using sqlite and the target OS is alpine.

To start the subscribe process, first create the sqlite database at a given path which can
be set by passing the `--db-uri` flag, or exporting an `DB_URI` env var:

```bash
./event-writer database create --db-uri test.db
```

then run the event-writer passing the database path, the NATS server URL and the NATS subject:

```bash
./event-writer run --db-uri test.db --nats-url nats://nats-server-url:4222 --nats-subject test.subject
```

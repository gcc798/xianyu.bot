# PostgreSQL initialization

The Go application uses GORM and runs its schema migrations against PostgreSQL at startup. It uses a new, empty `xianyu_assistant` database; existing SQLite data was intentionally discarded, and the prior `xianyu` PostgreSQL database was left untouched.

The Docker server connects to `postgres:5432`; local development connects to `127.0.0.1:15432`. Both use `xianyu_assistant` (user/password: `xianyu` / `xianyu_assistant`). After starting the Compose server, verify `http://127.0.0.1:18888/health` and inspect the schema through the host-mapped PostgreSQL port.

GORM removes the need for hand-written SQL for ordinary models, but it does not automatically translate/import a SQLite database file. If a future migration needs to preserve data, it must explicitly map values and verify row counts and constraints before retiring the source database.

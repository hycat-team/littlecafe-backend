# Database bootstrap

`00-init.sql` runs once when the Postgres data volume is first created
(mounted into `/docker-entrypoint-initdb.d` by `deployments/docker-compose.yml`).
It only creates extensions — never tables.

Schema evolution belongs to per-service Goose migrations:

```bash
make migrate-create svc=<service> name=<change>
make migrate-up svc=<service>
```

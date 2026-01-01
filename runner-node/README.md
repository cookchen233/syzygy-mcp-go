# syzygy-runner-node

Node runner for Syzygy spec.

## Install

```bash
npm install
npx playwright install
```

## Run

```bash
MYSQL_HOST=... MYSQL_USER=... MYSQL_PASSWORD=... MYSQL_DATABASE=... \
node ./bin/syzygy-runner.js /path/to/spec.json
```

Env:
- `MYSQL_HOST` `MYSQL_PORT` `MYSQL_USER` `MYSQL_PASSWORD` `MYSQL_DATABASE`
- `HEADLESS=0` to run headed

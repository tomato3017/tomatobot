### Tomatobot ###
A modular bot for telegram that allows topic publishing, callbacks and more. Designed for expandability.

See yaml example for more.

## Docker Compose

The Compose stack starts Tomatobot with PostgreSQL and stores app/database data in Docker-managed volumes.

```bash
cp .env.example .env
cp tomatobot.postgres.example.yml tomatobot.yml
docker compose up --build
```

Fill in `TELEGRAM_TOKEN` and `WEATHER_API_KEY` in `.env` before starting the bot. Change `POSTGRES_PASSWORD` before using this stack anywhere outside local development.

Tomatobot expands environment variables in YAML config values, so `tomatobot.postgres.example.yml` can reference `${POSTGRES_USER}`, `${POSTGRES_PASSWORD}`, and `${POSTGRES_DB}` from `.env`. Compose supplies `.env` to the app container and PostgreSQL service; `tomatobot.yml` remains the source of truth for the bot database configuration.

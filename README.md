# weather-stats

A lightweight Go CLI and container that fetches the current temperature (in °C) for a postcode/location from [WeatherAPI.com](https://www.weatherapi.com/) and posts it to [Emoncms](https://emoncms.org/) as an input.

## Project Structure

- `cmd/weather-stats/main.go` — CLI entrypoint, flags, signal handling, and wiring
- `internal/config/config.go` — Environment variable loading and validation
- `internal/weather/client.go` — WeatherAPI.com HTTP client
- `internal/emoncms/client.go` — Emoncms `/input/post` HTTP client
- `internal/syncer/syncer.go` — Fetch/post orchestration and polling loop

## Environment Variables

| Variable | Required | Default | Description |
| :--- | :---: | :--- | :--- |
| `WEATHER_API_KEY` | Yes | — | Your WeatherAPI.com API key |
| `EMONCMS_API_KEY` | Yes | — | Your Emoncms **Read & Write** API key |
| `EMONCMS_NODE` | Yes | — | Emoncms node ID or name (e.g. `weather` or `10`) |
| `POSTCODE` | Yes | — | Location query such as a UK postcode (e.g. `SW1A 1AA`) |
| `EMONCMS_INPUT_NAME` | No | `temperature` | Field name created under the Emoncms node |
| `EMONCMS_URL` | No | `https://emoncms.org` | Base URL for Emoncms (change for self-hosted instances) |
| `POLL_INTERVAL` | No | `15m` | Go duration between polls (e.g. `5m`, `15m`, `1h`). Set to `0` to run once and exit |

## Deploying in Portainer (Git Repository Stack)

1. Push this repository to GitHub.
2. In Portainer, go to **Stacks → Add stack**.
3. Give the stack a name (e.g. `weather-stats`) and select **Repository**.
4. Enter your GitHub repository URL, branch (e.g. `refs/heads/main`), and leave **Compose path** as `docker-compose.yml`.
5. Under **Environment variables**, add the required variables (`WEATHER_API_KEY`, `EMONCMS_API_KEY`, `EMONCMS_NODE`, `POSTCODE`) and any optional overrides (`EMONCMS_INPUT_NAME`, `EMONCMS_URL`, `POLL_INTERVAL`).
6. Click **Deploy the stack**. Portainer will build the minimal Go image directly from the `Dockerfile` and start polling every 15 minutes.

## Running Locally

```bash
cp .env.example .env
# Edit .env with your keys and postcode
set -a && source .env && set +a

# Run once and exit
go run ./cmd/weather-stats --once

# Run continuously on POLL_INTERVAL
go run ./cmd/weather-stats
```

## Testing

```bash
go test -v ./...
```

# Spotify Now Playing (SSE)

Small Go HTTP server that polls Spotify’s **Currently Playing** endpoint and streams updates to clients via **Server-Sent Events (SSE)**.

## What it does

- Loads environment variables from a local `.env` file (dotenv-style)
- Refreshes a Spotify access token using a **refresh token** (cached for ~1 hour)
- Polls `GET https://api.spotify.com/v1/me/player/currently-playing` once per second
- Exposes:
    - `GET /` — returns the current payload as JSON (or `null` if nothing is playing)
    - `GET /sse` — SSE stream of updates (`event: spotify`)
- Sends SSE **keep-alive** comments every 15 seconds

## Project structure

- `main.go` — HTTP server, routes, graceful shutdown, and `.env` loading (godotenv)
- `types.go` — exported JSON payload types + minimal Spotify response structs
- `spotify.go` — token refresh + Spotify API call + token cache
- `sse.go` — SSE endpoint, connected-client registry, broadcaster loop
- `helpers.go` — small helpers (Spotify URLs, etc.)
- `cors.go` — adds CORS headers (useful when running the browser example on a different port)
- `index.html` — browser example client (connects to `/sse`)

## Requirements

- Go 1.20+ (newer is fine)
- Spotify app credentials + refresh token
- A `.env` file in the project root

## Setup (.env)

1. Copy the example file:

```bash
cp .env.example .env
```

2. Edit `.env` and fill in your values:

```env
SPOTIFY_CLIENT_ID=your_client_id_here
SPOTIFY_CLIENT_SECRET=your_client_secret_here
SPOTIFY_REFRESH_TOKEN=your_refresh_token_here
```

## Spotify setup (Client ID, Secret, and Refresh Token)

This project **does not** store a long-lived Spotify access token. Instead it stores a **refresh token** and uses it to fetch short-lived access tokens automatically.

You need:

- `SPOTIFY_CLIENT_ID` and `SPOTIFY_CLIENT_SECRET` (from your Spotify app in the Spotify Developer Dashboard)
- `SPOTIFY_REFRESH_TOKEN` (generated once via Spotify OAuth)

### 1) Create a Spotify app

1. Go to the [Spotify Developer Dashboard](https://developer.spotify.com/dashboard) and create an app.
2. Copy the **Client ID** and **Client Secret** into your `.env`.

### 2) Add a Redirect URI

In your Spotify app settings, add a Redirect URI, for example:

- `http://[::1]:3000`

This must match the `redirect_uri` you use in the authorize/token steps below.

### 3) Get a Refresh Token (Authorization Code flow)

#### Step A: Build the authorization URL

Pick the scopes you need. For “currently playing”, these are typical:

- `user-read-currently-playing`
- `user-read-playback-state`

Open this in your browser (replace `YOUR_CLIENT_ID` and ensure the `redirect_uri` matches your app settings):

```text
https://accounts.spotify.com/authorize?response_type=code&client_id=YOUR_CLIENT_ID&scope=user-read-currently-playing%20user-read-playback-state&redirect_uri=http%3A%2F%2Flocalhost%3A8888%2Fcallback
```

After you approve, Spotify will redirect to something like:

```text
http://[::1]:3000?code=PASTE_THIS_CODE
```

Copy the `code` value.

#### Step B: Exchange the code for tokens

Make a Base64 `client_id:client_secret` value:

```bash
printf "%s" "YOUR_CLIENT_ID:YOUR_CLIENT_SECRET" | base64
```

Then request tokens:

```bash
curl -X POST "https://accounts.spotify.com/api/token" \
  -H "Authorization: Basic YOUR_BASE64_VALUE" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "code=THE_CODE_YOU_COPIED" \
  -d "redirect_uri=http://[::1]:3000"
```

The response JSON includes:

- `access_token` (short-lived)
- `refresh_token` (this is what you want)
- `expires_in`

Copy the `refresh_token` into your `.env` as `SPOTIFY_REFRESH_TOKEN`.

> Note: Spotify may only return a refresh token the first time you authorize for a given user/app/scope combo.  
> If you don’t get one, try removing the app’s access from your Spotify account settings and repeat the authorize flow.

## Install & run (server)

1. Install dependencies:

```bash
go get github.com/joho/godotenv
```

2. Run the server:

```bash
go run .
```

Server starts on:

- `http://localhost:9700`

## Browser Example

Open `index.html` in your web browser. The page establishes a connection to the SSE endpoint at `http://localhost:9700/sse` and listens for updates from the server. Whenever your Spotify playback changes (track, pause, resume, etc.), the server streams an event to all connected clients, and the browser displays the latest currently playing information in real-time.

## Endpoints

### `GET /`

Returns the latest Spotify data as JSON.

- If Spotify returns HTTP 204 (nothing currently playing), this endpoint returns `null`.

Example:

```bash
curl http://localhost:9700/
```

### `GET /sse`

SSE stream. You’ll receive events like:

- `event: spotify`
- `data: { ...json... }`

Keep-alives are sent as comment lines:

- `: keep-alive`

Example:

```bash
curl -N http://localhost:9700/sse
```

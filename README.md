# listening

Small tool to get currently playing song from Spotify and persist it in memory.
Written in Go using just the standard library.

It also provides an endpoint to get the currently playing song.
We use a stale-while-revalidate cache to avoid hitting the Spotify API too often.

## Prerequisites

You need to define the `SL_SPOTIFY_CLIENT_ID` and `SL_SPOTIFY_CLIENT_SECRET`
environment variables with the values from your [Spotify application](https://developer.spotify.com/documentation/web-api/concepts/apps).

You also need `go` installed. You can get it from [here](https://golang.org/dl/).

## Usage

To run it locally:

```bash
go -C cmd/listening run .
```

To build it:

```bash
go -C cmd/listening build .
```

Visit `http://localhost:5050/` to begin the OAuth flow.
The currently playing song will be available at `http://localhost:5050/current`.

## Configuration

Besides the spotify client id and secret there are a few other environment
variables you can configure:

- `SL_HOST`: the host to listen on (default: `localhost`)
- `SL_PORT`: the port to listen on (default: `5050`)
- `SL_ADDR`: the address Spotify will redirect to after the OAuth flow (default: `http://$SL_HOST:$SL_PORT`)
- `SL_DEV_ORIGIN`: One of the two allowed origins for the CORS policy (default: `http://localhost:4321`)
- `SL_PROD_ORIGIN`: The other allowed origin for the CORS policy (default: `https://sraj.me`)

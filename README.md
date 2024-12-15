# listening

Small tool to get currently playing song from Spotify and persist it to memory.

It also provides an endpoint to get the currently playing song.
We use a stale-while-revalidate cache to avoid hitting the Spotify API too often.

## Prerequisites

You need to define the `SL_TOKEN` environment variable with the Spotify token.

## Usage

To run it locally:

```bash
go -C cmd/listening run .
```

To build it:

```bash
go -C cmd/listening build .
```

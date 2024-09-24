# YT2PC (YouTube to Podcast)

`yt2pc` is a program that allows you to subscribe to a YouTube playlist as if it were a podcast. It takes care of everything, from listing the episodes to downloading the audio and converting it into a standard format that any device can play.

## Requirements

- A YouTube API Key. Can be obtained from the Google Developers Console.

## Installation

### From source

1. Clone the repo: `git clone https://github.com/cquintana92/yt2pc`.
2. Build the server: `go build`.
3. Copy the `.env.sample` file into a `.env`: `cp .env.sample .env`.
4. Fill your variables in `.env`.
5. `./yt2pc`.

### Docker-compose

1. Clone the repo: `git clone https://github.com/cquintana92/yt2pc`.
2. Set up the following docker-compose file:

```yaml
services:
  yt2pc:
    build: ./
    ports:
      - "8080:8080"
    volumes:
      - "./audio_cache:/audio_cache"
    environment:
      YOUTUBE_API_KEY: "YOUR_API_KEY_HERE"
      SERVER_URL: "https://your.server.url.here"
      PORT: "8080" # Optional. Default is 8080
    restart: unless-stopped
```

## Usage

The server will listen to incoming connections at the port you specified. In order to check if the server is up, you can just try to access the `/health` route:

### Episodes

In order to get the list of episodes, you can enter an URL like the following into your podcast client:

```
https://your.server.url.here/PLAYLIST_ID.xml
```

The `PLAYLIST_ID` is the final part of a playlist URL. That means, if the playlist URL is `https://www.youtube.com/playlist?list=ABCDEFG`, you would need to use `ABCDEFG` as the `PLAYLIST_ID`.

This request will give back an OPML response that will list the videos on that playlist as if it were podcast episodes, and an URL that allows the podcast client to download the episodes in MP3 format.

Keep in mind that when you ask it to download an episode, it may take a while the first time, as it needs to download the audio from the video and then convert it into mp3, which can take some time (depending on the length of the video and the power of your sever). The converted videos are stored in the `audio_cache` directory, so once it has been downloaded, it won't need to do the process again.

For reference, the URL for downloading a video is:

```
https://your.server.here/PLAYLIST_ID/VIDEO_ID.mp3
```

> **NOTE**: The fetching for the playlist videos is cached for 2 hours, so if you re-request the list of episodes from a playlist within that time range, it won't actually fetch it from YouTube. It's an in-memory cache, so if you restart the server it will be evicted.

### Healthcheck

If you want to set up a healthcheck you can use the following URL: 

```shell
$ curl https://your.server.here/health
```

If you receive an OK response, it means the server is running.

## License

```
MIT License

Copyright (c) 2024 Carlos Quintana

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
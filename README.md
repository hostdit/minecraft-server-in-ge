# Minecraft world on planet earth

A working Minecraft 1.8.9 server whose world sits on the actual Earth. Real protocol and no server software. Blocks you place land on the planet view at the coordinates you're standing on, and Google Earth draws them within two seconds.

The world is the map, and the map is the world. Build in Minecraft and the blocks appear as points on the globe in the city and country you're standing in and the server tells you also. The seven wonders are prebuilt at their real coordinates, so you can warp to the Colosseum, build in Rome, then warp to Giza and build in Egypt (to an extent obviously).

## Why

Excel, Outlook, OBS, Obsidian, Blender, PowerPoint, VLC, Word, The Sims 4, Terraria, FL Studio, VS Code, DaVinci Resolve, Unity, Garry's Mod, now the whole planet.

Every previous episode put a world inside one program. This one puts it on the map. Minecraft coordinates are made up. These ones aren't.

## What's different about this one

The series is normally about the host app holding the port. This one isn't, it is more of a fun communicator.

Google Earth can't run code. No plugin system, no scripting runtime, nothing loads into the process. What it does do is fetch a URL on a timer.

So the server is standalone and Earth is a client polling it.

## What it does

- Server list ping with MOTD, player count and a favicon
- Offline mode login, no encryption
- Flat world, 5x5 chunks of bedrock, dirt and grass
- Creative mode with flight
- Keep-alives
- Chat, with commands in it
- Spawn is a real place. You start at the Great Pyramid of Giza
- Every block placed is converted to real latitude and longitude and appears in Google Earth within two seconds
- Every block placed is reverse geocoded offline, so chat says "Placed in Rome, Italy" as you build and the globe says the same
- The seven wonders are prebuilt in Minecraft (calm yourself, they are not perfect) at the real coordinates of the real wonders, and starred on the globe
- `/warp <name>` travels to another wonder. You reappear at the real site, and blocks you place there land in that country

## What it doesn't do

Almost everything else. No mobs, no inventory persistence, no world saving, no survival, no other dimensions.

A few limits in this build:

**One way.** Minecraft to Earth is live. Earth to Minecraft doesn't exist, because Earth only ever makes outbound requests and has no way to tell anyone what the user did.

**One player.** The status ping and the join are handled separately so they don't fight, but only one connection gets to be the player.

**Sixteen blocks tall.** The world is one chunk section, so the chunk payload stays at 12,544 bytes.

**Nothing is saved.** Close the server and the world is gone. The blocks live in a map in memory, not a Minecraft world file.

It's a server in the sense that a client connects to it and receives a world. Set your expectations accordingly.

## Requirements

Minecraft Java 1.8.9, protocol 47.

Google Earth Pro on desktop, free from google.com/earth/versions, using the "Download Earth Pro on desktop" button. Google is ending new downloads on 25 June 2027, so get it while it's there. Installed copies keep working.

Go 1.22 or newer to build. Nothing else, no runtime or libraries. Prebuilt binaries are in Releases if you'd rather not.

macOS, Windows or Linux. Built and tested on macOS 27, Apple Silicon.

## Install

**1.** Build the server. One binary, no dependencies, the geo data is baked in.

macOS or Linux:

```
cd minecraft-server-in-ge
./build.sh
```

Windows:

```
cd minecraft-server-in-ge
build.bat
```

`./build.sh all` cross compiles every platform into `dist/` if you want the lot.

**2.** Run it.

```
./gemc
```

You want `Minecraft server listening on :25565` and a KML endpoint address. Leave it running.

**3.** Connect Minecraft 1.8.9 to `localhost` via Add Server or Direct Connect.

**4.** Open the planet. In a browser go to `http://127.0.0.1:8080/earth.kml`, which downloads a small file, and open that in Google Earth Pro. It's a live link that refreshes every two seconds. The wonders show up as stars straight away.

If nothing appears, add the link instead: Add > Network Link, put `http://127.0.0.1:8080/live.kml` in the Link field.

### Favicon

A 64x64 PNG named `favicon.png` in the directory you run the binary from. Anything other than exactly 64x64 gets dropped by the client.

It's read at startup, so a new icon needs a restart.

## Use

The server starts when you run it. Nothing else to do.

That's this program. Google Earth never listens on anything, it only fetches the KML endpoint, which you can watch in the server log.

**3.** Place a block. Chat names the spot it landed on the planet.

**4.** Warp somewhere else and build there.

### Chat commands

| Command | Effect |
| --- | --- |
| `/list` | the wonders you can travel to |
| `/warp <name>` | travel to a wonder and rebuild the world around it |

Names are `colosseum`, `tajmahal`, `christredeemer`, `machupicchu`, `chichenitza`, `greatwall`, `petra`, `giza`.

### Flags

| Flag | Default | Effect |
| --- | --- | --- |
| `-mc` | `:25565` | Minecraft listen address |
| `-http` | `127.0.0.1:8080` | KML endpoint address |
| `-url` | `http://127.0.0.1:8080` | the URL Google Earth uses to reach the endpoint |

`-url` only matters when Google Earth is on a different machine from the server. Point it at the server's address, bind `-http` to `0.0.0.0:8080`, and open `earth.kml` from there.

### Endpoints

| Path | What it is |
| --- | --- |
| `/earth.kml` | the file you open in Google Earth once, a live link to the one below |
| `/live.kml` | the current world, refetched every two seconds |
| `/tour.kml` | a playable fly through of all seven wonders |

## Troubleshooting

**Nothing appears in Google Earth.** Check `curl -s http://127.0.0.1:8080/earth.kml` returns KML. If it does, the server is fine and it's the load step, so add the network link by hand as above. If Earth never fetches at all, nothing shows in the server log and the fix is to bind to your LAN IP and point the link there instead of 127.0.0.1.

**Everything says the open ocean.** The anchor isn't where you think. `/warp` somewhere and try again.

**Placing is slow with thousands of blocks.** The nearest city search is linear over ten thousand cities. Fine at the rate a person places blocks, not fine if you script it.

## How it works

**The server.** A standalone Go program, no external dependencies, one binary per platform.

**Chunks.** The reason this targets 1.8.9. In 1.8 a chunk section is a flat array of `(id << 4) | meta` shorts, then block light, then sky light. 12,544 bytes for one section plus biome data, sent uncompressed. Modern versions use palette encoded, bit packed longs and expect zlib. The join sequence is 25 chunks, so about 314KB in one burst.

**Earth as a client.** Google Earth Pro fetches a KML NetworkLink on a fixed interval. The server hands it a `<NetworkLink>` pointed at its own live endpoint with a two second refresh, so every couple of seconds Earth asks for the current world and redraws it. Nothing is pushed and nothing is diffed, the whole world goes out each time.

**Coordinates.** A per wonder anchor ties Minecraft spawn to that wonder's real latitude and longitude at one metre per block, then a local tangent plane conversion turns any Minecraft position into a real one, with the longitude scale corrected for latitude so a block stays a metre wherever you are. Over the 80 'metres' a build covers, treating the Earth as flat is accurate to well under a block.

**Reverse geocoding, offline.** Two datasets are embedded in the binary: 242 simplified country borders and 10,567 cities with coordinates. Country is a ray casting point in polygon test, with holes handled so Lesotho doesn't come back as South Africa. City is nearest by great circle distance.

**The wonders.** Each is a compact voxel model generated in code (making it look how it does was hard enough, just take it), hollow so the join doesn't ship three thousand block changes, placed at the coordinates of the real thing.

## Tests

No Minecraft and no Google Earth needed. The tests run the real server over an in memory pipe and speak the protocol to it.

```
go test ./...
```

That covers the protocol codec, the coordinate mapping and its roundtrip, the geocoder against known wonder locations, footprint sanity, and a full server flow: login, placing a block and checking it geocodes to the right country, digging it, and warping.

## Data sources

Both datasets were downloaded, cut down to the fields this needs and embedded.

Country borders: Natural Earth 50m admin 0 countries, public domain, via [github.com/nvkelso/natural-earth-vector](https://github.com/nvkelso/natural-earth-vector). Reduced to name, ISO code and geometry, with coordinates rounded to four decimal places.

Cities: geonames cities5000, CC BY 4.0, from [geonames.org](https://www.geonames.org/). Reduced to name, country, latitude and longitude.

## Licence

MIT. Do what you like with it.

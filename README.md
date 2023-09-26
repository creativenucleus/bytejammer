# Ticjammer

An attempt to replicate the [Bytejam launcher](https://github.com/glastonbridge/bytejams), in a single executable, to use WebSockets.

## Licenses

This project operates with the following sub-licenses:

- Included TIC 80 built binaries - [MIT License](https://github.com/nesbox/TIC-80/blob/main/LICENSE)

## Outline

A standalone TIC-80 battlejam launcher, that can be used to coordinate Bytejams, for education, or to showcase.

For Windows, Mac, and Linux.

The standalone includes a TIC-80 binary, which can be written to the filesystem and run.

## **IMPORTANT**

This is a work in progress **USE AT YOUR OWN RISK**.

## Runnning

(Swap `ticjammer.exe` for `ticjammer` on Linux/Mac)

### Run a Jukebox

`ticjammer.exe`

`ticjammer.exe jukebox --playlist .\playlist\trains.json`

`ticjammer.exe jukebox --playlist .\playlist\nanogems-test-selection.zip`

### Run a Server

`ticjammer.exe server --port 4444`

### Run a client

You must first create an identity:

`ticjammer.exe make-identity`

Then you can run this each time:

`ticjammer.exe client --host localhost --port 4444`

### Run a client-jukebox

`ticjammer.exe client-jukebox --host localhost --port 4444` // optional --playlist as above


### Jukebox mode

Default (no arguments) mode will launch into jukebox mode, playing random Bytejams from LCDZ.
It could also read a JSON file playlist (from remote and local), play in order or shuffle, and with a specified rotation time.
Applications:
    - To project onto a wall at parties to preach the good TIC and Bytejam words.
    - To play at events like the recent Unesco one.
    - For DJ visuals.
    - Run by a server as a placeholder player for Bytejams.
    - For people to just enjoy in their own homes.
    - An ad-hoc retro-style kiosk/ad runner.
    - For a party to showcase all the entries in a competition.

### Client mode

User specifies a server to connect to, and a port.
First run, will ask for a display name and create a private key.
Launches a TIC.
Perhaps interaction / stats will be via a web panel?
Users may create multiple identities on a machine, and copy their keys elsewhere.
The client can take snapshots of the player's code as they go.

### Server mode

Starts a server, which is open to clients on the specified port.
A web panel is available for the operator.
When clients connect, they are identified by using their key, and the connection may be approved or rejected in the web panel.
Clients can be wired to display TICs, which will spawn as appropriate.
The jukebox is also a client, just run by a robot.
The panel will allow the server operator to snapshot code, switch the links between clients and display TICs, and push code to clients.
If input+output can work on a client, also maybe useful for education:
    - The server can send some code to all clients
    - After an exercise, pull each in turn to display for showcase.

Food for thought:
Can data be sent around (palette, sprite, music)?
Automatic snapshot on each code run.
Auto-bundle for LCDZ? (or should folks be allowed to submit their own best?)
Act as a relay, fan out one code to many, or converge / round robin to one display? (applications?)
Is possible: Auto layout of OBS Studio, or a layer in between?
Code posting to a web client via WebSocket?

## TODO

- TIC-80 Version conflict
- Authentication
- Limit clients
- Clean close/open
- Web panel
- Backups
- Auto DJ


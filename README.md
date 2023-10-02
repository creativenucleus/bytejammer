# ByteJammer

(Latest Code and documentation - GitHub)[https://github.com/creativenucleus/bytejammer]

For celebration of the TIC-80 livecoding / effects scene.

- A jukebox / robot VJ that plays TIC-80 effects for personal enjoyment.
- A standalone client-server for running ByteJams.

Please read the documentation before running.

### **IMPORTANT**

This is a work in progress **USE AT YOUR OWN RISK**. It's currently early days, so things are more likely to go wrong, but [so far] nobody has reported any ill effects.

Features should currently be considered experimental, and liable to change, possibly in a way that is not backward compatible. The format of arguments for the CLI is likely to be particularly in flux, and this documentation may lag behind development (I tend to code chunks, and periodically review docs). Please feel welcome to contact me if you have questions.

## Larger Known Issues

Some files that are meant to be temporary don't get removed.

The Identity mechanism is incomplete. It's likely that this will be updated.

## Outline

This is intended as a standalone TIC-80 launcher that can be used to coordinate Bytejams, for education, or to showcase. It includes a TIC-80 binary, which will be written to the filesystem, then run.

Cross-compatible on Windows, Mac, and Linux.

## Runnning

(Swap `bytejammer.exe` for `bytejammer` on Linux/Mac. You may need to make the file executable with some chmod magic)

### Jukebox Mode

Default (no arguments) mode will launch into jukebox mode, playing random Bytejams from LCDZ.  
It can be provided with a JSON file playlist (from remote and local) or .zip file.
Applications:  
    - To project onto a wall at parties to preach the good TIC and Bytejam words.
    - To play at events like the recent Unesco one.
    - For DJ visuals.
    - Run by a server as a placeholder player for Bytejams.
    - For people to just enjoy in their own homes.
    - An ad-hoc retro-style kiosk/ad runner.
    - For a party to showcase all the entries in a competition.

`bytejammer.exe` (default - Livecode DemoZoo ByteJam playlist)

`bytejammer.exe jukebox --playlist .\playlist\trains.json` (play a JSON playlist)

`bytejammer.exe jukebox --playlist .\playlist\nanogems-test-selection.zip` (play files from a .zip)

### Server Mode

Starts a server, which is open to clients on the specified port.  
A web panel is available for the operator.  
When clients connect, they are identified by using their key, and the connection may be approved or rejected in the web panel.  
Clients can be wired to display TICs, which will spawn as appropriate.  
The jukebox is also a client, just run by a robot.  
The panel will allow the server operator to snapshot code, switch the links between clients and display TICs, and push code to clients.  
If input+output can work on a client, also maybe useful for education:  
    - The server can send some code to all clients
    - After an exercise, pull each in turn to display for showcase.

`bytejammer.exe server` // (Optional: can specify `--localport 4444`)

The CLI will provide a link to a local server. Open that in a web browser.

### Client Mode

User specifies a server to connect to, and a port.  
First run, will ask for a display name and create a private key.  
Launches a TIC.  
Perhaps interaction / stats will be via a web panel?  
Users may create multiple identities on a machine, and copy their keys elsewhere.  
The client can take snapshots of the player's code as they go.  

`bytejammer.exe client`  // (Optional: can specify `--localport 1000`)

The CLI will provide a link to a local server. Open that in a web browser.  

You must first create an identity.  

Then you can connect to a server with an identity.  

### Client-jukebox Mode

You can connect a client-jukebox to a remote server, as you would a regular client.

`bytejammer.exe client-jukebox --host localhost --port 4444` // optional --playlist as above

## Development and Ideas

### TODO

- Logs on web panels.
- Automatic snapshot on each code run.
- Better gatekeeping of client 'lobby'.
- Improve web panels.
- TIC-80 version management
- Authentication by key.
- Clean close/open.
- Messaging feature.
- Obfuscate session
- Rationalise capitalisation/skewer of AJAX/WS data.

### Ideas

- Jukebox - play in order or shuffle, and with a specified rotation time.
- Server - set up different playlists
- Server - push code snapshots around to clients
- Autospawn + Limit clients
- Can data be sent around (palette, sprite, music)?
- Auto-bundle code for submission to LCDZ? (or should folks be allowed to submit their own best?)
- Act as a relay (hub), fan out one code to many, or converge / round robin to one display? (applications?)
- Is possible: Auto layout of OBS Studio, or a layer in between?
- Code posting to a web client via WebSocket?
- Support other fantasy consoles.

## Alternatives

ByteJammer builds on the [Bytejam launcher](https://github.com/glastonbridge/bytejams) which is less featured, but battle tested.

## Thanks

For the previous work ByteJammer builds on, testing, good-will, and support:

Aldroid, Gasman, Lex Bailey, Mantratronic, NesBox, NuSan, PS, Raccoon Violet, Superogue, Totetmatt.

Thanks to those whose work features on the [walkthrough video](https://youtube.com/watch?v=erhyvrGxwZY):

Alia, Aldroid, Dave84, Gasman, Lex Bailey, Gigabates, Mantratronic, NuSan, PS, Superogue, Suule, Synesthesia, TôBach.

Thanks to the TIC-80 livecoders, Monday Night ByteJammers, and anyone I forgot!

## Licenses

This project operates with the following sub-licenses:

- Included TIC 80 built binaries - [MIT License](https://github.com/nesbox/TIC-80/blob/main/LICENSE)

## Links and References

[Listing / DemoZoo](https://demozoo.org/productions/330626/)
[Listing / Pouët](https://pouet.net/prod.php?which=95232)
[Overview Video / YouTube](https://youtube.com/watch?v=erhyvrGxwZY)

[Bytejammer title.tic]

Hello!

I'd like to tell you about my tool: ByteJammer, but first I'd like to make sure you know about TIC-80.

[TIC-80 tic]

-> TIC is Fantasy Console that runs on your computer.

-> It has a retro style that feels like something from the Super Nintendo era.

-> It's very accessible - the default language is Lua and there's a sensible collection of commands you can use.
If you're feeling old-school, you can even peek and poke around to get the most out of it.

It's capable of lots of different kinds of effects and I love it because I can just pick it up and start swishing code around like a brush on canvas.

-> It makes me feel like Bob Ross, the jazz-style painter...

-> [Bob Ross .tic]

One corner of demoscene, Field-FX, love TIC-80 too.

Some excellent members of that community organise Byte Jams - they're a mix of social chat and creative play with live music.
It happens on Monday evenings, and it's beamed out on Twitch.

Four people dial in to someone's computer, and they all work on their own code, each having a corner of the screen.
-> It's like having four Bob Rosses all at once.
[four bob rosses .tic]

-> This tickles the same creative bits as shader showdowns, but with a lower barrier to entry, and it's a lot more casual.
-> There's a music mix, live DJ, and even live instruments sometimes.
-> I'll play a few people's effects while I talk, so you can get an idea of TIC's capabilities.
-> [playlist]
-> To make Byte Jams work, there's a modified Bytebattle edition of TIC, which dumps code to a file and posts it through the internet.
-> This version of TIC also has FFT integrated, so many of the effects are sound-responsive.

The software to run it is good, and has been a collaboration between shining stars of this community.
But there are a few things that could be improved:

-> Primarily, smoother setup for players and host
-> On the client side there a couple of scripts to run, and Python needs to be installed
-> On the server side, the host needs to set up a layout for the players, and co-ordinate them as they join.
-> The system communicates with UDP - that means data packets fly off into the internet, and sometimes to the wrong place without raising any errors, which can make joining tricky.
-> Folks also participate on PC, Mac and Linux, and there's usually a bit of fiddling to onboard new players.

Bytejammer is a new rebuild of the stack and it makes some helpful improvements:
-> There's now one standalone executable, no extra scripts and no extra installs
-> It's multiplatform (written in Go, and builds automatically on GitHub)
[setup client]
-> There's a setup panel for client
[setup server]
-> There's a setup panel for server
-> The client can set up an identity so they can be identified by the server
-> Bytejammer communicates via Websockets
    -> that means there's a managed connection - and we can tell if there are disconnects.
    -> Websockets can also talk both ways, so we can do interesting things like sending code the other way - so not just from client to server, but back from the server to the client. I've adapted the TIC Bytebattle version so it can both import and export code at the same time.
-> We're also able to shim the code as it passes through the system, which means we can do things like add some display code for player names that only runs on the server.
-> Best of all, it has a digital Bob Ross for an icon [image]

[playlist.tic]

There are some other things that Bytejammer will open up.
-> It'll be possible to do periodic snapshots for better recovery when people disconnect, and we can also use that to replay effects being built and to archive them.
-> The two-way connection means we can experiment with new ways to play, for example:
    -> We can pass code between people - so a few players can build an effect together.
    -> We could send all players a themed bit of code as a starting point
    -> We can select code from one of the players and send it to the rest so that we have an effect that evolves by selection.
-> Pushing code around also opens up a workshop setup - an educator can send some starter code to students, and pull back from individuals to showcase or review.
-> For Byte Jam hosts, the options for setup will expand:
    -> I've been working with Noosan on some experimental modifications to the Bonzomatic launcher, so it can help manage screen layouts for TIC code, and we might also be able to use ByteJammer to add party features to other fantasy consoles. 
-> Overall, Bytejammer offers better management, and that means it's easier to set up more players. Imagine Bob Rosses, all doing their thing! [16 bob ross .tic]

-> So, these are all good improvements, but the thing I'm most excited about is another mode, built to help people enjoy and showcase the amazing things creative folks have made with TIC-80.
-> Let me introduce the ByteJammer Jukebox!
    - This runs locally, on your own computer, and internally wraps some of the server and client functionality together.
    - In fact, you've been watching a recording of it while I've been speaking.
-> We can put bunch of code files in a zip to play [demo]. This is great for showcasing a collection, maybe your favourites, one person's catalogue, or submissions for a compo. You could even use it to play though retro-style animated slides, like a kiosk mode.
-> With the playlist of FFT bangers I put together, it could also be your virtual VJ.
-> If you want to add extra context for each effect, you can add an index file [demo]
-> And - and I think this is the most fun - if you don't provide a playlist, it'll connect to Live Code Demo Zoo - a brilliant archive preserving the effects from previous Byte jams. It's a treasure trove of wonder, and currently contains 179 delights.

-> And as a final flourish, the jukebox can also be a client for the server version, so livecoders can jam alongside one or more windows that play existing effects.

So, please try Bytejammer. Sit back and enjoy the creativity of our community, and maybe join us on Monday nights for ByteJam and chat!
You should follow Field-FX on Mastodon, X, or Discord for more details.

And If you'd like to see Bytejammer running, please ask Violet nicely and she'll show you :)

Bye!


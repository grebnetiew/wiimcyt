wiimcyt
=======

Youtube search proxy for WiiMC.

How to youtube
--------------

I am assuming you have WiiMC set up. The wiimc.org search proxy is down, 
so you must roll your own. I recommend a server, but this can be your PC as well.

First, you need to download the file `youtube.go` from above to your server.
Then, you need to download the compiler for `go`, the language this program is
written in. It's available at http://golang.org/

You will then need the server's IP address.
By default the program serves on port 8089, but you can change this in the code.
Open `online_media.xml` (on the wii sd card, `apps/wiimc`) and replace the url 
for Youtube - Search with 

    http://yourServersIP:8089/youtube?q=

Additionally, you can have the program load recent videos from your subscriptions
by replacing the URL in the regular Youtube line by

    http://yourServersIP:8089/youtube?s=YourYoutubeUsername

Run the program on your server by opening a command window (in Windows, go to the
downloaded file, hold Shift and right-click near it, and choose 'open command window
here') and entering:

    go build youtube.go && ./youtube

The process will log on your command line.

Settings
--------

Two things can be tuned to your liking (without much coding, obviously). Find the 
`const` declaration in the file `youtube.go`. You can set a different port there for
the server to listen on.

Also, you can instruct the proxy to send unicode characters in the file list. The
default behaviour replaces them, as the WiiMC font doesn't include glyphs to display
them. If you change the font you might want to set this to `true`.

Troubleshooting
---------------

WiiMC's youtube support is flaky. I had to set the desired quality to Medium to get 
any results (Settings, Online Media). Even then, you may experience random crashes, 
especially when switching videos. Nobody told you this was easy. :)

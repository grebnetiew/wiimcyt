wiimcyt
=======

Youtube search proxy for WiiMC.

How to youtube
--------------

I am assuming you have WiiMC set up. The wiimc.org search proxy is down, 
so you must roll your own. I recommend a server, but this can be your PC as well.

You will need the server's IP address.
By default the proxy serves on port 8089, but you can change this in the code.
Open `online_media.xml` (on the wii sd card, `apps/wiimc`) and replace the url 
for Youtube - Search with 

    http://yourServersIP:8089/youtube?q=

Run the program on your server:

    go run youtube.go

The process will log on your command line.

Settings
--------

Two things can be tuned to your liking (without much coding, obviously). Find the 
`const` declaration in the file `youtube.go`. You can set a different port there for
the server to listen on, and you can change the user whose subscription feed is
loaded when the request is empty.

Troubleshooting
---------------

WiiMC's youtube support is flaky. I had to set the desired quality to Medium to get 
any results (Settings, Online Media). Even then, you may experience random crashes, 
especially when switching videos. Nobody told you this was easy. :)

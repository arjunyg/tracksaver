
Tracksaver
==========
There are two components to the Tracksaver application. The web server, written
in Golang, has a basic auth page and a web API for adding songs to the signed-in
user's Spotify library. Since the web server listens on 127.0.0.1, it is safe
to run it as a single user system.

The second component, is an Applescript which can be bound to a global key
shortcut using applications such as Quicksilver, etc. The Applescript takes
advantage of Spotify's Applescript API to send the track ID of the currently
playing song to the web server, effectively creating a global shortcut for
saving the currently playing song.
Obviously the Applescript is specific to Mac OS X, and Windows/Linux users
will need to find another way to get the track ID, and make the API call to
the local server.

API Keys
========
For the Spotify API to function, one must register the web app on Spotify's
developer tools, and fill in your client ID and client secret in main.go.

Bugs
====
Sometimes the API key will not be renewed properly (the renewal token expires)
and the user must sign in again. If it's not working, try signing in again at
http://127.0.0.1:3001/

Add Song API
============
`http://127.0.0.1:3001/addSong?ids=<track id array>`

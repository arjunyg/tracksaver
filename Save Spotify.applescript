tell application "Spotify" to set t to id of current track
set text item delimiters of AppleScript to ":"
do shell script "curl -d id=" & third text item of t & " http://localhost:3001/addSong"
tell application "Spotify"
	display notification "by " & artist of current track with title "Saved Spotify Track" subtitle "" & name of current track
end tell
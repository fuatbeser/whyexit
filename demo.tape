# Render with: vhs demo.tape
Output demo.gif

Set Shell "bash"
Set FontSize 16
Set Width 1100
Set Height 380
Set Padding 20
Set TypingSpeed 40ms

Hide
Type `set -o history; export PS1='$ '; eval "$(whyexit init bash)"; clear`
Enter
Show

Type "whyexit 137"
Sleep 300ms
Enter
Sleep 2s

Type "whyexit --context curl 6"
Sleep 300ms
Enter
Sleep 2s

Type "curl -s localhost:1"
Sleep 300ms
Enter
Sleep 2s

Type "nosuchcommand"
Sleep 300ms
Enter
Sleep 3s

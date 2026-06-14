# Demo tape for csdn. Rendered with ascii-gif (github.com/tamnd/ascii-gif),
# which supplies the window chrome and theme; this file is just the action.
#
#   ascii-gif render docs/demo/csdn.tape -o docs/static/demo.gif
#
# csdn must be on PATH inside the recording shell.

Hide
Type "export PS1='$ ' PATH=/tmp:$PATH"
Enter
Type "clear"
Enter
Show

Sleep 500ms
Type "csdn hot --fields rank,title,score -n 8"
Sleep 800ms
Enter
Sleep 3.5s

Type "csdn hot -n 5 -o jsonl | jq -r '.title'"
Sleep 800ms
Enter
Sleep 4s

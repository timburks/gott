
# gott

**gott** is a little text editor that was created as a hobby project.
It was inspired by [antirez's kilo](http://antirez.com/news/108)
and Jeremy Ruten's
[Build Your Own Text Editor](http://viewsourcecode.org/snaptoken/kilo/).
**gott** means "good" in Swedish, perhaps a presumptuous thing to
call a little console-based text editor.

## Implementation and Goals

**gott** is written in Go and uses
[nsf/termbox-go](https://github.com/nsf/termbox-go) for screen display.
All the termbox dependencies are isolated in the screen package
in the hope that the rest of **gott** can be used on other platforms via
[gomobile](https://godoc.org/golang.org/x/mobile/cmd/gomobile).

**gott** integrates [golisp](https://github.com/SteelSeries/golisp)
for scripting. I guess it's clear where I'd like to go with that.

Along those lines, I hope to integrate many of the Go-specific features
that are often added as 
[emacs extensions](http://tleyden.github.io/blog/2014/05/22/configure-emacs-as-a-go-editor-from-scratch/).

## Legal

**gott** is released under the Apache License, version 2.0.

## Author

Tim Burks<br/>
Los Altos, CA

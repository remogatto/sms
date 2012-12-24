# SMS - A concurrent Sega Master System emulator

SMS is a free Sega Master System emulator written in
[Go](http://golang.org). I think it's the first SMS emulator written
in this language.

# Quick start

Installing and starting SMS with Go is simple:

    go get -v github.com/remogatto/sms/
    ./sms game.sms

# Description

SMS is based on a
[concurrent](http://github.com/remogatto/gospeccy/wiki/Architecture)
architecture very similar to
[GoSpeccy](http://github.com/remogatto/gospeccy), another emulator
written in Go.

The primary source of inspiration for SMS was
[Miracle](http://xania.org/miracle/miracle.html), a cool
Javascript SMS emulator.

If you like this project, please star it on
[github](http://github.com/remogatto/sms)! Bug reports and testing are
also appreciated! And don't forget to fork and send patches, of
course.

# Features

* Complete Zilog Z80 emulation
* Concurrent [architecture](http://github.com/remogatto/gospeccy/wiki/Architecture)
* SDL backend
* 2x scaler and fullscreen

# Todo

* Sound support
* Write more tests

# Key bindings

    Host computer   Sega Master System
    ----------------------------------
    Arrows          Joypad directions
    X               Fire 1
    Z               Fire 2

For more info about key bindings see file <tt>input.go</tt>

# Proprietary games

Generally, SMS games are protected by copyright so none of them
is included in GoSpeccy. However, it is possible to find tons of games
for the Sega Master System on the Internet.

# Credits

* Thanks to [⚛](http://github.com/0xe2-0x9a-0x9b) for his work on
  GoSpeccy which served as inspiration for the SMS architecture.
* Thanks to Matt Godbolt for his SMS Javascript emulator.

# Contacts

* andrea.fazzi@alcacoop.it
* [Twitter](http://twitter.com/remogatto)
* [Google+](https://plus.google.com/u/0/100271912081202470197/posts/p/pub)

# License

Copyright (c) 2012 Andrea Fazzi

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.


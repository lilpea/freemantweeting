# freemantweeting
A twitter bot ([@ramblingfreeman](https://twitter.com/ramblingfreeman)) that uses markov chains to procedurally generate freeman's mind quotes.

## Installing
This uses [golang](https://golang.org/), so you are going to have to [install that.](https://golang.org/doc/install) After that's done, do `go get github.com/lilpea/freemantweeting` and then `go install github.com/lilpea/freemantweeting` to get the executable (put it in the same directory as `data.txt`, `authentication.json`, and `configuration.json`). I'm currently using Task Scheduler to schedule the program, so there isn't anything to manage that.

## Credits
I based this off https://golang.org/doc/codewalk/markov/, check that if you want more info on how this works.

Special thanks to Danielsangeo for transscribing all of Freeman's dialogue.

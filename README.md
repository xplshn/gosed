Fork of GoSed for https://github.com/xplshn/a-utils

W.I.P: Removing old Go conventions and making the codebase more idiomatic
W.I.P: Implement all of the functionality/commands described in [Unix's 10th edition `Sed` implementation](https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed), OpenBSD's `sed` [`manpage`](https://man.openbsd.org/sed.1) as reference when clarification is needed. 

- Using POSIX RegExps
- Added: Support for both 'n' and 'N' in n_cmd.go
- Added: Made this work, compile and run the tests just fine within Go1.20+
- Added: Hope and faith
- Added: Better Usage/Help page using "Consistent CMD"(ccmd) from [https://github.com/xplshn/a-utils](https://github.com/xplshn/a-utils/tree/master/pkg/ccmd)
- Removed non-POSIX flag `-l`
- Fixed +50 warnings/errors `revive` detected
- Added comments to the code

ORIGINAL README
---------------
This is my Go language learning project. It's a basic implementation of the
utility sed. I'm not really looking for criticism, but if someone wants to
help out turning this int a real sed I wouldn't mind.

I mainly put this up as sample code for others to read. Probably not good
sample code, but it is something.

I decided on the MIT open source license. See LICENSE for more information.

The sed.html documentation is copyright The Open Group and I lifted it from
their site. I used it as a specification and this version implements it.
(Mostly.)

For what is implemented gosed acts like sed. There is one major difference,
gosed uses Go's regular expression library which is different from the one sed
uses.

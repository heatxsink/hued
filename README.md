hued
==========
A tablet/phone web interface using Philips HUE API.

Prerequisites
---

* golang 1.12 or higher.

Quickstart
---

	$ go get -u github.com/magefile/mage
	$ cd $GOPATH/src/github.com/magefile/mage
	$ go run bootstrap.go
	$ go get github.com/heatxsink/hued
	$ cd $GOPATH/src/github.com/heatxsink/hued
	$ mage build
	$ ./hued -logtostdout


Bugs and Contribution(s)
---
Please feel free to reach out. Issues and PR's are always welcome!

License
---
The MIT License (MIT)

Copyright (c) 2020 Nicholas Granado

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

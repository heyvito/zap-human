# zap-human ⚡️🧍

**zap-human** is a [uber-go/zap](https://github.com/uber-go/zap) encoder for 
humans. It does not attempt to be performant, but readable, and **is not intended
for use in production environments**.

## Installing

Download the package:

```console
$ go get -u github.com/heyvito/zap-human
```

Then, import it for side-effects:

```go
package main

import (
	_ "github.com/heyvito/zap-human"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)
```

Finally, configure your logger using `"human"` as the encoder:

```go
func main() {
    config := zap.NewDevelopmentConfig()
    config.Encoding = "human"
    config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
    config.DisableCaller = true
    logger, err = config.Build()
	if err != nil {
		// ...
	}
	zap.ReplaceGlobals(logger)
}
```

## Acknowledgments
zap-human is built on top of zap's `JSONEncoder`, and uses several portions of
it. Portions are Copyright (c) 2016-2017 Uber Technologies, Inc.

## LICENSE

zap-human is licensed under the same license as Uber's zap:

```
Copyright (c) 2016-2017 Uber Technologies, Inc.
Copyright (c) 2023 Victor Gama de Oliveira 

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
```

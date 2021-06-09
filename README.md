# glogp
glogp is a Golang logging tool.

## Usage

~~~ go
package main

import (
	"time"

	"github.com/nikchis/glogp"
)

func main() {
	glogp.SetLevel("INFO")

	glogp.Infof("Info message")
	glogp.Debugf("Debug message")

	time.Sleep(time.Millisecond * 10)
	glogp.Warnf("Warn message")
}
~~~

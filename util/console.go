package util

import (
	"fmt"

	"github.com/mbndr/figlet4go"
)

func ConsoleGreet(name, version, host string, port interface{}) {
	ascii := figlet4go.NewAsciiRender()
	options := figlet4go.NewRenderOptions()
	options.FontColor = []figlet4go.Color{figlet4go.ColorRed}
	renderStr, _ := ascii.RenderOpts(name, options)
	if host == "" {
		host = "0.0.0.0"
	}
	fmt.Print(renderStr)
	fmt.Printf("\nv%s\n", version)
	fmt.Printf("\nlistening on %s:%v\n", host, port)
}

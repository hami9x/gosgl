package main

import (
	"fmt"
	"runtime"
	"time"
	m "github.com/phaikawl/gomain"

	gl "github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
	sgl "github.com/phaikawl/gosgl"
)

func errorCallback(err glfw.ErrorCode, desc string) {
	fmt.Printf("%v: %v\n", err, desc)
}

func main() {
	go m.Do(func() {
		defer m.Exit()
		runtime.LockOSThread()
		glfw.SetErrorCallback(errorCallback)

		if !glfw.Init() {
			panic("Can't init glfw!")
		}
		defer glfw.Terminate()

		w, h := 400, 300
		window, err := glfw.CreateWindow(w, h, "Testing", nil, nil)
		if err != nil {
			panic(err)
		}

		window.MakeContextCurrent()
		gl.Init()
		sgl.Init()
		canv := sgl.MakeCanvas(w, h)
		c := sgl.MakeQuadraticCurve(50, 50, 300, 0, 400, 300)

		gl.ClearColor(0, 0, 0, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		for !window.ShouldClose() {
			c.Draw(canv)
			//Do OpenGL stuff
			window.SwapBuffers()
			glfw.PollEvents()
			time.Sleep(50 * time.Millisecond)
		}
	})
	m.Main()
}

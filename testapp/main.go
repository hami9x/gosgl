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

		w, h := 800, 600
		window, err := glfw.CreateWindow(w, h, "Testing", nil, nil)
		if err != nil {
			panic(err)
		}

		window.MakeContextCurrent()
		gl.Init()
		sgl.Init()
		canv := sgl.MakeCanvas(w, h)
		pa := sgl.NewPath().StartAt(sgl.Pt(50, 50))
		pa.QuadraticTo(sgl.Pt(80, 50), sgl.Pt(150, 0))
		pa.QuadraticTo(sgl.Pt(300, 300), sgl.Pt(800, 100))
		pa.QuadraticTo(sgl.Pt(100, 600), sgl.Pt(0, 300))
		pa.BezierTo(sgl.Pt(50, 50), sgl.Pt(-100, 200), sgl.Pt(100, 100))

		gl.ClearColor(1, 1, 1, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		pa.Draw(canv)
		for !window.ShouldClose() {
			//Do OpenGL stuff
			window.SwapBuffers()
			glfw.PollEvents()
			time.Sleep(100 * time.Millisecond)
		}
	})
	m.Main()
}

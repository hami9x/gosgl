package main

import (
	"fmt"
	"runtime"
	"time"
	m "github.com/phaikawl/gomain"

	gl "github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
	glh "github.com/go-gl/glh"
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

		window, err := glfw.CreateWindow(640, 480, "Testing", nil, nil)
		if err != nil {
			panic(err)
		}

		window.MakeContextCurrent()
		gl.Init()
		vao := gl.GenVertexArray()
		vao.Bind()

		vbo := gl.GenBuffer()
		vbo.Bind(gl.ARRAY_BUFFER)
		vertices := []float32{
			0.0, 0.5, 0.0, 0.0,
			0.5, -0.5, 0.5, 0.0,
			-0.8, -0.5, 1.0, 1.0,
		}
		gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, vertices, gl.STATIC_DRAW)

		vsh := sgl.ShaderFromFile(gl.VERTEX_SHADER, "vshader.glsl")
		fsh := sgl.ShaderFromFile(gl.FRAGMENT_SHADER, "fshader.glsl")

		program := glh.NewProgram(vsh, fsh)
		program.BindFragDataLocation(0, "outColor")
		program.Use()

		posAttr := program.GetAttribLocation("position")
		posAttr.AttribPointer(2, gl.FLOAT, false, 4*4, uintptr(0))
		posAttr.EnableArray()

		texAttr := program.GetAttribLocation("texcoord")
		texAttr.AttribPointer(2, gl.FLOAT, false, 4*4, uintptr(8))
		texAttr.EnableArray()

		tex := glh.NewTexture(1, 1)
		tex.Init()

		gl.ClearColor(0, 0, 0, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		for !window.ShouldClose() {
			gl.DrawArrays(gl.TRIANGLES, 0, 3)
			//Do OpenGL stuff
			window.SwapBuffers()
			glfw.PollEvents()
			time.Sleep(50 * time.Millisecond)
		}
	})
	m.Main()
}

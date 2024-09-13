package main

import (
	"fmt"
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	width  = 1920
	height = 1080
)

var (
	vertices = []uint16{
		0, 0, // Top-left
		1920, 0, // Top-right
		0, 1080, // Bottom-left
		1920, 0, // Top-right
		1920, 1080, // Bottom-right
		0, 1080, // Bottom-left
	}
	colors = []uint8{
		150, 100, 100, // First triangle color
		255, 200, 200, // Second triangle color
	}
	vertexShaderSource = `
		#version 330 core
		layout (location = 0) in vec2 aPos;
		uniform vec3 uColor;
		out vec3 ourColor;
		void main() {
			gl_Position = vec4(aPos.x / 960.0 - 1.0, 1.0 - aPos.y / 540.0, 0.0, 1.0);
			ourColor = uColor;
		}
	` + "\x00"
	fragmentShaderSource = `
		#version 330 core
		in vec3 ourColor;
		out vec4 FragColor;
		void main() {
			FragColor = vec4(ourColor, 1.0);
		}
	` + "\x00"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "OpenGL Triangles", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	program := initOpenGL()

	for !window.ShouldClose() {
		draw(window, program)
	}
}

func initOpenGL() uint32 {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)

	return prog
}

func draw(window *glfw.Window, program uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT)

	gl.UseProgram(program)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 2*len(vertices), gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 2, gl.UNSIGNED_SHORT, false, 2*2, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	colorUniform := gl.GetUniformLocation(program, gl.Str("uColor\x00"))

	// Draw first triangle
	color1 := []float32{float32(colors[0]) / 255.0, float32(colors[1]) / 255.0, float32(colors[2]) / 255.0}
	gl.Uniform3fv(colorUniform, 1, &color1[0])
	gl.DrawArrays(gl.TRIANGLES, 0, 3)

	// Draw second triangle
	color2 := []float32{float32(colors[3]) / 255.0, float32(colors[4]) / 255.0, float32(colors[5]) / 255.0}
	gl.Uniform3fv(colorUniform, 1, &color2[0])
	gl.DrawArrays(gl.TRIANGLES, 3, 3)

	window.SwapBuffers()
	glfw.PollEvents()
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := make([]byte, logLength)
		gl.GetShaderInfoLog(shader, logLength, nil, &log[0])

		return 0, fmt.Errorf("failed to compile %v: %v", source, string(log))
	}

	return shader, nil
}

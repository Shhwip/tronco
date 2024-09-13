package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	width          = 1920
	height         = 1080
	numTriangles   = 50 // Number of triangles per scene
	numScenes      = 50 // Number of scenes to generate
	framesPerScene = 60 // Number of frames to display each scene (60 frames = 1 second at 60 FPS)
)

var (
	vertices [][]uint16
	colors   [][]uint8
)

var (
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
	rand.Seed(time.Now().UnixNano())
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

	window, err := glfw.CreateWindow(width, height, "OpenGL Multiple Scenes of Triangles", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	program := initOpenGL()
	generateScenes(numScenes, numTriangles)

	vao, vbo := setupBuffers()

	currentScene := 0
	frameCount := 0

	for !window.ShouldClose() {
		draw(window, program, vao, vbo, currentScene)

		frameCount++
		if frameCount >= framesPerScene {
			frameCount = 0
			currentScene = (currentScene + 1) % numScenes
		}
	}
}

func generateScenes(scenes, triangles int) {
	vertices = make([][]uint16, scenes)
	colors = make([][]uint8, scenes)

	for s := 0; s < scenes; s++ {
		vertices[s] = make([]uint16, triangles*6) // 3 vertices per triangle, 2 coordinates per vertex
		colors[s] = make([]uint8, triangles*3)    // 1 color per triangle, 3 components per color

		for i := 0; i < triangles; i++ {
			// Generate random triangle vertices
			for j := 0; j < 6; j++ {
				if j%2 == 0 {
					vertices[s][i*6+j] = uint16(rand.Intn(width))
				} else {
					vertices[s][i*6+j] = uint16(rand.Intn(height))
				}
			}

			// Generate random color for the triangle
			for j := 0; j < 3; j++ {
				colors[s][i*3+j] = uint8(rand.Intn(256))
			}
		}
	}
}

func setupBuffers() (uint32, uint32) {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)

	gl.VertexAttribPointer(0, 2, gl.UNSIGNED_SHORT, false, 2*2, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	return vao, vbo
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

func draw(window *glfw.Window, program uint32, vao uint32, vbo uint32, scene int) {
	gl.Clear(gl.COLOR_BUFFER_BIT)

	gl.UseProgram(program)
	gl.BindVertexArray(vao)

	// Update buffer data for the current scene
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 2*len(vertices[scene]), gl.Ptr(vertices[scene]), gl.STATIC_DRAW)

	colorUniform := gl.GetUniformLocation(program, gl.Str("uColor\x00"))

	for i := 0; i < numTriangles; i++ {
		color := []float32{
			float32(colors[scene][i*3]) / 255.0,
			float32(colors[scene][i*3+1]) / 255.0,
			float32(colors[scene][i*3+2]) / 255.0,
		}
		gl.Uniform3fv(colorUniform, 1, &color[0])
		gl.DrawArrays(gl.TRIANGLES, int32(i*3), 3)
	}

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

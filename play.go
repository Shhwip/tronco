package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type triangles struct {
	points []uint16
	colors []uint8
	length uint16
}

func readData(folderPath string) ([]triangles, error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	var triangles []triangles
	for i, file := range files {
		if file.IsDir() {
			continue
		}
		if i == 0 {
			continue
		}
		triangles = append(triangles, processFile(folderPath+"/"+"frame"+strconv.FormatInt(int64(i), 10)+".bin"))
	}

	return triangles, nil

}

func processFile(filePath string) triangles {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return triangles{}
	}
	defer file.Close()

	length, nodes, colors, err := readBinaryData(file)

	if err != nil {
		fmt.Printf("Error reading binary data: %v\n", err)
		return triangles{}
	}

	var tri triangles
	tri.length = uint16(length)
	tri.points = nodes
	tri.colors = colors

	return tri
}

func readBinaryData(r io.Reader) (length int, nodes []uint16, colors []byte, err error) {
	// Read length (2 bytes, big-endian)
	var lengthBytes [2]byte
	if _, err := io.ReadFull(r, lengthBytes[:]); err != nil {
		return 0, nil, nil, err
	}
	length = int(binary.BigEndian.Uint16(lengthBytes[:]))

	if length == 0 {
		err = fmt.Errorf("length is 0")
		return 0, nil, nil, err
	}

	// Read nodes
	nodes = make([]uint16, length)
	for i := 0; i < length; i++ {
		var nodeBytes [2]byte
		if _, err := io.ReadFull(r, nodeBytes[:]); err != nil {
			return 0, nil, nil, err
		}
		nodes[i] = binary.BigEndian.Uint16(nodeBytes[:])
	}

	// Read colors
	colorLength := length / 2
	colors = make([]byte, colorLength)
	if _, err := io.ReadFull(r, colors); err != nil {
		return 0, nil, nil, err
	}

	return length, nodes, colors, nil
}

const (
	width          = 1920 // TODO: this needs to be changed based on video resolution
	height         = 1080
	framesPerScene = 2 // TODO: this needs to be changed to framerate
)

var (
	tris []triangles
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
}

func Play(folderPath string) {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	var err error
	tris, err = readData(folderPath)

	if err != nil {
		fmt.Printf("Error reading data: %v\n", err)
		return
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, folderPath, nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	program := initOpenGL()

	vao, vbo := setupBuffers()

	currentScene := 0
	frameCount := 0

	for !window.ShouldClose() {
		draw(window, program, vao, vbo, currentScene)

		// TODO: this needs to be changed to a ms check with time
		frameCount++
		if frameCount >= framesPerScene {
			frameCount = 0
			currentScene = (currentScene + 1) % len(tris)
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
	gl.BufferData(gl.ARRAY_BUFFER, 2*len(tris[scene].points), gl.Ptr(tris[scene].points), gl.STATIC_DRAW)

	colorUniform := gl.GetUniformLocation(program, gl.Str("uColor\x00"))

	// TODO: this is slow and shouldn't be here
	// If you cant hit frame rate, move out of draw loop and into setup loop
	for i := 0; i*3+2 < len(tris[scene].colors); i++ {
		color := []float32{
			float32(tris[scene].colors[i*3]) / 255.0,
			float32(tris[scene].colors[i*3+1]) / 255.0,
			float32(tris[scene].colors[i*3+2]) / 255.0,
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

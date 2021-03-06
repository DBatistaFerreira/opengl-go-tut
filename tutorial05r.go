/*
Third tutorial in opengl-tutorials.org.

Matrix transformations.
*/
package main

import (
	"fmt"
	gl "github.com/chsc/gogl/gl33"
	"github.com/go-gl/glfw"
	"github.com/jragonmiris/mathgl"
	"math"
	"os"
	"runtime"
	"unsafe"
)

const (
	Title  = "Tutorial 05"
	Width  = 800
	Height = 600
)

const (
	VertexFile   = "shaders/cube_texture.vertexshader"
	FragmentFile = "shaders/cube_texture.fragmentshader"
	TextureFile  = "art/raine.tga"
)

func loadTGA(imagePath string) gl.Uint {
	// Create one OpenGL texture
	var txid gl.Uint
	gl.GenTextures(1, &txid)

	// "Bind" the newly created texture: all future functions will modify this texture
	gl.BindTexture(gl.TEXTURE_2D, txid)

	// Read the file, call glTexImage2d with the right parameters
	glfw.LoadTexture2D(imagePath, 0)

	// Nice trilinear filtering.
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.GenerateMipmap(gl.TEXTURE_2D)

	// Return the ID of the texture we just created
	return txid
}

func xForm(data []gl.Float, xform mathgl.Mat4f) {
	// Apply the provided transformation matrix to all vertices in the 
	// provided data.  
	vertexCount := (int)(len(data) / 3)
	for i := 0; i < vertexCount*3; i += 3 {
		var V = mathgl.Vec4f{
			(float32)(data[i]),
			(float32)(data[i+1]),
			(float32)(data[i+2]),
			1}
		V = xform.Mul4x1(V)
		data[i] = (gl.Float)(V[0])
		data[i+1] = (gl.Float)(V[1])
		data[i+2] = (gl.Float)(V[2])
	}
}

func main() {
	runtime.LockOSThread()
	// Always call init first
	if err := glfw.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "glfw: %s\n", err)
		return
	}

	// Set Window hints - necessary information before we can
	// call the underlying OpenGL context.
	glfw.OpenWindowHint(glfw.FsaaSamples, 4)        // 4x antialiasing
	glfw.OpenWindowHint(glfw.OpenGLVersionMajor, 3) // OpenGL 3.3
	glfw.OpenWindowHint(glfw.OpenGLVersionMinor, 2)
	// We want the new OpenGL
	glfw.OpenWindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

	// Open a window and initialize its OpenGL context
	if err := glfw.OpenWindow(Width, Height, 0, 0, 0, 0, 32, 0, glfw.Windowed); err != nil {
		fmt.Fprintf(os.Stderr, "glfw: %s\n")
		return
	}
	defer glfw.Terminate() // Make sure this gets called if we crash.

	// Set the Window title
	glfw.SetWindowTitle(Title)

	// Make sure we can capture the escape key
	glfw.Enable(glfw.StickyKeys)

	// Initialize OpenGL, make sure we terminate before leaving.
	gl.Init()

	// Dark blue background
	gl.ClearColor(0.0, 0.0, 0.01, 0.0)

	// Load Shaders
	var programID gl.Uint = LoadShaders(
		VertexFile,
		FragmentFile)
	gl.ValidateProgram(programID)
	var validationErr gl.Int
	gl.GetProgramiv(programID, gl.VALIDATE_STATUS, &validationErr)
	if validationErr == gl.FALSE {
		fmt.Fprintf(os.Stderr, "Shader program failed validation!\n")
	}

	// Initialize our Vertex arrays, needed for OpenGL 3+ 
	var vertexArrayID gl.Uint = 0
	gl.GenVertexArrays(1, &vertexArrayID)
	gl.BindVertexArray(vertexArrayID)
	defer gl.DeleteVertexArrays(1, &vertexArrayID) // Make sure this gets called before main is done

	// Get a handle for our "MVP" uniform
	matrixID := gl.GetUniformLocation(programID, gl.GLString("MVP"))

	// Projection matrix: 45° Field of View, 4:3 ratio, display range : 0.1 unit <-> 100 units
	projection := mathgl.Perspective(45.0, 4.0/3.0, 0.1, 100.0)

	// Camera matrix
	view := mathgl.LookAt(
		0.0, 0.0, -6.0,
		0.0, 0.0, 0.0,
		0.0, 1.0, 0.0)

	// Model matrix: and identity matrix (model will be at the origin)
	model := mathgl.Ident4f() // Changes for each model!

	// Our ModelViewProjection : multiplication of our 3 matrices - remember, matrix mult is other way around
	MVP := projection.Mul4(view).Mul4(model) // projection * view * model

	// An array of 3 vectors which represents 3 vertices of a triangle
	/*vertexBufferData2 := [9]gl.Float{	// N.B. We can't use []gl.Float, as that is a slice
		-1.0, -1.0, 0.0,				// We always want to use raw arrays when passing pointers
		1.0, -1.0, 0.0,					// to OpenGL
		0.0, 1.0, 0.0,
	}*/

	// Load the texture
	texture := loadTGA(TextureFile)
	var textureID gl.Int = gl.GetUniformLocation(programID, gl.GLString("myTextureSampler"))

	// Three consecutive floats give a single 3D vertex
	// A cube has 6 faces with 2 triangles each, so this makes 6*2 = 12 triangles,
	// and 12 * 3 vertices
	vertexBufferData1 := []gl.Float{
		-1, -1, -1, // face 1
		1, -1, 1,
		-1, -1, 1,
		-1, -1, -1,
		1, -1, 1,
		1, -1, -1,
		1, -1, -1, // face 2
		1, 1, 1,
		1, -1, 1,
		1, -1, -1,
		1, 1, 1,
		1, 1, -1,
		1, 1, -1, // face 3
		-1, 1, 1,
		1, 1, 1,
		1, 1, -1,
		-1, 1, 1,
		-1, 1, -1,
		-1, 1, -1, // face 4
		-1, -1, 1,
		-1, 1, 1,
		-1, 1, -1,
		-1, -1, 1,
		-1, -1, -1,
		1, -1, -1, // face 5
		-1, 1, -1,
		1, 1, -1,
		1, -1, -1,
		-1, 1, -1,
		-1, -1, -1,
		1, 1, 1, // face 6
		-1, -1, 1,
		1, -1, 1,
		1, 1, 1,
		-1, -1, 1,
		-1, 1, 1,
	}

	// Make two copies of the cube.
	vertexBufferData2 := make([]gl.Float, len(vertexBufferData1))
	vertexBufferData3 := make([]gl.Float, len(vertexBufferData1))
	copy(vertexBufferData2, vertexBufferData1)
	copy(vertexBufferData3, vertexBufferData1)

	// Displace the objects to either side of the main cube, along X axis
	var displaceXplus4 = mathgl.Mat4f{
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		3, 0, 0, 1,
	}

	var displaceXminus4 = mathgl.Mat4f{
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		-3, 0, 0, 1,
	}

	xForm(vertexBufferData2, displaceXplus4)
	xForm(vertexBufferData3, displaceXminus4)

	// Create a random number generator to produce colors
	//now := time.Now()
	//rnd := rand.New(rand.NewSource(now.Unix()))

	//var colorBufferData [3*12*3]gl.Float
	//for i := 0; i < 3*12*3; i += 3 {
	//	colorBufferData[i] = (gl.Float)(rnd.Float32())	// red
	//	colorBufferData[i+1] = (gl.Float)(rnd.Float32()) // blue
	//	colorBufferData[i+2] = (gl.Float)(rnd.Float32()) // green
	//}

	// Two UV coordinated for each vertex.  We only need one 
	// texture array for all three cubes.
	uvBufferData := []gl.Float{
		0, 0,
		1, 1,
		0, 1,
		0, 0,
		1, 1,
		1, 0,
		0, 0,
		1, 1,
		0, 1,
		0, 0,
		1, 1,
		1, 0,
		0, 0,
		1, 1,
		0, 1,
		0, 0,
		1, 1,
		1, 0,
		0, 0,
		1, 1,
		0, 1,
		0, 0,
		1, 1,
		1, 0,
		0, 0,
		1, 1,
		0, 1,
		0, 0,
		1, 1,
		1, 0,
		0, 0,
		1, 1,
		0, 1,
		0, 0,
		1, 1,
		1, 0,
	}

	// Create Vertex buffers
	var vertexBuffer gl.Uint                 // id the vertex buffer
	gl.GenBuffers(1, &vertexBuffer)          // Generate 1 buffer, grab the id
	defer gl.DeleteBuffers(1, &vertexBuffer) // Make sure we delete this, no matter what happens

	// Set up the UV buffer
	var uvBuffer gl.Uint
	gl.GenBuffers(1, &uvBuffer)
	defer gl.DeleteBuffers(1, &uvBuffer)

	// Enable Z-buffer
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	// Precalculate the radians we need, as (float32) conversions
	// become tiresome after a while.  We want a full rotation
	// every 6 seconds, which at 60fps works out to about 1 degree
	// per frame, or 2*pi/360 radians.
	cos_theta := (float32)(math.Cos((2 * math.Pi) / 360))
	sin_theta := (float32)(math.Sin((2 * math.Pi) / 360))

	var zRotationMatrix = mathgl.Mat4f{
		cos_theta, sin_theta, 0.0, 0.0,
		-sin_theta, cos_theta, 0.0, 0.0,
		0.0, 0.0, 1.0, 0.0,
		0.0, 0.0, 0.0, 1.0,
	}
	var xRotationMatrix = mathgl.Mat4f{
		1.0, 0.0, 0.0, 0.0,
		0.0, cos_theta, sin_theta, 0.0,
		0.0, -sin_theta, cos_theta, 0.0,
		0.0, 0.0, 0.0, 1.0,
	}
	var yRotationMatrix = mathgl.Mat4f{
		cos_theta, 0.0, -sin_theta, 0.0,
		0.0, 1.0, 0.0, 0.0,
		sin_theta, 0.0, cos_theta, 0.0,
		0.0, 0.0, 0.0, 1.0,
	}

	// Clamp FPS to vertical sync rate
	glfw.SetSwapInterval(1)

	// Main loop - run until it dies, or we find something better
	for (glfw.Key(glfw.KeyEsc) != glfw.KeyPress) &&
		(glfw.WindowParam(glfw.Opened) == 1) {

		// Clear the screen
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Want to use our loaded shaders
		gl.UseProgram(programID)

		// Rotate the cubes.  For each vertex in vertexBufferData, apply
		// the rotation
		xForm(vertexBufferData1, yRotationMatrix)
		xForm(vertexBufferData2, xRotationMatrix)
		xForm(vertexBufferData3, zRotationMatrix)

		// Create Mondo buffers
		tmp := append(vertexBufferData1, vertexBufferData2...)
		mondoVertexData := append(tmp, vertexBufferData3...)
		tmp = append(uvBufferData, uvBufferData...)
		mondoUVData := append(tmp, uvBufferData...)

		// Perform the translation of the camera viewpoint
		// by sending the requested operation to the vertex shader
		//mvpm := [16]gl.Float{0.93, -0.85, -0.68, -0.68, 0.0, 1.77, -0.51, -0.51, -1.24, -0.63, -0.51, -0.51, 0.0, 0.0, 5.65, 5.83}
		gl.UniformMatrix4fv(matrixID, 1, gl.FALSE, (*gl.Float)(&MVP[0]))

		// Draw each of the cubes
		render(vertexBuffer, uvBuffer, mondoVertexData, mondoUVData, textureID, texture)

		// Perform the translation of the camera viewpoint
		// by sending the requested operation to the vertex shader
		//mvpm := [16]gl.Float{0.93, -0.85, -0.68, -0.68, 0.0, 1.77, -0.51, -0.51, -1.24, -0.63, -0.51, -0.51, 0.0, 0.0, 5.65, 5.83}
		gl.UniformMatrix4fv(matrixID, 1, gl.FALSE, (*gl.Float)(&MVP[0]))

		glfw.SwapBuffers()
	}

}

func render(vertexBuffer, uvBuffer gl.Uint, vertexData, uvData []gl.Float, textureID gl.Int, texture gl.Uint) {
	// Draw an object with the given vertex, UV data, texture buffer and texture
	// Buffer the new data
	vBufferLen := unsafe.Sizeof(vertexData[0]) * (uintptr)(len(vertexData))
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(
		gl.ARRAY_BUFFER,
		gl.Sizeiptr(vBufferLen),
		gl.Pointer(&vertexData[0]),
		gl.STATIC_DRAW)

	// Buffer the UV data
	uvBufferLen := unsafe.Sizeof(uvData[0]) * (uintptr)(len(uvData))
	gl.BindBuffer(gl.ARRAY_BUFFER, uvBuffer)
	gl.BufferData(
		gl.ARRAY_BUFFER,
		gl.Sizeiptr(uvBufferLen),
		gl.Pointer(&uvData[0]),
		gl.STATIC_DRAW)

	// texture in Texture Unit 0
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.Uniform1i(textureID, 0)

	// 1st attribute buffer: vertices
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.VertexAttribPointer(
		0,        // Attribute 0. No particular reason for 0, but must match layout in shader
		3,        // size
		gl.FLOAT, // Type
		gl.FALSE, // normalized?
		0,        // stride
		nil)      // array buffer offset

	// 2nd attribute buffer: UVs
	gl.EnableVertexAttribArray(1)
	gl.BindBuffer(gl.ARRAY_BUFFER, uvBuffer)
	gl.VertexAttribPointer(
		1,        // Attribute 1.  Again, no particular reason, but must match layout
		2,        // size
		gl.FLOAT, // Type
		gl.FALSE, // normalized?
		0,
		nil) // array buffer offset

	// Draw the cube!
	gl.DrawArrays(gl.TRIANGLES, 0, gl.Sizei(vBufferLen/3)) // Starting from vertex 0, 3 vertices total -> triangle

	// Unbind 
	gl.DisableVertexAttribArray(0)
	gl.DisableVertexAttribArray(1)
}

package main

import (
	"errors"
	"io/ioutil"
	"log"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/jonas-p/go-shp"
)

const (
	MILLI_SEC_PER_SEC = float32(1000.0)
	DEG_PER_TWO_SEC   = MILLI_SEC_PER_SEC * 360.0
)

/*
 * Global variables
 */

var (
	worldmap = WorldMap{
		data: []float32{
			+0.0, +0.5, +0.0, // Top
			+1.0, +0.0, +0.0, // Red

			+0.5, -0.5, +0.0, // Bottom Right
			+0.0, +1.0, +0.0, // Green

			-0.5, -0.5, +0.0, // Bottom Left
			+0.0, +0.0, +1.0, // Blue
		},
		attributes: positionAttribute + colorAttribute,
		angle:      0.0,
	}

	camX, camY float32
	camZ       float32 = 300

	modelIndex, viewIndex, projectionIndex int32
	vao, positionIndex, colorIndex         uint32
	epoch                                  int64
	texas                                  = StatePolygon{attributes: 1}
)

func realize(glarea *gtk.GLArea) {
	log.Println("realize")

	shapeFiles := []string{
		"data/shapefiles/cb_2018_us_nation_20m/cb_2018_us_nation_20m.shp",
		"data/shapefiles/cb_2018_us_state_20m/cb_2018_us_state_20m.shp",
	}
	shapes := loadShapeFiles(shapeFiles)
	// spew.Dump(shapes)
	_ = shapes

	states = loadStates("data/shapefiles/cb_2018_us_state_20m/cb_2018_us_state_20m.shp")
	// states = loadStates("data/shapefiles/cb_2018_us_state_500k/cb_2018_us_state_500k.shp")
	txp := states["TX"].(*shp.Polygon)
	texas = NewStatePolygon(txp)
	// texas = NewStatePolygonMerc(txp)

	// Make the the GLArea's GdkGLContext the current OpenGL context.
	glarea.MakeCurrent()

	// Initialize OpenGL.
	err := gl.Init()
	errorCheck(err)

	// Initialize shaders.
	err = initShaders()
	errorCheck(err)

	// Initialize buffer.
	vao, err = initBuffer()
	errorCheck(err)

	// Enable depth test.
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	// Bind our callback function to the glarea, the callback with be called 60
	// times in one second.
	glarea.AddTickCallback(update)

	// Set the epoch.
	epoch = glarea.GetFrameClock().GetFrameTime()

	// Log out OpenGL version and window dimensions.
	log.Printf("opengl version %s", gl.GoStr(gl.GetString(gl.VERSION)))
	log.Printf("window - width: %d height: %d", winWidth, winHeight)
}

func unrealize(glarea *gtk.GLArea) {
	log.Println("unrealize")

	// Make sure the area that is being cleaned up is the current context
	// otherwise opengl api calls with panic.
	glarea.MakeCurrent()

	// Clean up all of the allocated data.
	gl.DeleteVertexArrays(1, &vao)
	gl.DeleteProgram(program)
}

func render(glarea *gtk.GLArea) bool {
	// log.Println("render")

	// Enable attribute index 0 as being used.
	gl.EnableVertexAttribArray(0)

	// Set background color (rgba).
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	texas.Draw()

	// Flush the contents of the pipeline.
	gl.Flush()

	return true
}

func initBuffer() (vao uint32, e error) {
	// Allocate, assign, and bind a single Vertex Array Object.
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	// Allocate and assign Vertex Buffer Object.
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	// Bind Vertex Buffer Object as being the active buffer and storing vertex
	// attributes(coordinates and color data).
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	// Copy the vertex data from array to our buffer.
	gl.BufferData(gl.ARRAY_BUFFER, texas.Size(), texas.Data(), gl.STATIC_DRAW)

	// Enable and set the position attribute with the local data.
	gl.EnableVertexAttribArray(positionIndex)
	gl.VertexAttribPointer(positionIndex, 3, gl.FLOAT, false, texas.Stride(),
		gl.PtrOffset(positionOffset))

	// Enable and set the color attribute with the local data.
	gl.EnableVertexAttribArray(colorIndex)
	gl.VertexAttribPointer(colorIndex, 3, gl.FLOAT, false, texas.Stride(),
		gl.PtrOffset(colorOffset))

	// Finished loading data unbind it vao.
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	// No longer need the vbo.
	gl.DeleteBuffers(1, &vbo)

	if vao == 0 {
		e = errors.New("Error initializing buffers.")
	}
	return
}

func initShaders() (err error) {

	fragmentShaderSrc, err := ioutil.ReadFile("shaders/fragment.glsl")
	if err != nil {
		log.Panic("could not open fragment shader")
	}
	fragmentShader = string(fragmentShaderSrc)

	vertexShaderSrc, err := ioutil.ReadFile("shaders/vertex.glsl")
	if err != nil {
		log.Panic("could not open fragment shader")
	}
	vertexShader = string(vertexShaderSrc)

	// Configure the vertex and fragment shaders.
	program, err = newProgram(vertexShader, fragmentShader)

	// Get the location of the "position" and "color" attributes.
	positionIndex = uint32(gl.GetAttribLocation(program, gl.Str("position\x00")))
	colorIndex = uint32(gl.GetAttribLocation(program, gl.Str("color\x00")))

	// Get the location of the "model", "view", "projection" uniforms.
	modelIndex = gl.GetUniformLocation(program, gl.Str("model\x00"))
	viewIndex = gl.GetUniformLocation(program, gl.Str("view\x00"))
	projectionIndex = gl.GetUniformLocation(program, gl.Str("projection\x00"))
	return
}

func update(widget *gtk.Widget, frameClock *gdk.FrameClock) bool {
	// Queue up the re-rendering of the GLArea.
	widget.QueueDraw()
	return true
}

func onButtonPress(widget *gtk.GLArea, ev *gdk.Event) bool {
	log.Println(ev)
	return true
}

func onKeyPress(widget *gtk.GLArea, ev *gdk.Event) bool {
	log.Println(ev)
	return true
}

const scrollInc = float32(2.0)

func onScroll(widget *gtk.GLArea, e *gdk.Event) bool {
	ev := gdk.EventScroll{Event: e}
	dy := ev.DeltaY()
	if dy == float64(1.0) {
		camZ += scrollInc

	} else if dy == float64(-1.0) {
		camZ -= scrollInc
	}

	if camZ <= 0 {
		camZ = 0
	}

	return true
}

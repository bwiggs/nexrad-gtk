package main

import (
	"log"
	"unsafe"

	"github.com/StefanSchroeder/Golang-Ellipsoid/ellipsoid"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang/geo/s2"
	"github.com/jonas-p/go-shp"
	"golang.org/x/image/colornames"
)

var geo1 = ellipsoid.Init("WGS84", ellipsoid.Degrees, ellipsoid.Meter, ellipsoid.LongitudeIsSymmetric, ellipsoid.BearingIsSymmetric)
var merc = s2.NewMercatorProjection(float64(180.0))
var statesStrokeColor = colornames.Darkblue

type StatePolygon struct {
	attributes int
	data       []float32
	numVerts   int32
}

func NewStatePolygonECEF(p *shp.Polygon) StatePolygon {
	sp := StatePolygon{attributes: 2, numVerts: p.NumPoints}
	for _, pt := range p.Points {
		x, y, z := geo1.ToECEF(pt.X, pt.Y, 0)
		_ = z
		log.Println(x, y, z)
		sp.data = append(sp.data,
			// x, y, z
			float32(x), float32(y), float32(z),
			// r, g, b
			float32(statesStrokeColor.R), float32(statesStrokeColor.G), float32(statesStrokeColor.B),
		)
	}
	return sp
}

func NewStatePolygonMerc(p *shp.Polygon) StatePolygon {
	sp := StatePolygon{attributes: 2, numVerts: p.NumPoints}
	for _, pt := range p.Points {
		mpt := merc.FromLatLng(s2.LatLngFromDegrees(pt.X, pt.Y))
		sp.data = append(sp.data,
			// x, y, z
			float32(mpt.X), float32(mpt.Y), 0.0,
			// r, g, b
			0.0, 1.0, 0.0,
		)
	}
	return sp
}

func NewStatePolygon(p *shp.Polygon) StatePolygon {
	sp := StatePolygon{attributes: 2, numVerts: p.NumPoints}
	for _, pt := range p.Points {
		sp.data = append(sp.data,
			// x, y, z
			float32(pt.X), float32(pt.Y), 0.0,
			// r, g, b
			1.0, 1.0, 0.0,
		)
	}
	return sp
}

func (s *StatePolygon) Draw() {
	// Load the shader program into the rendering pipeline.
	gl.UseProgram(program)

	s.mvp()

	// Bind to the data in the buffer
	gl.BindVertexArray(vao)

	// Render the data
	gl.DrawArrays(gl.LINE_LOOP, 0, s.numVerts)
	// gl.DrawArrays(gl.POLYGON_MODE, 0, s.numVerts)

	// Done with the buffer and program so unbind them
	gl.BindVertexArray(0)
	gl.UseProgram(0)
}

func (s *StatePolygon) Data() unsafe.Pointer {
	// Return the address of the array containing all of the vertex data.
	return gl.Ptr(s.data)
}

func (s *StatePolygon) Stride() int32 {
	// Return the total number of bytes of data that describes each vertex.
	return int32(3 * s.attributes * floatSize)
}

func (s *StatePolygon) Size() int {
	// Return size of the data in number of bytes.
	return len(s.data) * floatSize
}

// mvp Model View Projection
func (s *StatePolygon) mvp() {
	// Get 4x4 identity matrix for the model's transformations
	model := mgl32.Ident4()

	// Apply the change in angle to the model's set of transformations
	// model = mgl32.HomogRotate3DY(90)

	// Set the handle to point to the address of the model matrix.
	gl.UniformMatrix4fv(modelIndex, 1, false, &model[0])

	// Get 4x4 projection matrix with a 60 degree field of view, an aspect ratio
	// of the window dimensions, near clipping plane, and a far clipping plane.
	projection := mgl32.Perspective(
		mgl32.DegToRad(90.0), float32(winWidth/winHeight), 0.1, -1.0,
	)
	// Set the handle to point to the address of the projection matrix.
	gl.UniformMatrix4fv(projectionIndex, 1, false, &projection[0])

	// Get 4x4 view matrix with an eye position, target position,
	// and the up direction with a positive bias in the y-axis.
	// Right-handed coordinate system.
	view := mgl32.LookAtV(
		mgl32.Vec3{0, 0, camZ}, mgl32.Vec3{camX, camY, 0}, mgl32.Vec3{0, 1, 0},
	)
	// Set the handle to point to the address of the view matrix.
	gl.UniformMatrix4fv(viewIndex, 1, false, &view[0])
}

func loadStates(file string) map[string]shp.Shape {
	// open a shapefile for reading
	states := map[string]shp.Shape{}
	_ = states
	// shape, err := shp.Open("data/cb_2018_us_nation_20m/cb_2018_us_nation_20m.shp")
	shape, err := shp.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer shape.Close()

	// fields from the attribute table (DBF)
	fields := shape.Fields()

	// loop through all features in the shapefile
	for shape.Next() {
		n, p := shape.Shape()

		// print feature
		// fmt.Println(reflect.TypeOf(p).Elem(), p.BBox())

		// print attributes
		for k, f := range fields {
			val := shape.ReadAttribute(n, k)
			// fmt.Printf("\t%v: %v\n", f, val)
			if f.String() == "STUSPS" {
				states[val] = p
			}
		}
	}
	return states
}

func loadShapeFiles(files []string) (shapes []*shp.Shape) {
	for _, file := range files {
		// open a shapefile for reading
		shape, err := shp.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		defer shape.Close()

		// fields := shape.Fields()

		// loop through all features in the shapefile
		for shape.Next() {
			_, p := shape.Shape()
			// log.Println(reflect.TypeOf(p).Elem(), p.BBox())

			shapes = append(shapes, &p)
			// for k, f := range fields {
			// 	val := shape.ReadAttribute(n, k)
			// 	log.Printf("\t%v: %v\n", f, val)
			// }
		}
	}

	return
}

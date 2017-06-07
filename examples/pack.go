package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	. "github.com/fogleman/fauxgl"
	"github.com/nfnt/resize"
)

const (
	scale  = 4    // optional supersampling
	width  = 1600 // output width in pixels
	height = 1600 // output height in pixels
	fovy   = 30   // vertical field of view in degrees
	near   = 1    // near clipping plane
	far    = 10   // far clipping plane
)

var (
	eye        = V(4, 4, 2)                  // camera position
	center     = V(0, 0, 0)                  // view center position
	up         = V(0, 0, 1)                  // up vector
	light      = V(0.25, 0.5, 1).Normalize() // light direction
	color      = HexColor("#FEB41C")         // object color
	background = HexColor("#24221F")         // background color
)

func timed(name string) func() {
	if len(name) > 0 {
		fmt.Printf("%s... ", name)
	}
	start := time.Now()
	return func() {
		fmt.Println(time.Since(start))
	}
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	var done func()

	done = timed("loading mesh")
	mesh, err := LoadSTL(os.Args[1])
	if err != nil {
		panic(err)
	}
	done()

	fmt.Println(len(mesh.Triangles))

	done = timed("transforming mesh")
	// mesh.Transform(Rotate(V(1, 0, 0), Radians(-90)))
	mesh.BiUnitCube()
	mesh.Transform(Scale(V(60, 60, 60)))
	done()

	done = timed("building bvh tree")
	model := NewPackModel()
	model.Add(mesh, 8)
	done()

	fmt.Println(model.Volume())
	fmt.Println(model.Energy())

	const m = 100
	model = Anneal(model, 1e0*m, 1e-5*m, 3000000).(*PackModel)

	fmt.Printf("volume = %g\n", model.Volume())

	mesh = NewEmptyMesh()
	// cubes := NewEmptyMesh()
	for _, item := range model.Items {
		m := item.Mesh.Copy()
		m.Transform(item.Matrix())
		mesh.Add(m)
		// for _, box := range item.Tree.Leaves(-1) {
		// 	m := Translate(Vector{0.5, 0.5, 0.5})
		// 	m = m.Scale(box.Size())
		// 	m = m.Translate(box.Min)
		// 	cube := NewCube()
		// 	cube.Transform(m)
		// 	cubes.Add(cube)
		// }
	}

	done = timed("transforming mesh")
	mesh.BiUnitCube()
	// cubes.BiUnitCube()
	done()

	context := NewContext(width*scale, height*scale)
	context.ClearColorBufferWith(background)

	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)

	shader := NewPhongShader(matrix, light, eye)
	shader.ObjectColor = color
	shader.DiffuseColor = Gray(0.9)
	shader.SpecularColor = Gray(0.25)
	shader.SpecularPower = 100
	context.Shader = shader
	done = timed("rendering mesh")
	context.DrawMesh(mesh)
	done()

	done = timed("downsampling image")
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)
	done()

	done = timed("writing output")
	SavePNG("out.png", image)
	done()

	done = timed("writing mesh")
	mesh.SaveSTL("out.stl")
	done()
}

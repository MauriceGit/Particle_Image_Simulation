package geometry

import (
    "github.com/go-gl/gl/v4.5-core/gl"
    "github.com/go-gl/mathgl/mgl32"
    "fmt"
    "unsafe"
    "image"
    "os"
    //"golang.org/x/image/bmp"
    "image/png"
    "math/rand"
)

type Geometry struct {
    // Important geometry attributes
    ArrayBuffer     uint32
    VertexObject    uint32
    VertexCount     int32
}

type Particle struct {
    OriginalPos       mgl32.Vec4
    Pos               mgl32.Vec4
    StartColor        mgl32.Vec4
    EndColor          mgl32.Vec4
    Accelleration     mgl32.Vec4
    Speed             mgl32.Vec4
}

type Particles struct {
    Positions       []Particle
    GeoAttrib       Geometry
}


func setRenderingAttributes(vertexArrayObject, arrayBuffer, location uint32, size int32, normalized bool, stride int, offset int) {
    // Find the last bindings so we don't overwrite them
    var previousVertexArrayObject int32
    gl.GetIntegerv(gl.VERTEX_ARRAY_BINDING, &previousVertexArrayObject)
    var previousArrayBuffer int32
    gl.GetIntegerv(gl.ARRAY_BUFFER, &previousArrayBuffer)

    // Set our vertex attributes and pointers
    gl.BindVertexArray(vertexArrayObject)
    gl.BindBuffer(gl.ARRAY_BUFFER, arrayBuffer)
    gl.EnableVertexAttribArray(location)
    gl.VertexAttribPointer(location, size, gl.FLOAT, normalized, int32(stride), gl.PtrOffset(offset))

    // Reset the old bindings.
    gl.BindBuffer(gl.ARRAY_BUFFER, uint32(previousArrayBuffer))
    gl.BindVertexArray(uint32(previousVertexArrayObject))
}

func createParticleBuffers(particles []Particle) Geometry {
    var particleGeo = Geometry{}

    particleGeo.VertexCount = int32(len(particles))

    gl.GenBuffers(1, &particleGeo.ArrayBuffer)
    gl.BindBuffer(gl.ARRAY_BUFFER, particleGeo.ArrayBuffer)
    emptyParticle := Particle{}
    gl.BufferData(gl.ARRAY_BUFFER, len(particles)*int(unsafe.Sizeof(emptyParticle)), gl.Ptr(particles), gl.STATIC_DRAW)
    gl.BindBuffer(gl.ARRAY_BUFFER, 0)

    gl.GenVertexArrays(1, &particleGeo.VertexObject)
    gl.BindVertexArray(particleGeo.VertexObject)

    emptyVec := mgl32.Vec4{}

    // We only use a vec3, even though we have a vec4 in the data structure. If we need the w component, we can edit the 3 --> 4
    // and change the compute shader buffer to vec4.
    // But, we HAVE to use a vec4 because of memory layout issues otherwise. The GPU only gets byte chunks the size of a vec4.
    setRenderingAttributes(particleGeo.VertexObject, particleGeo.ArrayBuffer, 0, 4, false, int(unsafe.Sizeof(emptyParticle)), 1*int(unsafe.Sizeof(emptyVec)))
    setRenderingAttributes(particleGeo.VertexObject, particleGeo.ArrayBuffer, 1, 4, false, int(unsafe.Sizeof(emptyParticle)), 2*int(unsafe.Sizeof(emptyVec)))

    return particleGeo
}

func CreateParticles(rows, cols, maxWidth, maxHeight int) Particles {

    var widthOffset   float32 = -float32(maxWidth) / 2.0
    var heightOffset  float32 = -float32(maxHeight) / 2.0
    var widthFactor  float32 = float32(maxWidth) / float32(rows)
    var heightFactor float32 = float32(maxHeight) / float32(cols)

    particles := Particles{}
    positions := make([]Particle, rows*cols)



    image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)

    fImg, err := os.Open("testImage.png")
    if err != nil {
        fmt.Printf("Error while trying to load image testImage.bmp: %v.\n", err)
        return particles
    }
    image, _, err := image.Decode(fImg)
    if err != nil {
        fmt.Printf("Error while trying to decode image testImage.bmp: %v.\n", err)
        return particles
    }

    imageRangeX := image.Bounds().Max.X - image.Bounds().Min.X
    imageRangeY := image.Bounds().Max.Y - image.Bounds().Min.Y

    scaleFactorX := float32(imageRangeX / cols)
    scaleFactorY := float32(imageRangeY / rows)

    rand.Seed(42)

    for i := 0; i < cols*rows; i++ {

        x := float32(i % rows) * widthFactor  + widthOffset
        y := float32(i / cols) * heightFactor + heightOffset
        z := float32(-10.0)

        coordX := int((x-widthOffset)*scaleFactorX/widthFactor)
        coordY := int((y-heightOffset)*scaleFactorY/heightFactor)

        r,g,b,a := image.At(coordX, imageRangeY-coordY).RGBA()
        nr  := float32(r/257)
        ng  := float32(g/257)
        nb  := float32(b/257)
        na  := float32(a/257)

        randAccX := float32(rand.Float64()/ 1000.0)-float32(rand.Float64()/ 1000.0)
        randAccY := float32(rand.Float64()/ 100.0)
        //randAccZ := float32(rand.Float64()/ 1000.0)-float32(rand.Float64()/ 1000.0)

        positions[i] = Particle{
            OriginalPos:      mgl32.Vec4{x, y, z, 0},
            Pos:              mgl32.Vec4{x, y, z, 0},
            StartColor:       mgl32.Vec4{nr, ng, nb, na},
            EndColor:         mgl32.Vec4{0,0,0,1},
            Accelleration:    mgl32.Vec4{randAccX/100.0,-randAccY/100.0,0.0,0},
            Speed:            mgl32.Vec4{0,0,0,0},
        }
    }

    particles.Positions = positions
    particles.GeoAttrib = createParticleBuffers(positions)

    return particles

}




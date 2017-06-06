package geometry

import (
    . "ParticleImage/Image"
    "github.com/go-gl/gl/v4.5-core/gl"
    "github.com/go-gl/mathgl/mgl32"
    "fmt"
    "unsafe"
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
    setRenderingAttributes(particleGeo.VertexObject, particleGeo.ArrayBuffer, 2, 4, false, int(unsafe.Sizeof(emptyParticle)), 3*int(unsafe.Sizeof(emptyVec)))

    return particleGeo
}

func CreateParticles(rows, cols, maxWidth, maxHeight int) Particles {

    var widthOffset   float32 = -float32(maxWidth) / 2.0
    var heightOffset  float32 = -float32(maxHeight) / 2.0
    var widthFactor  float32 = float32(maxWidth) / float32(rows)
    var heightFactor float32 = float32(maxHeight) / float32(cols)

    particles := Particles{}
    positions := make([]Particle, rows*cols)

    // Load the end image.
    startImage, err := LoadImage("image2.png")
    if err != nil {
        fmt.Printf("Error: %v.\n", err)
        return particles
    }
    startScaleFactorX := float32(startImage.RangeX() / cols)
    startScaleFactorY := float32(startImage.RangeY() / rows)

    endImage, err := LoadImage("image1.png")
    if err != nil {
        fmt.Printf("Error: %v.\n", err)
        return particles
    }
    endScaleFactorX := float32(endImage.RangeX() / cols)
    endScaleFactorY := float32(endImage.RangeY() / rows)

    rand.Seed(42)

    for i := 0; i < cols*rows; i++ {

        x := float32(i % rows) * widthFactor  + widthOffset
        y := float32(i / cols) * heightFactor + heightOffset
        z := float32(-10.0)

        startCoordX := int((x-widthOffset)*startScaleFactorX/widthFactor)
        startCoordY := int((y-heightOffset)*startScaleFactorY/heightFactor)

        sr,sg,sb,sa := startImage.RGBAAt(startCoordX, startCoordY, true)

        endCoordX := int((x-widthOffset)*endScaleFactorX/widthFactor)
        endCoordY := int((y-heightOffset)*endScaleFactorY/heightFactor)

        er, eg, eb, ea := endImage.RGBAAt(endCoordX, endCoordY, true)

        //randAccX := float32(rand.Float64()/ 1000.0)-float32(rand.Float64()/ 1000.0)
        randAccY := float32(rand.Float64() * 1.0)

        positions[i] = Particle{
            OriginalPos:      mgl32.Vec4{x, y, z, 0},
            Pos:              mgl32.Vec4{x, y, z, 0},
            StartColor:       mgl32.Vec4{sr, sg, sb, sa},
            EndColor:         mgl32.Vec4{er, eg, eb, ea},
            Accelleration:    mgl32.Vec4{0.0,-randAccY,0.0,0},
            Speed:            mgl32.Vec4{0,0,0,0},
        }
    }

    particles.Positions = positions
    particles.GeoAttrib = createParticleBuffers(positions)

    return particles

}




package main

import (
    . "ParticleImage/Geometry"
    "strings"
    "runtime"
    "github.com/go-gl/mathgl/mgl32"
    "fmt"
    "github.com/go-gl/gl/v4.5-core/gl"
    "github.com/go-gl/glfw/v3.2/glfw"
    //"io/ioutil"
    //"math"
    "bytes"
    "os"
    "io"
)

// Constants and global variables

const g_WindowWidth  = 1000
const g_WindowHeight = 1000

const g_ParticleCountX = 1000
const g_ParticleCountY = 1000

const g_WindowTitle  = "Particle Image"
var g_ShaderID uint32
var g_ComputeShaderID uint32

// Rendering perspective parameter. Not changable right now. But very easy to implement.
var g_cameraPos = mgl32.Vec3{0, 0, 1}
var g_center    = mgl32.Vec3{0,0,0}
var g_up        = mgl32.Vec3{0,1,0}
var g_fovy      = mgl32.DegToRad(90.0)
var g_aspect    = float32(g_WindowWidth)/g_WindowHeight
var g_nearPlane = float32(0.1)
var g_farPlane  = float32(1000.0)

var g_viewMatrix          mgl32.Mat4

var g_interval float32 = 0.0
var g_lastCallTime float64 = 0.0
var g_frameCount int = 0
var g_fps float32 = 60.0

var g_particles Particles

var g_collapseImage int32 = 0

var g_colorInterpolation float32 = 0.0;

func init() {
    // GLFW event handling must run on the main OS thread
    runtime.LockOSThread()
}


func printHelp() {
    fmt.Println(
`Info:  Usage for the sample implementation of Solar rack visualization.
        All Changes and parameters refer to the solar rack that is focused!

h/H             Prints this help output
q/Q/ESC         Exits this simulation
PageUp/PageDown Changes the focus to another Solar rack (arra)
r               Rotates the rack -10° around the Y-Axis
R               Rotates the rack +10° around the Y-Axis
d               Decreases the ground distance by 1
D               Increases the ground distance by 1
a               Decreases the pitch angle by 10°
A               Increases the pitch angle by 10°
LEFT            Increases the number of cols by 1
RIGHT           Decreases the number of cols by 1
UP              Increases the number of rows by 1
DOWN            Decreases the number of rows by 1`,
    )
}

// Set OpenGL version, profile and compatibility
func initGraphicContext() (*glfw.Window, error) {
    glfw.WindowHint(glfw.Resizable, glfw.True)
    glfw.WindowHint(glfw.ContextVersionMajor, 4)
    glfw.WindowHint(glfw.ContextVersionMinor, 3)
    glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
    glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

    window, err := glfw.CreateWindow(g_WindowWidth, g_WindowHeight, g_WindowTitle, nil, nil)
    if err != nil {
        return nil, err
    }
    window.MakeContextCurrent()

    // Initialize Glow
    if err := gl.Init(); err != nil {
        return nil, err
    }

    return window, nil
}

// Taken from Sample. Not changed at all.
func compileShader(source string, shaderType uint32) (uint32, error) {
    shader := gl.CreateShader(shaderType)

    csources, free := gl.Strs(source)
    var csourceslength int32 = int32(len(source))
    gl.ShaderSource(shader, 1, csources, &csourceslength)
    free()
    gl.CompileShader(shader)

    var status int32
    gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
    if status == gl.FALSE {
        var logLength int32
        gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

        log := strings.Repeat("\x00", int(logLength+1))
        gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

        return 0, fmt.Errorf("failed to compile %v: %v", source, log)
    }

    return shader, nil
}

func readFile(name string) (string, error) {

    buf := bytes.NewBuffer(nil)
    f, err := os.Open(name)
    if err != nil {
        return "", err
    }
    io.Copy(buf, f)
    f.Close()

    return string(buf.Bytes()), nil
}

// Mostly taken from the Demo. But compiling and linking shaders
// just should be done like this anyways.
func newProgram(vertexShaderName, fragmentShaderName string) (uint32, error) {

    vertexShaderSource, err := readFile(vertexShaderName)
    if err != nil {
        return 0, err
    }

    fragmentShaderSource, err := readFile(fragmentShaderName)
    if err != nil {
        return 0, err
    }

    vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
    if err != nil {
        return 0, err
    }

    fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
    if err != nil {
        return 0, err
    }

    program := gl.CreateProgram()

    gl.AttachShader(program, vertexShader)
    gl.AttachShader(program, fragmentShader)
    gl.LinkProgram(program)

    var status int32
    gl.GetProgramiv(program, gl.LINK_STATUS, &status)
    if status == gl.FALSE {
        var logLength int32
        gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

        log := strings.Repeat("\x00", int(logLength+1))
        gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

        return 0, fmt.Errorf("failed to link program: %v", log)
    }

    gl.DeleteShader(vertexShader)
    gl.DeleteShader(fragmentShader)

    return program, nil
}

func newComputeProgram(computeShaderName string) (uint32, error) {

    computeShaderSource, err := readFile(computeShaderName)
    if err != nil {
        return 0, err
    }
    computeShader, err := compileShader(computeShaderSource, gl.COMPUTE_SHADER)
    if err != nil {
        return 0, err
    }
    program := gl.CreateProgram()

    gl.AttachShader(program, computeShader)
    gl.LinkProgram(program)

    var status int32
    gl.GetProgramiv(program, gl.LINK_STATUS, &status)
    if status == gl.FALSE {
        var logLength int32
        gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

        log := strings.Repeat("\x00", int(logLength+1))
        gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

        return 0, fmt.Errorf("failed to link program: %v", log)
    }

    gl.DeleteShader(computeShader)

    return program, nil
}

// Defines the Model-View-Projection matrices for the shader.
func defineMatrices(shader uint32) {
    // Initialize and set projection matrix to be used in the vertex shader later.
    // With aspect ratio and near/far plane.
    //projection := mgl32.Perspective(g_fovy, g_aspect, g_nearPlane, g_farPlane)

    projection := mgl32.Ortho(-50, 50, -50, 50, g_nearPlane, g_farPlane)


    projectionUniform := gl.GetUniformLocation(shader, gl.Str("projectionMat\x00"))
    // Set as uniform, to be transfered to the GPU
    gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

    // Equivalent of gluLookAt, to set where the eye, center, up vectors are.
    camera := mgl32.LookAtV(g_cameraPos, g_center, g_up)
    cameraUniform := gl.GetUniformLocation(shader, gl.Str("cameraMat\x00"))
    gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])

    // No model specific transformation. Already in world-coordinates.
    // Still calculated for the completeness of it and later extensibility.
    // Can be altered.
    model := mgl32.Ident4()
    modelUniform := gl.GetUniformLocation(shader, gl.Str("modelMat\x00"))
    gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])
}

// This renders a Solar Rack. Completely independent on any concrete parameter.
// As no memory is changed, changes in Solar-Rack sizes, ... are very cheap!
func renderParticles(shader uint32, color mgl32.Vec3) {

    // Which Shader to use.
    gl.UseProgram(shader)

    // This is limited to Nvidia and Desktop applications!!!
    //gl.Enable(gl.PROGRAM_POINT_SIZE)
    //gl.PointSize(5)

    // Set initial Matrices for the Shader.
    defineMatrices(shader)

    gl.BindVertexArray(g_particles.GeoAttrib.VertexObject)

    gl.Uniform1fv(gl.GetUniformLocation(shader, gl.Str("colorInterpolation\x00")), 1, &g_colorInterpolation)

    // And draw one panel.
    gl.DrawArrays(gl.POINTS, 0, g_particles.GeoAttrib.VertexCount)

    gl.BindVertexArray(0)
    gl.UseProgram(0)
}

func recalculateParticles() {

    gl.UseProgram(g_ComputeShaderID)

    gl.Uniform1fv(gl.GetUniformLocation(g_ComputeShaderID, gl.Str("dt\x00")), 1, &g_interval)
    gl.Uniform1iv(gl.GetUniformLocation(g_ComputeShaderID, gl.Str("collapse\x00")), 1, &g_collapseImage)
    gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 0, g_particles.GeoAttrib.ArrayBuffer)

    gl.DispatchCompute(g_ParticleCountX*g_ParticleCountY/240+1, 1, 1)
    gl.MemoryBarrier(gl.ALL_BARRIER_BITS)

    gl.UseProgram(0)
}

func render(window *glfw.Window) {
    gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
    gl.Enable(gl.DEPTH_TEST)
    // Nice blueish background
    //gl.ClearColor(135.0/255.,206.0/255.,235.0/255., 1.0)
    //gl.ClearColor(1,1,1, 1.0)
    gl.ClearColor(0,0,0, 1.0)

    gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

    gl.Enable(gl.BLEND);
    gl.BlendFunc(gl.ONE, gl.ONE);

    gl.Viewport(0, 0, g_WindowWidth, g_WindowHeight)

    recalculateParticles()

    color := mgl32.Vec3{1, 0, 0}
    renderParticles(g_ShaderID, color)

}

// Callback method for a keyboard press
func cbKeyboard(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

    // All changes come VERY easy now.
    if action == glfw.Press {
        switch key {
            // Close the Simulation.
            case glfw.KeyEscape, glfw.KeyQ:
                window.SetShouldClose(true)
            case glfw.KeyH:
                printHelp()
            case glfw.KeySpace:
                if g_collapseImage == 0 {
                    g_collapseImage = 1
                } else {
                    g_collapseImage = 0
                }
        }
    }

}


// Register all needed callbacks
func registerCallBacks (window *glfw.Window) {

    window.SetKeyCallback(cbKeyboard)
}


func displayFPS(window *glfw.Window) {
    currentTime := glfw.GetTime()
    g_interval = float32(currentTime - g_lastCallTime)


    if g_collapseImage == 1 {
        if g_colorInterpolation < 1.0 {
            g_colorInterpolation += g_interval/5.0
        }
    }

    if g_frameCount%60 == 0 {
        g_fps = float32(1.0) / g_interval

        s := fmt.Sprintf("FPS: %.2f", g_fps)
        window.SetTitle(s)
    }

    g_lastCallTime = currentTime
    g_frameCount += 1

}

// Mainloop for graphics updates and object animation
func mainLoop (window *glfw.Window) {

    registerCallBacks(window)

    for !window.ShouldClose() {

        displayFPS(window)

        // This actually renders everything.
        render(window)

        window.SwapBuffers()
        glfw.PollEvents()
    }

}

func main() {
    var err error = nil
    if err = glfw.Init(); err != nil {
        panic(err)
    }
    // Terminate as soon, as this the function is finished.
    defer glfw.Terminate()

    window, err := initGraphicContext()
    if err != nil {
        // Decision to panic or do something different is taken in the main
        // method and not in sub-functions
        panic(err)
    }

    path := "../Go/src/ParticleImage/"
    g_ShaderID, err = newProgram(path+"vertexShader.vert", path+"fragmentShader.frag")
    if err != nil {
        panic(err)
    }
    g_ComputeShaderID, err = newComputeProgram(path+"computeShader.comp")
    if err != nil {
        panic(err)
    }

    g_particles = CreateParticles(g_ParticleCountX, g_ParticleCountY, 100, 100)

    mainLoop(window)

}





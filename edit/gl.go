package edit

import (
	"fmt"
	"image"
	"math"
	"runtime"
	"strings"
	"sync"

	"github.com/frizinak/phodo/img48"
	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	fs       = 4
	stride   = 4
	vertices = 4
)

type points [stride * vertices]float32

func buf(d *points, x0, y0, x1, y1 float32) {
	d[0] = x1
	d[1] = y1
	d[4] = x1
	d[5] = y0
	d[8] = x0
	d[9] = y0
	d[12] = x0
	d[13] = y1
	d[2], d[3] = 1, 1
	d[6], d[7] = 1, 0
	d[10], d[11] = 0, 0
	d[14], d[15] = 0, 1
}

func imgTexture(img *img48.Img) (uint32, error) {
	b := img.Bounds()
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	const m = gl.NEAREST
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, m)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, m)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGB,
		int32(b.Dx()),
		int32(b.Dy()),
		0,
		gl.RGB,
		gl.UNSIGNED_SHORT,
		gl.Ptr(img.Pix),
	)

	return texture, nil
}

func releaseTexture(tex uint32) error {
	gl.DeleteTextures(1, &tex)
	return nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	l := int32(len(source))
	gl.ShaderSource(shader, 1, csources, &l)
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

func newProgram() (uint32, error) {
	vertexShaderSrc := `#version 410 core
layout (location = 0) in vec2 pos;
layout (location = 1) in vec2 tex;
out vec2 TexCoord;
out vec3 FragPos;

uniform mat4 projection;
uniform mat4 model;

void main()
{
    FragPos = vec3(model * vec4(pos, 0.0, 1.0));
    gl_Position = projection * model * vec4(pos, 0.0, 1.0);
    TexCoord = tex;
}`

	fragShaderSrc := `#version 410 core
out vec4 color;
in vec2 TexCoord;
in vec3 FragPos;

uniform sampler2D texture1;
uniform mat4 projection;

void main()
{
    color = texture(texture1, TexCoord);
}`
	vertexShader, err := compileShader(vertexShaderSrc, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragShaderSrc, gl.FRAGMENT_SHADER)
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

func initialize() (*glfw.Window, error) {
	if err := glfw.Init(); err != nil {
		return nil, err
	}
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.DoubleBuffer, 1)
	window, err := glfw.CreateWindow(
		800,
		800,
		"phodo edit",
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	window.MakeContextCurrent()
	return window, nil
}

type Viewer struct {
	window                             *glfw.Window
	monitor                            *glfw.Monitor
	videoMode                          *glfw.VidMode
	windowX, windowY, windowW, windowH int

	sem   sync.Mutex
	img   *img48.Img
	inval bool

	realWidth, realHeight int

	proj mgl32.Mat4

	cursor struct {
		x, y float64
		down bool
	}

	pos struct {
		x, y       int
		maxX, maxY int
		scale      float64
	}

	onkey   func(rune)
	onclick func(x, y int)
}

func (v *Viewer) onCursor(w *glfw.Window, x, y float64) {
	v.cursor.x, v.cursor.y = x, y
	v.reportCursor()
}

func (v *Viewer) onClick(w *glfw.Window, btn glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	v.cursor.down = btn == glfw.MouseButton1 && action == glfw.Press
	if btn != glfw.MouseButton1 || action != glfw.Press || v.pos.scale == 0 {
		return
	}

	v.reportCursor()
}

func (v *Viewer) reportCursor() {
	if !v.cursor.down || v.pos.scale == 0 {
		return
	}

	rx := v.pos.scale * (v.cursor.x - float64(v.pos.x))
	ry := v.pos.scale * (v.cursor.y - float64(v.pos.y))
	x := int(math.Round(rx))
	y := int(math.Round(ry))
	if x > 0 && y > 0 && x <= v.pos.maxX && y < v.pos.maxY {
		v.onclick(x, y)
	}
}

func (v *Viewer) onText(w *glfw.Window, char rune) {
	if v.onkey != nil {
		v.onkey(char)
	}
}

func (v *Viewer) onResize(wnd *glfw.Window, width, height int) {
	v.realWidth, v.realHeight = width, height
	gl.Viewport(0, 0, int32(width), int32(height))
	v.proj = mgl32.Ortho2D(0, float32(width), float32(height), 0)
	v.windowW, v.windowH = width, height
}

func (v *Viewer) onPos(wnd *glfw.Window, x, y int) {
	v.windowX, v.windowY = x, y
}

func (v *Viewer) run() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var err error
	v.window, err = initialize()
	defer glfw.Terminate()
	if err != nil {
		return err
	}
	v.monitor = glfw.GetPrimaryMonitor()
	v.videoMode = v.monitor.GetVideoMode()
	v.windowX, v.windowY = v.window.GetPos()
	v.windowW, v.windowH = v.window.GetSize()
	v.proj = mgl32.Ortho2D(0, 800, 800, 0)

	if err := gl.Init(); err != nil {
		return err
	}

	v.window.SetFramebufferSizeCallback(v.onResize)
	v.window.SetPosCallback(v.onPos)
	v.window.SetCharCallback(v.onText)
	if v.onclick != nil {
		v.window.SetCursorPosCallback(v.onCursor)
		v.window.SetMouseButtonCallback(v.onClick)
	}

	w, h := v.window.GetFramebufferSize()
	v.onResize(v.window, w, h)

	program, err := newProgram()
	if err != nil {
		return err
	}
	gl.UseProgram(program)
	gl.Enable(gl.TEXTURE_2D)

	var tex uint32 = 0
	model := mgl32.Ident4()

	lastProjection := mgl32.Ident4()
	var lastTex uint32 = 0

	modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))
	projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))

	var ebo uint32
	indices := []uint32{0, 1, 3, 1, 2, 3}
	gl.GenBuffers(1, &ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, 6*fs, gl.Ptr(indices), gl.STATIC_DRAW)

	var vao, vbo uint32
	var obounds image.Point
	updateVAO := func(bounds image.Point) {
		first := vao == 0
		if !first && obounds == bounds {
			return
		}
		obounds = bounds
		d := points{}
		buf(&d, 0, 0, float32(bounds.X), float32(bounds.Y))
		if first {
			gl.GenVertexArrays(1, &vao)
			gl.GenBuffers(1, &vbo)
			gl.BindVertexArray(vao)
			gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
		}

		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.BufferData(gl.ARRAY_BUFFER, stride*vertices*fs, gl.Ptr(&d[0]), gl.DYNAMIC_DRAW)
		if !first {
			gl.BindBuffer(gl.ARRAY_BUFFER, 0)
			return
		}

		gl.EnableVertexAttribArray(0)
		gl.VertexAttribPointer(0, 2, gl.FLOAT, false, stride*fs, gl.PtrOffset(0))
		gl.EnableVertexAttribArray(1)
		gl.VertexAttribPointer(1, 2, gl.FLOAT, false, stride*fs, gl.PtrOffset(2*fs))

		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
		gl.BindVertexArray(0)
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
	}

	update := func() (image.Point, error) {
		v.sem.Lock()
		defer v.sem.Unlock()
		var bounds image.Point
		if v.img != nil {
			bounds = image.Point{v.img.Rect.Dx(), v.img.Rect.Dy()}
			if bounds.X != 0 && bounds.Y != 0 {
				rat := float64(bounds.X) / float64(bounds.Y)
				rw, rh := v.realWidth, v.realHeight
				if rw > bounds.X {
					rw = bounds.X
				}
				if rh > bounds.Y {
					rh = bounds.Y
				}
				bounds.X, bounds.Y = rw, int(float64(rw)/rat)
				if float64(rh)/float64(bounds.Y) < float64(rw)/float64(bounds.X) {
					bounds.X, bounds.Y = int(float64(rh)*rat), rh
				}
			}

			updateVAO(bounds)
		}

		if !v.inval || v.img == nil {
			return bounds, nil
		}

		v.inval = false
		if tex != 0 {
			if err := releaseTexture(tex); err != nil {
				return bounds, err
			}
		}

		var err error
		tex, err = imgTexture(v.img)
		if err != nil {
			return bounds, err
		}

		return bounds, err
	}

	var lastBounds image.Point
	frame := func() error {
		bounds, err := update()
		if err != nil {
			return err
		}
		if tex == 0 {
			return nil
		}
		recenter := false
		if tex != lastTex {
			lastTex = tex
			gl.BindTexture(gl.TEXTURE_2D, tex)
			gl.BindVertexArray(vao)
		}

		if v.proj != lastProjection {
			gl.UniformMatrix4fv(projectionUniform, 1, false, &v.proj[0])
			lastProjection = v.proj
			recenter = true
		}

		if bounds != lastBounds {
			lastBounds = bounds
			recenter = true
		}

		if recenter {
			v.pos.x = v.realWidth/2 - lastBounds.X/2
			v.pos.y = v.realHeight/2 - lastBounds.Y/2
			v.pos.scale = 0
			if bounds.X != 0 && v.img != nil {
				v.pos.maxX = v.img.Rect.Dx()
				v.pos.maxY = v.img.Rect.Dy()
				v.pos.scale = float64(v.pos.maxX) / float64(bounds.X)
			}
			model = mgl32.Translate3D(float32(v.pos.x), float32(v.pos.y), 0)
			gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])
		}

		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, gl.PtrOffset(0))
		return nil
	}

	for !v.window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)
		if err = frame(); err != nil {
			return err
		}
		v.window.SwapBuffers()
		glfw.PollEvents()
	}

	return nil
}

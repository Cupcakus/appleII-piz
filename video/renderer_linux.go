package video

/* renderer_linux.go -- Raspberry PI-Zero framebuffer renderer
   Copyright (C) 2020 Cupcakus

   This program is free software; you can redistribute it and/or
   modify it under the terms of the GNU General Public License
   as published by the Free Software Foundation; Version 2
   of the License ONLY.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program; if not, write to the Free Software
   Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
*/

/*
#include <sys/ioctl.h>
#include <linux/fb.h>
struct fb_fix_screeninfo getFixScreenInfo(int fd) {
	struct fb_fix_screeninfo info;
	ioctl(fd, FBIOGET_FSCREENINFO, &info);
	return info;
}
struct fb_var_screeninfo getVarScreenInfo(int fd) {
	struct fb_var_screeninfo info;
	ioctl(fd, FBIOGET_VSCREENINFO, &info);
	return info;
}
*/
import "C"
import (
	"image"
	"image/color"
	"image/draw"
	"os"
	"syscall"

	"github.com/nfnt/resize"
)

//RendererLinux is a PIZero specific renderer
type RendererLinux struct {
	blink int
	dev   *Device
}

//NewRenderer makes and returns a new renderer
func NewRenderer() *RendererLinux {
	ren := RendererLinux{}

	file, err := os.OpenFile("/dev/fb0", os.O_RDWR, os.ModeDevice)
	if err != nil {
		panic(err)
	}

	fixInfo := C.getFixScreenInfo(C.int(file.Fd()))
	varInfo := C.getVarScreenInfo(C.int(file.Fd()))

	pixels, err := syscall.Mmap(
		int(file.Fd()),
		0, int(varInfo.xres*varInfo.yres*varInfo.bits_per_pixel),
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED,
	)
	if err != nil {
		file.Close()
		panic(err)
	}

	ren.dev = &Device{
		file,
		pixels,
		int(fixInfo.line_length),
		image.Rect(0, 0, int(varInfo.xres), int(varInfo.yres)),
	}

	return &ren
}

//Init set up the renderer
func (r *RendererLinux) Init() {
}

// Device represents the frame buffer. It implements the draw.Image interface.
type Device struct {
	file   *os.File
	pixels []byte
	pitch  int
	bounds image.Rectangle
}

// Close unmaps the framebuffer memory and closes the device file. Call this
// function when you are done using the frame buffer.
func (d *Device) Close() {
	syscall.Munmap(d.pixels)
	d.file.Close()
}

// Bounds implements the image.Image (and draw.Image) interface.
func (d *Device) Bounds() image.Rectangle {
	return d.bounds
}

// At implements the image.Image (and draw.Image) interface.
func (d *Device) At(x, y int) color.Color {
	if x < d.bounds.Min.X || x >= d.bounds.Max.X ||
		y < d.bounds.Min.Y || y >= d.bounds.Max.Y {
		return color.RGBA{0, 0, 0, 0}
	}
	i := y*d.pitch + 4*x
	return color.RGBA{d.pixels[i], d.pixels[i+1], d.pixels[i+2], 255}
}

// ColorModel implements the image.Image (and draw.Image) interface.
func (d *Device) ColorModel() color.Model {
	return color.RGBAModel
}

// Set implements the draw.Image interface.
func (d *Device) Set(x, y int, c color.Color) {
	// the min bounds are at 0,0 (see Open)
	if x >= 0 && x < d.bounds.Max.X &&
		y >= 0 && y < d.bounds.Max.Y {
		r, g, b, _ := c.RGBA()
		i := y*d.pitch + 4*x
		d.pixels[i] = byte(b)
		d.pixels[i+1] = byte(g)
		d.pixels[i+2] = byte(r)
		d.pixels[i+3] = byte(255)
	}
}

//Render renders the current display buffer
func (r *RendererLinux) Render(src draw.Image) {
	rect := image.Rect(0, 0, r.dev.bounds.Max.X, r.dev.bounds.Max.Y)
	img := resize.Resize(uint(rect.Size().X), uint(rect.Size().Y), src, resize.Lanczos3)
	draw.Draw(r.dev, rect, img, image.Point{0, 0}, draw.Src)
}

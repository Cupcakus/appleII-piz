package video

/* video.go -- Platform independent rendering system
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

import (
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"log"

	"github.com/cupcakus/appleII-piz/appleii"
)

//Renderer device specific renderer, framebuffer on Pi-Zero, GLFW window on Windows
type Renderer interface {
	Init()
	Render(draw.Image, draw.Image) image.Rectangle
}

//System Apple IIe Generic video system
type System struct {
	rom     []byte       //4k Character rom for text mode
	display *image.RGBA  //Display buffer
	bus     *appleii.Bus //Memory module holds the status flags
	ren     Renderer
}

var rowOffsets = [24]uint16{0x0, 0x80, 0x100, 0x180, 0x200, 0x280, 0x300, 0x380, 0x28, 0xA8, 0x128, 0x1A8, 0x228, 0x2A8, 0x328, 0x3A8, 0x50, 0xD0, 0x150, 0x1D0, 0x250, 0x2D0, 0x350, 0x3D0}

var lowResColors = [16]color.RGBA{{0, 0, 0, 255}, {147, 11, 124, 255}, {31, 53, 211, 255}, {187, 54, 255, 255}, {0, 118, 12, 255},
	{86, 86, 86, 255}, {7, 168, 224, 255}, {157, 172, 255, 255}, {98, 76, 0, 255}, {249, 86, 29, 255}, {126, 126, 126, 255},
	{255, 129, 236, 255}, {67, 200, 0, 255}, {220, 205, 22, 255}, {93, 247, 132, 255}, {255, 255, 255, 0}}

//var hiResColors = [16]int{black, darkblue, darkgreen, mediumblue, brown, gray2, lightgreen, aquamarine, magenta, purple, gray, lightblue, orange, pink, yellow, white}

var hiResColors = [16]uint8{black, magenta, brown, orange, darkgreen, gray, lightgreen, yellow, darkblue, purple, gray2, pink, mediumblue, lightblue, aquamarine, white}

//Color constants
const (
	black = iota
	magenta
	darkblue
	purple
	darkgreen
	gray
	mediumblue
	lightblue
	brown
	orange
	gray2
	pink
	lightgreen
	yellow
	aquamarine
	white
)

//NewVideo requires a pointer to the memory system, and a valid renderer
func NewVideo(b *appleii.Bus, r Renderer) *System {
	data, err := ioutil.ReadFile("./data/video.bin")
	if err != nil {
		log.Fatal("Failed to video ROM")
	}

	rect := image.Rect(0, 0, 560, 384)
	sys := System{rom: data, bus: b, ren: r, display: image.NewRGBA(rect)}
	r.Init()
	return &sys
}

func (s *System) drawHiResGlyph(addr int, mem []uint8) draw.Image {
	rect := image.Rect(0, 0, 14, 16)
	img := image.NewRGBA(rect)

	for y := 0; y < 16; y += 2 {
		data := mem[addr+((y>>1)*0x400)]
		alpha := uint8(255)
		if data&(1<<7) != 0 {
			//Stash the color block for these 7 pixels into the alpha channel
			alpha = 0
		}
		for x := 0; x < 14; x += 2 {
			if data&(1<<(x>>1)) != 0 {
				img.Set(x, y, color.RGBA{255, 255, 255, alpha})
				img.Set(x+1, y, color.RGBA{255, 255, 255, alpha})
				img.Set(x, y+1, color.RGBA{255, 255, 255, alpha})
				img.Set(x+1, y+1, color.RGBA{255, 255, 255, alpha})
			}
		}
	}

	return img
}

func getHiResColor(bits uint8) color.RGBA {
	return lowResColors[hiResColors[bits]]
}

//Double Hi Resolution on the Apple IIe is nutz! Documentation is scattered all around
//the interwebs.  In the event someone else stumbles on my code here is the relevent documentation
//for how this works...
//
// 1. http://www.appleoldies.ca/graphics/dhgr/dhgrtechnote.txt (Tech Note #3 describes memory layout and 4 pixel block pattern)
// 2. http://lukazi.blogspot.com/2017/03/double-high-resolution-graphics-dhgr.html (Lukazi explains how the moving window of the color burst causes interference on certain color transitions)
// 3. https://groups.google.com/g/comp.emulators.apple2/c/l_yFH3HIyQU/m/sWG9zrT1tegJ (Apparently bit 7 OFF on AUX1 byte of 4 byte group indicates color should be turned off
func (s *System) drawDblHiResGlyph(addr int, mem []uint8, aux []uint8) draw.Image {
	rect := image.Rect(0, 0, 28, 16)
	img := image.NewRGBA(rect)

	//A hires glyph is 28 pixels wide by 16 tall. The color resolution is only 7 pixels
	for y := 0; y < 16; y += 2 {
		data0 := aux[addr+((y>>1)*0x400)]
		data1 := mem[addr+((y>>1)*0x400)]
		data2 := aux[(addr+((y>>1)*0x400))+1]
		data3 := mem[(addr+((y>>1)*0x400))+1]

		for x := 0; x < 7; x++ {
			if data0&(1<<x) != 0 {
				img.Set(x, y, lowResColors[white])
				img.Set(x, y+1, lowResColors[white])
			}
			if data1&(1<<x) != 0 {
				img.Set(x+7, y, lowResColors[white])
				img.Set(x+7, y+1, lowResColors[white])
			}
			if data2&(1<<x) != 0 {
				img.Set(x+14, y, lowResColors[white])
				img.Set(x+14, y+1, lowResColors[white])
			}
			if data3&(1<<x) != 0 {
				img.Set(x+21, y, lowResColors[white])
				img.Set(x+21, y+1, lowResColors[white])
			}
		}
	}

	return img
}

func (s *System) drawLoResGlyph(addr int, mem []uint8) draw.Image {
	rect := image.Rect(0, 0, 14, 16)
	img := image.NewRGBA(rect)

	topColor := lowResColors[mem[addr]>>4]
	botColor := lowResColors[mem[addr]&0x0F]

	for y := 0; y < 16; y += 2 {
		for x := 0; x < 14; x += 2 {
			c := botColor
			if y >= 8 {
				c = topColor
			}
			img.Set(x, y, c)
			img.Set(x+1, y, c)
			img.Set(x, y+1, c)
			img.Set(x+1, y+1, c)
		}
	}

	return img
}

func (s *System) drawLoRes80Glyph(addr int, mem []uint8) draw.Image {
	rect := image.Rect(0, 0, 7, 16)
	img := image.NewRGBA(rect)

	topColor := lowResColors[mem[addr]>>4]
	botColor := lowResColors[mem[addr]&0x0F]

	for y := 0; y < 16; y += 2 {
		for x := 0; x < 7; x++ {
			c := botColor
			if y >= 8 {
				c = topColor
			}
			img.Set(x, y, c)
			img.Set(x, y+1, c)
		}
	}

	return img
}

func (s *System) drawGlyph(glyph uint8) draw.Image {
	rect := image.Rect(0, 0, 14, 16)
	img := image.NewRGBA(rect)
	offset := int(glyph) * 8

	for y := 0; y < 16; y += 2 {
		data := s.rom[offset+(y>>1)]
		for x := 0; x < 14; x += 2 {
			if data&(1<<(x>>1)) == 0 {
				img.Set(x, y, color.RGBA{255, 255, 255, 255})
				img.Set(x+1, y, color.RGBA{255, 255, 255, 255})
				img.Set(x, y+1, color.RGBA{255, 255, 255, 255})
				img.Set(x+1, y+1, color.RGBA{255, 255, 255, 255})
			}
		}
	}

	return img
}

func (s *System) draw80Glyph(glyph uint8) draw.Image {
	rect := image.Rect(0, 0, 7, 16)
	img := image.NewRGBA(rect)
	offset := int(glyph) * 8

	for y := 0; y < 16; y += 2 {
		data := s.rom[offset+(y>>1)]
		for x := 0; x < 7; x++ {
			if data&(1<<x) == 0 {
				img.Set(x, y, color.RGBA{255, 255, 255, 255})
				img.Set(x, y+1, color.RGBA{255, 255, 255, 255})
			}
		}
	}

	return img
}

var doubleHiResBlockFrom = [16][16]int{
	{0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000},
	{0x0000, 0x1110, 0x0000, 0x1110, 0x0000, 0x1110, 0x0000, 0x1110, 0x0000, 0x1110, 0x0000, 0x1110, 0x0000, 0x1110, 0x0000, 0x1110},
	{0x0000, 0x3300, 0x2200, 0x3300, 0x0000, 0x3300, 0x2200, 0x3300, 0x0000, 0x3300, 0x2200, 0x3300, 0x0000, 0x3300, 0x2200, 0x3300},
	{0x0000, 0x3300, 0x2200, 0x3300, 0x0000, 0x3300, 0x2200, 0x3300, 0x0000, 0x3300, 0x2200, 0x3300, 0x0000, 0x3300, 0x2200, 0x3300},
	{0x0400, 0x5500, 0x6400, 0x7500, 0x4400, 0x5500, 0x6400, 0x7500, 0x0400, 0x5500, 0x6400, 0x7500, 0x4400, 0x5500, 0x6400, 0x7500},
	{0x0500, 0x5500, 0x6500, 0x7500, 0x4500, 0x5500, 0x6500, 0x7500, 0x0500, 0x5500, 0x6500, 0x7500, 0x4500, 0x5500, 0x6500, 0x7500},
	{0x0600, 0x7700, 0x6600, 0x7700, 0x4600, 0x7700, 0x6600, 0x7700, 0x0600, 0x7700, 0x6600, 0x7700, 0x4600, 0x7700, 0x6600, 0x7700},
	{0x0700, 0x7700, 0x6700, 0x7700, 0x4700, 0x7700, 0x6700, 0x7700, 0x0700, 0x7700, 0x6700, 0x7700, 0x4700, 0x7700, 0x6700, 0x7700},
	{0x8000, 0x9000, 0xA000, 0xB000, 0x8000, 0x9000, 0xA000, 0xB000, 0x8000, 0x9000, 0xA000, 0xB000, 0x8000, 0x9000, 0xA000, 0xB000},
	{0x8990, 0x9990, 0xA990, 0xB990, 0x8990, 0x9990, 0xA990, 0xB990, 0x8990, 0x9990, 0xA990, 0xB990, 0x8990, 0x9990, 0xA990, 0xB990},
	{0xAAA0, 0xBBA0, 0xAAA0, 0xBBA0, 0xAAA0, 0xBBA0, 0xAAA0, 0xBBA0, 0xAAA0, 0xBBA0, 0xAAA0, 0xBBA0, 0xAAA0, 0xBBA0, 0xAAA0, 0xBBA0},
	{0xABB0, 0xBBB0, 0xABB0, 0xBBB0, 0xABB0, 0xBBB0, 0xABB0, 0xBBB0, 0xABB0, 0xBBB0, 0xABB0, 0xBBB0, 0xABB0, 0xBBB0, 0xABB0, 0xBBB0},
	{0xCC00, 0xDD00, 0xEC00, 0xFD00, 0xCC00, 0xDD00, 0xEC00, 0xFD00, 0xCC00, 0xDD00, 0xEC00, 0xFD00, 0xCC00, 0xDD00, 0xEC00, 0xFD00},
	{0xCDD0, 0xDDD0, 0xEDD0, 0xFDD0, 0xCDD0, 0xDDD0, 0xEDD0, 0xFDD0, 0xCDD0, 0xDDD0, 0xEDD0, 0xFDD0, 0xCDD0, 0xDDD0, 0xEDD0, 0xFDD0},
	{0xEEE0, 0xFFE0, 0xEEE0, 0xFFE0, 0xEEE0, 0xFFE0, 0xEEE0, 0xFFE0, 0xEEE0, 0xFFE0, 0xEEE0, 0xFFE0, 0xEEE0, 0xFFE0, 0xEEE0, 0xFFE0},
	{0xEFF0, 0xFFF0, 0xEFF0, 0xFFF0, 0xEFF0, 0xFFF0, 0xEFF0, 0xFFF0, 0xEFF0, 0xFFF0, 0xEFF0, 0xFFF0, 0xEFF0, 0xFFF0, 0xEFF0, 0xFFF0},
}

var doubleHiResBlockTo = [16][16]int{
	{0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000},
	{0x0001, 0x0001, 0x0001, 0x0001, 0x0005, 0x0005, 0x0005, 0x0005, 0x0009, 0x0009, 0x0009, 0x0009, 0x000D, 0x000D, 0x000D, 0x000D},
	{0x0020, 0x0020, 0x0022, 0x0022, 0x0026, 0x0026, 0x0026, 0x0026, 0x00AA, 0x00AA, 0x00AA, 0x00AA, 0x00AE, 0x00AE, 0x00AE, 0x00AE},
	{0x0033, 0x0033, 0x0033, 0x0033, 0x0037, 0x0037, 0x0037, 0x0037, 0x00BB, 0x00BB, 0x00BB, 0x00BB, 0x00BF, 0x00BF, 0x00BF, 0x00BF},
	{0x0000, 0x0000, 0x0000, 0x0000, 0x0044, 0x0044, 0x0044, 0x0044, 0x00CC, 0x00CC, 0x00CC, 0x00CC, 0x00CC, 0x00CC, 0x00CC, 0x00CC},
	{0x0055, 0x0055, 0x0055, 0x0055, 0x0055, 0x0055, 0x0055, 0x0055, 0x00DD, 0x00DD, 0x00DD, 0x00DD, 0x00DD, 0x00DD, 0x00DD, 0x00DD},
	{0x0060, 0x0060, 0x0062, 0x0062, 0x0066, 0x0066, 0x0066, 0x0066, 0x00EE, 0x00EE, 0x00EE, 0x00EE, 0x00EE, 0x00EE, 0x00EE, 0x00EE},
	{0x0077, 0x0077, 0x0077, 0x0077, 0x0077, 0x0077, 0x0077, 0x0077, 0x00FF, 0x00FF, 0x00FF, 0x00FF, 0x00FF, 0x00FF, 0x00FF, 0x00FF},
	{0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0888, 0x0888, 0x0888, 0x0888, 0x0888, 0x0888, 0x0888, 0x0888},
	{0x0001, 0x0001, 0x0001, 0x0001, 0x0005, 0x0005, 0x0005, 0x0005, 0x0009, 0x0009, 0x0009, 0x0009, 0x000D, 0x000D, 0x000D, 0x000D},
	{0x0000, 0x0000, 0x0002, 0x0002, 0x0006, 0x0006, 0x0006, 0x0006, 0x000A, 0x000A, 0x000A, 0x000A, 0x000E, 0x000E, 0x000E, 0x000E},
	{0x0003, 0x0003, 0x0003, 0x0003, 0x0007, 0x0007, 0x0007, 0x0007, 0x000B, 0x000B, 0x000B, 0x000B, 0x000F, 0x000F, 0x000F, 0x000F},
	{0x0000, 0x0000, 0x0000, 0x0000, 0x0044, 0x0044, 0x0044, 0x0044, 0x00CC, 0x00CC, 0x00CC, 0x00CC, 0x00CC, 0x00CC, 0x00CC, 0x00CC},
	{0x0005, 0x0005, 0x0005, 0x0005, 0x0005, 0x0005, 0x0005, 0x0005, 0x000D, 0x000D, 0x000D, 0x000D, 0x000D, 0x000D, 0x000D, 0x000D},
	{0x0000, 0x0000, 0x0002, 0x0002, 0x0006, 0x0006, 0x0006, 0x0006, 0x000E, 0x000E, 0x000E, 0x000E, 0x000E, 0x000E, 0x000E, 0x000E},
	{0x0007, 0x0007, 0x0007, 0x0007, 0x0007, 0x0007, 0x0007, 0x0007, 0x000F, 0x000F, 0x000F, 0x000F, 0x000F, 0x000F, 0x000F, 0x000F},
}

func (s *System) colorizeDblHiResDisplay(img *image.RGBA) *image.RGBA {
	rect := image.Rect(0, 0, 560, 384)
	ret := image.NewRGBA(rect)

	for y := 0; y < 384; y += 2 {
		for x := 0; x < 560; x += 4 {
			f1, f2, f3, f4 := uint32(0), uint32(0), uint32(0), uint32(0)
			if x > 0 {
				f1, _, _, _ = img.RGBAAt(x-4, y).RGBA()
				f2, _, _, _ = img.RGBAAt(x-3, y).RGBA()
				f3, _, _, _ = img.RGBAAt(x-2, y).RGBA()
				f4, _, _, _ = img.RGBAAt(x-1, y).RGBA()
			}
			fromColor := 0
			if f1 != 0 {
				fromColor |= 0x8
			}
			if f2 != 0 {
				fromColor |= 0x4
			}
			if f3 != 0 {
				fromColor |= 0x2
			}
			if f4 != 0 {
				fromColor |= 1
			}

			c1, _, _, _ := img.RGBAAt(x, y).RGBA()
			c2, _, _, _ := img.RGBAAt(x+1, y).RGBA()
			c3, _, _, _ := img.RGBAAt(x+2, y).RGBA()
			c4, _, _, _ := img.RGBAAt(x+3, y).RGBA()
			curColor := 0
			if c1 != 0 {
				curColor |= 0x8
			}
			if c2 != 0 {
				curColor |= 0x4
			}
			if c3 != 0 {
				curColor |= 0x2
			}
			if c4 != 0 {
				curColor |= 1
			}

			t1, t2, t3, t4 := uint32(0), uint32(0), uint32(0), uint32(0)
			if x < 556 {
				t1, _, _, _ = img.RGBAAt(x+4, y).RGBA()
				t2, _, _, _ = img.RGBAAt(x+5, y).RGBA()
				t3, _, _, _ = img.RGBAAt(x+6, y).RGBA()
				t4, _, _, _ = img.RGBAAt(x+7, y).RGBA()
			}
			toColor := 0
			if t1 != 0 {
				toColor |= 0x8
			}
			if t2 != 0 {
				toColor |= 0x4
			}
			if t3 != 0 {
				toColor |= 0x2
			}
			if t4 != 0 {
				toColor |= 1
			}

			resultColor := doubleHiResBlockFrom[curColor][fromColor] | doubleHiResBlockTo[curColor][toColor]

			ret.SetRGBA(x, y, lowResColors[hiResColors[(resultColor>>12)&0xF]])
			ret.SetRGBA(x, y+1, lowResColors[hiResColors[(resultColor>>12)&0xF]])

			ret.SetRGBA(x+1, y, lowResColors[hiResColors[(resultColor>>8)&0xF]])
			ret.SetRGBA(x+1, y+1, lowResColors[hiResColors[(resultColor>>8)&0xF]])

			ret.SetRGBA(x+2, y, lowResColors[hiResColors[(resultColor>>4)&0xF]])
			ret.SetRGBA(x+2, y+1, lowResColors[hiResColors[(resultColor>>4)&0xF]])

			ret.SetRGBA(x+3, y, lowResColors[hiResColors[resultColor&0xF]])
			ret.SetRGBA(x+3, y+1, lowResColors[hiResColors[resultColor&0xF]])
		}
	}
	return ret
}

func (s *System) colorizeDisplay(img *image.RGBA) *image.RGBA {
	rect := image.Rect(0, 0, 560, 384)
	ret := image.NewRGBA(rect)

	for y := 0; y < 384; y += 2 {
		for x := 0; x < 560; x += 2 {
			cBef := false
			cAt := false
			cNext := false
			c, _, _, a := img.RGBAAt(x, y).RGBA()
			//fmt.Printf("%d\n", c)
			if c != 0 {
				cAt = true
			}
			if x > 0 {
				cb, _, _, _ := img.RGBAAt(x-1, y).RGBA()
				if cb != 0 {
					cBef = true
				}
			}
			if x < 558 {
				ca, _, _, _ := img.RGBAAt(x+2, y).RGBA()
				if ca != 0 {
					cNext = true
				}
			}
			if cBef || cAt || cNext {

			}
			color := lowResColors[black]
			if cAt && !cBef && !cNext {
				if (x>>1)&1 != 0 {
					//Odd columns are either green or orange
					if a != 0 {
						color = lowResColors[lightgreen]
					} else {
						color = lowResColors[orange]
					}
				} else {
					//Even columns are either purple or blue
					if a != 0 {
						color = lowResColors[purple]
					} else {
						color = lowResColors[mediumblue]
					}
				}
			} else if (cAt && cBef) || (cAt && cNext) {
				color = lowResColors[white]
			}

			ret.SetRGBA(x, y, color)
			ret.SetRGBA(x+1, y, color)
			ret.SetRGBA(x, y+1, color)
			ret.SetRGBA(x+1, y+1, color)
		}
	}
	return ret
}

//RenderFrame render the current display screen based on the current graphics mode
func (s *System) RenderFrame(gpuMem []uint8, gpuAuxMem []uint8, gpuStart uint16, textMode bool, hiRes bool, col80 bool, mixed bool, dblhires bool) func(drw draw.Image) image.Rectangle {
	//fmt.Printf("RENDER: START: 0x%x | TXT: %t | HIRES: %t | 80COL: %t | MIX: %t | DBLHIRES %t\n", gpuStart, textMode, hiRes, col80, mixed, dblhires)
	if !col80 {
		for y := 0; y < 24; y++ {
			if mixed && hiRes && !textMode && y == 20 {
				gpuStart -= 0x1C00
			}
			for x := 0; x < 40; x++ {
				rect := image.Rect(x*14, y*16, (x*14)+14, (y*16)+16)
				if hiRes && !textMode && (!mixed || y <= 19) {
					addr := gpuStart + uint16(x) + rowOffsets[y]
					draw.Draw(s.display, rect, s.drawHiResGlyph(int(addr), gpuMem), image.Point{0, 0}, draw.Src)
				} else if !hiRes && !textMode && (!mixed || y <= 19) {
					addr := gpuStart + uint16(x) + rowOffsets[y]
					draw.Draw(s.display, rect, s.drawLoResGlyph(int(addr), gpuMem), image.Point{0, 0}, draw.Src)
				} else {
					addr := uint16(x) + rowOffsets[y]
					glyph := gpuMem[gpuStart+addr]
					draw.Draw(s.display, rect, s.drawGlyph(glyph), image.Point{0, 0}, draw.Src)
				}
			}
		}
		if hiRes && !textMode {
			img := s.colorizeDisplay(s.display)
			return func(drw draw.Image) image.Rectangle { return s.ren.Render(drw, img) }
		}
	} else {
		//80 Column mode
		for y := 0; y < 24; y++ {
			if mixed && hiRes && !textMode && y == 20 {
				gpuStart -= 0x1C00
			}
			for x := 0; x < 80; x += 2 {
				rect := image.Rect(x*7, y*16, (x*7)+7, (y*16)+16)
				rect2 := image.Rect((x*7)+7, y*16, (x*7)+14, (y*16)+16)
				if hiRes && !textMode && (!mixed || y <= 19) {
					if dblhires {
						rect = image.Rect(x*7, y*16, (x*7)+28, (y*16)+16)
						addr := gpuStart + uint16(x>>1) + rowOffsets[y]
						draw.Draw(s.display, rect, s.drawDblHiResGlyph(int(addr), gpuMem, gpuAuxMem), image.Point{0, 0}, draw.Src)
						x += 2
						//fmt.Printf("0x%x\n", addr)
					} else {
						rect = image.Rect(x*7, y*16, (x*7)+14, (y*16)+16)
						addr := gpuStart + uint16(x>>1) + rowOffsets[y]
						draw.Draw(s.display, rect, s.drawHiResGlyph(int(addr), gpuMem), image.Point{0, 0}, draw.Src)
					}
				} else if !hiRes && !textMode && (!mixed || y <= 19) {
					addr := gpuStart + uint16(x>>1) + rowOffsets[y]
					draw.Draw(s.display, rect, s.drawLoResGlyph(int(addr), gpuMem), image.Point{0, 0}, draw.Src)
					if dblhires {
						draw.Draw(s.display, rect2, s.drawLoResGlyph(int(addr), gpuAuxMem), image.Point{0, 0}, draw.Src)
					} else {
						draw.Draw(s.display, rect2, s.drawLoResGlyph(int(addr), gpuMem), image.Point{0, 0}, draw.Src)
					}
				} else {
					//80 COL TEXT MODE
					addr := uint16(x>>1) + rowOffsets[y]
					glyph := gpuMem[gpuStart+addr]
					glyph2 := gpuAuxMem[gpuStart+addr]
					draw.Draw(s.display, rect2, s.draw80Glyph(glyph), image.Point{0, 0}, draw.Src)
					draw.Draw(s.display, rect, s.draw80Glyph(glyph2), image.Point{0, 0}, draw.Src)
				}
			}
		}
		if hiRes && !textMode {
			if dblhires {
				img := s.colorizeDblHiResDisplay(s.display)
				return func(drw draw.Image) image.Rectangle { return s.ren.Render(drw, img) }
			}
			img := s.colorizeDisplay(s.display)
			return func(drw draw.Image) image.Rectangle { return s.ren.Render(drw, img) }
		}
	}
	return func(drw draw.Image) image.Rectangle { return s.ren.Render(drw, s.display) }
}

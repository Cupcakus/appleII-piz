package sys

/* sys_windows.go -- Windows specific runtime for the AppleII-PIZ Emulator
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
	"time"

	"github.com/cupcakus/appleII-piz/appleii"
	"github.com/cupcakus/appleII-piz/video"
	"github.com/faiface/gui"
	"github.com/faiface/gui/win"
	"github.com/faiface/mainthread"
)

//WindowsRunner windows specific emulator runtime
type WindowsRunner struct{}

func renderLoop(env gui.Env, vid *video.System, cpu *appleii.CPU, mem *appleii.Mem, bus *appleii.Bus) {
	for {
		//	fmt.Println("?")
		i := 0
		start := time.Now()
		for i <= 17030 {
			i += cpu.Tick()
			if i <= 4550 {
				mem.VBLANK = true
			} else {
				mem.VBLANK = false
			}
		}
		if !bus.GetFastMode() {
			env.Draw() <- vid.RenderFrame(mem.GetGPUMemory())
		}
		end := time.Now()
		sleepTime := 16 - end.Sub(start).Milliseconds()
		if sleepTime > 0 {
			if !bus.GetFastMode() {
				time.Sleep(time.Duration(sleepTime) * time.Millisecond)
			}
		}
	}
}

func getAppleKey(key win.Key) appleii.SysKey {
	switch key {
	case win.KeyHome:
		return appleii.KeyReset
	case win.KeyShift:
		return appleii.KeyShift
	case win.KeyCtrl:
		return appleii.KeyControl
	case win.KeyAlt:
		return appleii.KeyOpenApple
	case win.KeyEnd:
		return appleii.KeyFilledApple
	case win.KeyLeft, win.KeyBackspace:
		return appleii.KeyLeft
	case win.KeyRight:
		return appleii.KeyRight
	case win.KeyUp:
		return appleii.KeyUp
	case win.KeyDown:
		return appleii.KeyDown
	case win.KeyEscape:
		return appleii.KeyEscape
	case win.KeyEnter:
		return appleii.KeyReturn
	case win.KeyDelete:
		return appleii.KeyDelete

	default:
		return 0
	}
}

func run() {
	w, err := win.New(win.Title("Apple //e Emulator for Pi-Zero -- Windows Version For DEBUG ONLY"), win.Size(1024, 768))
	if err != nil {
		panic(err)
	}
	bus := appleii.NewBus()
	cpu := appleii.NewCPU(bus)
	mem := appleii.NewMem(bus)
	appleii.NewDsk(bus)
	bus.Add(mem, 0, 0xFFFF)
	kbd := appleii.NewKbd(mem, cpu)
	ren := video.NewRenderer()
	vid := video.NewVideo(bus, ren)

	cpu.Reset()

	r := image.Rect(0, 0, 640, 480)
	img := image.NewRGBA(r)
	for y := 0; y < 480; y++ {
		for x := 0; x < 640; x++ {
			img.SetRGBA(x, y, color.RGBA{50, 50, 50, 255})
		}
	}

	w.Draw() <- func(drw draw.Image) image.Rectangle {
		r := image.Rect(0, 0, 640, 480)
		draw.Draw(drw, r, img, image.ZP, draw.Src)
		return r
	}
	mux, env := gui.NewMux(w)
	go renderLoop(mux.MakeEnv(), vid, cpu, mem, bus)

	for event := range env.Events() {
		switch event.(type) {
		case win.WiClose:
			close(env.Draw())
		case win.KbType:
			kbd.KeyType(int(event.(win.KbType).Rune))
		case win.KbDown:
			//fmt.Printf("KEY %s\n", event.(win.KbDown).Key)
			kbd.SysKeyDn(getAppleKey(event.(win.KbDown).Key))
		case win.KbUp:
			kbd.SysKeyUp(getAppleKey(event.(win.KbUp).Key))
		}
	}
}

//NewRunner returns a new WindowsRunner
func NewRunner() *WindowsRunner {
	runner := WindowsRunner{}
	return &runner
}

//Init startup the runner
func (r *WindowsRunner) Init() {
}

//Run the runtime
func (r *WindowsRunner) Run() string {
	mainthread.Run(run)
	return ""
}

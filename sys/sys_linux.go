package sys

/* sys_linux.go -- Pi Zero specific runtime for the AppleII-PIZ Emulator
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
	"time"

	"github.com/cupcakus/appleII-piz/appleii"
	"github.com/cupcakus/appleII-piz/video"
)

//LinuxRunner pi zero specific emulator runtime
type LinuxRunner struct{}

func getAppleKey(key int) appleii.SysKey {
	switch key {
	case 0:
		return appleii.KeyReset
	default:
		return 0
	}
}

//NewRunner returns a new LinuxRunner
func NewRunner() *LinuxRunner {
	runner := LinuxRunner{}
	return &runner
}

//Init startup the runner
func (r *LinuxRunner) Init() {
}

//Run the runtime
func (r *LinuxRunner) Run() string {
	bus := appleii.NewBus()
	cpu := appleii.NewCPU(bus)
	mem := appleii.NewMem(bus, cpu)
	appleii.NewDsk(bus)
	bus.Add(mem, 0, 0xFFFF)
	//kbd := appleii.NewKbd(mem, cpu)
	ren := video.NewRenderer()
	vid := video.NewVideo(bus, ren)

	cpu.Reset()

	for {
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
			vid.RenderFrame(mem.GetGPUMemory())
		}
		end := time.Now()
		sleepTime := 16 - end.Sub(start).Milliseconds()
		if sleepTime > 0 {
			if !bus.GetFastMode() {
				//time.Sleep(time.Duration(sleepTime) * time.Millisecond)
			}
		}
	}

	return ""
}

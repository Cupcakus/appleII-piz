package appleii

/* mem.go -- Memory, IOU, and VIDEO
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
	"io/ioutil"
	"log"
)

//Mem is the AppleIIe memory space with extendend 80COL card...
type Mem struct {
	mem           []byte //64k Main Memory
	aux           []byte //64k Aux Memory
	rom           []byte //ROM from $C000-$FFFF
	boot          []byte //Disk ][ boot rom
	bus           *Bus
	keyboardLatch uint8
	preWrite      bool
	//MAIN/AUX is $0200 to $BFFF
	RDMAIN bool //True: Read from main, False: Read from aux
	WRMAIN bool //True: Write main, False: Write aux
	//MAINZP/AUXZP is $0000 to $01FF and $D000 to $FFFF
	MAINZP bool
	//ROM AREA FLAGS
	LCBNK2  bool //Which 4k bank is enabled from $D000-$DFFF
	LCRAM   bool //Do reads come from RAM or ROM?
	LCWRITE bool //Write to RAM Enabled
	//VIDEO MODE FLAGS
	STORE80  bool //Does PAGE2 flag swap to aux mem or not?
	PAGE2    bool //Sets Text Page to 0x800 - 0xC00
	HIRES    bool //Enables HIRES Mode
	VID80    bool //80 Column Mode
	ALTCHAR  bool //Alternate character ROM?
	TEXT     bool //Text mode or Graphics Mode
	MIXED    bool //Split screen
	VBLANK   bool //VBlank
	DBLHIRES bool //Double hi resolution Graphics Mode
	//SLOT FLAGS
	INTCXROM  bool
	SLOTC3ROM bool
	//KEYBOARD FLAGS
	KBDOAPPLE bool //Open Apple key
	KBDFAPPLE bool //Filled Apple key
	KBDSHIFT  bool //Shift key
}

//NewMem Creates a new RAM object (64k)
func NewMem(b *Bus) *Mem {
	m := Mem{bus: b, mem: make([]byte, 65536), aux: make([]byte, 65536), RDMAIN: true, WRMAIN: true, MAINZP: true}

	for i := 0; i < 65536; i += 4 {
		m.mem[i] = 0xFF
		m.mem[i+1] = 0xFF
		m.mem[i+2] = 0
		m.mem[i+3] = 0
		m.aux[i] = 0xFF
		m.aux[i+1] = 0xFF
		m.aux[i+2] = 0xFF
		m.aux[i+3] = 0xFF
	}

	data, err := ioutil.ReadFile("./data/system.bin")
	if err != nil {
		log.Fatal("Failed to load system ROM")
	}
	m.rom = data

	data, err = ioutil.ReadFile("./data/boot.bin")
	if err != nil {
		log.Fatal("Failed to load boot ROM")
	}
	m.boot = data

	return &m
}

//Reset the soft switches back to boot up
func (m *Mem) Reset() {
	m.RDMAIN = true
	m.WRMAIN = true
	m.MAINZP = true
	m.LCBNK2 = false
	m.LCRAM = false
	m.LCWRITE = false
	m.STORE80 = false
	m.PAGE2 = false
	m.HIRES = false
	m.VID80 = false
	m.ALTCHAR = false
	m.TEXT = false
	m.MIXED = false
	m.INTCXROM = false
	m.SLOTC3ROM = false
	m.DBLHIRES = false
	m.preWrite = false
}

//GetGPUMemory get a pointer to the memory of current GPU page
func (m *Mem) GetGPUMemory() ([]uint8, []uint8, uint16, bool, bool, bool, bool, bool) {
	addr := uint16(0x400)
	//Text and lowres mode
	if m.TEXT {
		if m.PAGE2 && !m.STORE80 {
			addr = 0x800
		}
	} else if m.HIRES {
		addr = 0x2000
		if m.PAGE2 && !m.STORE80 {
			addr = 0x4000
		}
	}

	if !m.VID80 && (!m.RDMAIN || (m.PAGE2 && m.STORE80)) {
		return m.aux, m.mem, addr, m.TEXT, m.HIRES, m.VID80, m.MIXED, m.DBLHIRES
	}

	return m.mem, m.aux, addr, m.TEXT, m.HIRES, m.VID80, m.MIXED, m.DBLHIRES
}

func (m *Mem) doLCBankSwitch(aRead bool) {
	switch m.bus.addr {
	case 0xC080:
		m.LCBNK2 = true
		m.LCRAM = true
		m.LCWRITE = false
	case 0xC081:
		m.LCBNK2 = true
		m.LCRAM = false
		if aRead {
			m.LCWRITE = m.preWrite
		}
		m.preWrite = aRead
	case 0xC082:
		m.LCBNK2 = true
		m.LCRAM = false
		m.LCWRITE = false
	case 0xC083:
		m.LCBNK2 = true
		m.LCRAM = true
		if aRead {
			m.LCWRITE = m.preWrite
		}
		m.preWrite = aRead
	case 0xC084:
		m.LCBNK2 = true
		m.LCRAM = true
		m.LCWRITE = false
	case 0xC085:
		m.LCBNK2 = true
		m.LCRAM = false
		if aRead {
			m.LCWRITE = m.preWrite
		}
		m.preWrite = aRead
	case 0xC086:
		m.LCBNK2 = true
		m.LCRAM = false
		m.LCWRITE = false
	case 0xC087:
		m.LCBNK2 = true
		m.LCRAM = true
		if aRead {
			m.LCWRITE = m.preWrite
		}
		m.preWrite = aRead
	case 0xC088:
		m.LCBNK2 = false
		m.LCRAM = true
		m.LCWRITE = false
	case 0xC089:
		m.LCBNK2 = false
		m.LCRAM = false
		if aRead {
			m.LCWRITE = m.preWrite
		}
		m.preWrite = aRead
	case 0xC08A:
		m.LCBNK2 = false
		m.LCRAM = false
		m.LCWRITE = false
	case 0xC08B:
		m.LCBNK2 = false
		m.LCRAM = true
		if aRead {
			m.LCWRITE = m.preWrite
		}
		m.preWrite = aRead
	case 0xC08C:
		m.LCBNK2 = false
		m.LCRAM = true
		m.LCWRITE = false
	case 0xC08D:
		m.LCBNK2 = false
		m.LCRAM = false
		if aRead {
			m.LCWRITE = m.preWrite
		}
		m.preWrite = aRead
	case 0xC08E:
		m.LCBNK2 = false
		m.LCRAM = false
		m.LCWRITE = false
	case 0xC08F:
		m.LCBNK2 = false
		m.LCRAM = true
		if aRead {
			m.LCWRITE = m.preWrite
		}
		m.preWrite = aRead
	}
}

//SysKeyDown tell the memory a key has been pressed
func (m *Mem) SysKeyDown(key int) {
	switch key {
	case KeyOpenApple:
		m.KBDOAPPLE = true
	case KeyFilledApple:
		m.KBDFAPPLE = true
	case KeyShift:
		m.KBDSHIFT = true
	}
}

//SysKeyUp tell the memory a key has been pressed
func (m *Mem) SysKeyUp(key int) {
	switch key {
	case KeyOpenApple:
		m.KBDOAPPLE = false
	case KeyFilledApple:
		m.KBDFAPPLE = false
	case KeyShift:
		m.KBDSHIFT = false
	}
}

//KeyDown tell the memory a key has been pressed
func (m *Mem) KeyDown(key int) {
	if m.keyboardLatch&(1<<7) == 0 {
		m.keyboardLatch = uint8(key)
		m.keyboardLatch |= 1 << 7
	}
}

func (m *Mem) ioRW(aRead bool) uint8 {
	if m.bus.addr >= 0xC080 && m.bus.addr <= 0xC08F {
		m.doLCBankSwitch(aRead)
	} else if aRead == false {
		//fmt.Printf("IOWRITE: 0x%x\n", m.bus.addr)
		//WRITE
		//Set the "Soft Switches"
		switch m.bus.addr {
		case 0xC000:
			m.STORE80 = false
		case 0xC001:
			m.STORE80 = true
		case 0xC002:
			m.RDMAIN = true
		case 0xC003:
			m.RDMAIN = false
		case 0xC004:
			m.WRMAIN = true
		case 0xC005:
			m.WRMAIN = false
		case 0xC006:
			m.INTCXROM = false
		case 0xC007:
			m.INTCXROM = true
		case 0xC008:
			m.MAINZP = true
		case 0xC009:
			m.MAINZP = false
		case 0xC00A:
			m.SLOTC3ROM = false
		case 0xC00B:
			m.SLOTC3ROM = true
		case 0xC00C:
			m.VID80 = false
		case 0xC00D:
			m.VID80 = true
		case 0xC00E:
			m.ALTCHAR = false
		case 0xC00F:
			m.ALTCHAR = true
		case 0xC010:
			m.keyboardLatch &= ^uint8(1 << 7)
		case 0xC050:
			m.TEXT = false
		case 0xC051:
			m.TEXT = true
		case 0xC052:
			m.MIXED = false
		case 0xC053:
			m.MIXED = true
		case 0xC054:
			m.PAGE2 = false
		case 0xC055:
			m.PAGE2 = true
		case 0xC056:
			m.HIRES = false
		case 0xC057:
			m.HIRES = true
		case 0xC05E:
			m.DBLHIRES = true
		case 0xC05F:
			m.DBLHIRES = false
		}
	} else {
		//fmt.Printf("IOREAD: 0x%x\n", m.bus.addr)
		switch m.bus.addr {
		case 0xC000:
			return m.keyboardLatch
		case 0xC010:
			m.keyboardLatch &= ^uint8(1 << 7)
		case 0xC011:
			if m.LCBNK2 {
				return 0x80
			}
		case 0xC012:
			if m.LCRAM {
				return 0x80
			}
		case 0xC013:
			if m.RDMAIN == false {
				return 0x80
			}
		case 0xC014:
			if m.WRMAIN == false {
				return 0x80
			}
		case 0xC015:
			if m.INTCXROM {
				return 0x80
			}
		case 0xC016:
			if m.MAINZP == false {
				return 0x80
			}
		case 0xC017:
			if m.SLOTC3ROM {
				return 0x80
			}
		case 0xC018:
			if m.STORE80 {
				return 0x80
			}
		case 0xC019:
			if m.VBLANK {
				return 0x80 //VBLank flag! Do we need this? I bet we do
			}
		case 0xC01A:
			if m.TEXT {
				return 0x80
			}
		case 0xC01B:
			if m.MIXED {
				return 0x80
			}
		case 0xC01C:
			if m.PAGE2 {
				return 0x80
			}
		case 0xC01D:
			if m.HIRES {
				return 0x80
			}
		case 0xC01E:
			if m.ALTCHAR {
				return 0x80
			}
		case 0xC01F:
			if m.VID80 {
				return 0x80
			}

		case 0xC050:
			m.TEXT = false
		case 0xC051:
			m.TEXT = true
		case 0xC052:
			m.MIXED = false
		case 0xC053:
			m.MIXED = true
		case 0xC054:
			m.PAGE2 = false
		case 0xC055:
			m.PAGE2 = true
		case 0xC056:
			m.HIRES = false
		case 0xC057:
			m.HIRES = true
		case 0xC05E:
			m.DBLHIRES = true
		case 0xC05F:
			m.DBLHIRES = false
		case 0xC061:
			if m.KBDOAPPLE {
				return 0x80
			}
		case 0xC062:
			if m.KBDFAPPLE {
				return 0x80
			}
		case 0xC063:
			if m.KBDSHIFT {
				return 0x80
			}
		case 0xC07F:
			if m.DBLHIRES {
				return 0x80
			}
		}
	}
	return 0
}

func (m *Mem) busUpdate() {
	if m.bus.cpuReadWrite {
		//MEMORY READ
		if m.bus.addr <= 0x1FF {
			//Zero page and stack...
			if m.MAINZP {
				m.bus.data = m.mem[m.bus.addr]
			} else {
				m.bus.data = m.aux[m.bus.addr]
			}
		} else if m.bus.addr >= 0x200 && m.bus.addr <= 0xBFFF {
			//Main 48k RAM area
			if m.STORE80 {
				if m.bus.addr >= 0x400 && m.bus.addr <= 0x7FF {
					if m.PAGE2 {
						m.bus.data = m.aux[m.bus.addr]
					} else {
						m.bus.data = m.mem[m.bus.addr]
					}
				} else if m.HIRES && m.bus.addr >= 0x2000 && m.bus.addr <= 0x3FFF {
					if m.PAGE2 {
						m.bus.data = m.aux[m.bus.addr]
					} else {
						m.bus.data = m.mem[m.bus.addr]
					}
				} else {
					if m.RDMAIN {
						m.bus.data = m.mem[m.bus.addr]
					} else {
						m.bus.data = m.aux[m.bus.addr]
					}
				}
			} else {
				if m.RDMAIN {
					m.bus.data = m.mem[m.bus.addr]
				} else {
					m.bus.data = m.aux[m.bus.addr]
				}
			}
		} else if m.bus.addr >= 0xC000 && m.bus.addr <= 0xCFFF {
			//IO Area, RAM isn't accessible here, you have to get it from a $D000 bank
			if m.bus.addr >= 0xC100 {
				if m.bus.addr >= 0xC300 && m.bus.addr <= 0xC3FF {
					m.bus.data = m.rom[m.bus.addr-0xC000]
				} else if !m.INTCXROM && m.bus.addr <= 0xC7FF {
					if m.bus.addr >= 0xC600 && m.bus.addr <= 0xC6FF {
						m.bus.data = m.boot[m.bus.addr-0xC600]
					} else {
						m.bus.data = 0 //No other slots connected
					}
				} else {
					m.bus.data = m.rom[m.bus.addr-0xC000]
				}
			} else {
				m.bus.data = m.ioRW(true)
			}
		} else if m.bus.addr >= 0xD000 && m.bus.addr <= 0xDFFF {
			//What 4k bank are we in?
			if m.LCRAM {
				//ROM is switched out, we are reading from RAM
				if m.LCBNK2 {
					if m.MAINZP {
						//Bank 2 -- Main RAM
						//fmt.Printf("READ 2 0x%x to 0x%x\n", m.bus.data, m.bus.addr)
						m.bus.data = m.mem[m.bus.addr]
					} else {
						//Bank 2 -- Aux RAM
						m.bus.data = m.aux[m.bus.addr]
					}
				} else {
					if m.MAINZP {
						//Bank 1 -- Main RAM
						//fmt.Printf("READ 1 0x%x to 0x%x\n", m.bus.data, m.bus.addr)
						m.bus.data = m.mem[m.bus.addr-0x1000]
					} else {
						//Bank 1 -- Aux RAM
						m.bus.data = m.aux[m.bus.addr-0x1000]
					}
				}
			} else {
				//Just give the ROM for this area
				//fmt.Printf("READ R 0x%x to 0x%x\n", m.bus.data, m.bus.addr)
				m.bus.data = m.rom[m.bus.addr-0xC000]
			}
		} else {
			//0xE000 to 0xFFFF
			if m.LCRAM {
				if m.MAINZP {
					m.bus.data = m.mem[m.bus.addr]
				} else {
					m.bus.data = m.aux[m.bus.addr]
				}
			} else {
				//Just give the ROM for this area
				m.bus.data = m.rom[m.bus.addr-0xC000]
			}
		}
	} else {
		//MEMORY WRITE
		if m.bus.addr <= 0x1FF {
			//Zero page and stack...
			if m.MAINZP {
				m.mem[m.bus.addr] = m.bus.data
			} else {
				m.aux[m.bus.addr] = m.bus.data
			}
		} else if m.bus.addr >= 0x200 && m.bus.addr <= 0xBFFF {
			//Main 48k RAM area
			if m.STORE80 {
				if m.bus.addr >= 0x400 && m.bus.addr <= 0x7FF {
					if m.PAGE2 {
						m.aux[m.bus.addr] = m.bus.data
					} else {
						m.mem[m.bus.addr] = m.bus.data
					}
				} else if m.HIRES && m.bus.addr >= 0x2000 && m.bus.addr <= 0x3FFF {
					if m.PAGE2 {
						m.aux[m.bus.addr] = m.bus.data
					} else {
						m.mem[m.bus.addr] = m.bus.data
					}
				} else {
					if m.WRMAIN {
						m.mem[m.bus.addr] = m.bus.data
					} else {
						m.aux[m.bus.addr] = m.bus.data
					}
				}
			} else {
				if m.WRMAIN {
					m.mem[m.bus.addr] = m.bus.data
				} else {
					m.aux[m.bus.addr] = m.bus.data
				}
			}
		} else if m.bus.addr >= 0xC000 && m.bus.addr <= 0xCFFF {
			//IO Area, RAM isn't accessible here, you have to get it from a $D000 bank
			m.ioRW(false)
		} else if m.bus.addr >= 0xD000 && m.bus.addr <= 0xDFFF {
			//What 4k bank are we in?
			if m.LCWRITE {
				//ROM is switched out, we are reading from RAM
				if m.LCBNK2 {
					if m.MAINZP {
						//Bank 2 -- Main RAM
						//fmt.Printf("WRITE 0x%x to 0x%x\n", m.bus.data, m.bus.addr)
						m.mem[m.bus.addr] = m.bus.data
					} else {
						//Bank 2 -- Aux RAM
						m.aux[m.bus.addr] = m.bus.data
					}
				} else {
					if m.MAINZP {
						//Bank 1 -- Main RAM
						m.mem[m.bus.addr-0x1000] = m.bus.data
					} else {
						//Bank 1 -- Aux RAM
						m.aux[m.bus.addr-0x1000] = m.bus.data
					}
				}
			}
		} else {
			//0xE000 to 0xFFFF
			if m.LCWRITE {
				if m.MAINZP {
					m.mem[m.bus.addr] = m.bus.data
				} else {
					m.aux[m.bus.addr] = m.bus.data
				}
			}
		}
	}
}

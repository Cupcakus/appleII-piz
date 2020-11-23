package appleii

/* kbd.go -- Platform independent keyboard interface to the AppleII
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

//Kbd is the keyboard object for the apple ii
type Kbd struct {
	mem      *Mem
	cpu      *CPU
	ctrlDown bool
}

//SysKey is a system key
type SysKey int

//System Keys
const (
	//KeyShift the Shift key
	KeyShift = iota
	//KeyOpenApple the Open Apple key
	KeyOpenApple
	//KeyFilledApple the Filled Apple key
	KeyFilledApple
	//KeyControl the CTRL key
	KeyControl
	//KeyLeft Left Arrow
	KeyLeft
	//KeyRight Right Arrow
	KeyRight
	//KeyUp Up Arrow
	KeyUp
	//KeyDown Down Arrow
	KeyDown
	//KeyReset the reset key
	KeyReset
	//KeyEscape the escape key
	KeyEscape
	//KeyReturn the return key
	KeyReturn
	//KeyDelete the delete key
	KeyDelete
)

//NewKbd create a new keyboard object
func NewKbd(m *Mem, c *CPU) *Kbd {
	k := Kbd{mem: m, cpu: c}
	return &k
}

//KeyType The last printable ASCII character typed ont the keyboard
func (k *Kbd) KeyType(key int) {
	key &= 0x7F
	k.mem.KeyDown(key)
}

//SysKeyDn A system key is being pressed down
func (k *Kbd) SysKeyDn(key SysKey) {
	switch key {
	case KeyShift:
		k.mem.SysKeyDown(KeyShift)
	case KeyOpenApple:
		k.mem.SysKeyDown(KeyOpenApple)
	case KeyFilledApple:
		k.mem.SysKeyDown(KeyFilledApple)
	case KeyControl:
		k.ctrlDown = true
	case KeyReset:
		if k.ctrlDown {
			k.cpu.Reset()
		}
	case KeyLeft:
		k.KeyType(0x08)
	case KeyRight:
		k.KeyType(0x15)
	case KeyUp:
		k.KeyType(0x0B)
	case KeyDown:
		k.KeyType(0x0A)
	case KeyEscape:
		k.KeyType(0x1B)
	case KeyReturn:
		k.KeyType(0x0D)
	case KeyDelete:
		k.KeyType(0x7F)
	}
}

//SysKeyUp A system key is no longer being pressed
func (k *Kbd) SysKeyUp(key SysKey) {
	switch key {
	case KeyShift:
		k.mem.SysKeyUp(KeyShift)
	case KeyOpenApple:
		k.mem.SysKeyUp(KeyOpenApple)
	case KeyFilledApple:
		k.mem.SysKeyUp(KeyFilledApple)
	case KeyControl:
		k.ctrlDown = false
	}
}

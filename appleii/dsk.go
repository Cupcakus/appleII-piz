package appleii

/* dsk.go -- AppleII dual disk controller card emulator
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

//Dsk ][ Controller
type Dsk struct {
	bus       *Bus
	motorOn   bool //Get your motor running
	drive2    bool //Selected drive
	q6        bool
	q7        bool
	dataLatch uint8
	disk1     *Diskette //Diskette in drive 1
	disk2     *Diskette //Diskette in drive 2
	phase     int       //Current stepper motor phase
	track     int       //Current track the head is on
	pos       int       //Current position on the track
}

//NewDsk create a new Disk ][ controller
func NewDsk(b *Bus) *Dsk {
	d := Dsk{bus: b}
	d.bus.Add(&d, 0xC0E0, 0xC0EF)

	//d.disk1 = NewDiskette("./disks/c.dsk")

	return &d
}

//Reset the disk controller
func (d *Dsk) Reset() {
	d.dataLatch = 0
	d.q6 = false
	d.q7 = false
	d.drive2 = false
	d.motorOn = false
	d.bus.SetFastMode(false)
}

//GetLEDStatus is the drive LED on?
func (d *Dsk) GetLEDStatus(disk2 bool) bool {
	if disk2 != d.drive2 {
		//Only one drive can be ON at any one time
		return false
	}
	return d.motorOn
}

func (d *Dsk) phaseChange(newPhase int) {
	if (d.phase == 1 && newPhase == 2) || (d.phase == 3 && newPhase == 0) {
		//When we move from phase 1 to phase 2 we go up one track
		d.track++
		if d.track > 34 {
			d.track = 34
		}
	} else if (d.phase == 1 && newPhase == 0) || (d.phase == 3 && newPhase == 2) {
		d.track--
		if d.track < 0 {
			d.track = 0
		}
	}
	//fmt.Printf("NEW TRACK! %d (PHASE %d)\n", d.track, newPhase)
	d.phase = newPhase
}

func (d *Dsk) updateData() {
	d.dataLatch = 0
	if !d.q6 && !d.q7 && d.motorOn { //READ mode, fill the data latch
		if !d.drive2 && d.disk1 != nil {
			d.dataLatch = d.disk1.Tracks[d.track][d.pos]
			d.pos++
			if d.pos == 6656 {
				d.pos = 0
			}
		}
	}
}

func (d *Dsk) busUpdate() {
	d.bus.data = 0
	switch d.bus.addr {
	case 0xC0E0:
		//I don't think I care about phases being turned off?
	case 0xC0E1:
		d.phaseChange(0)
	case 0xC0E2:
		d.bus.data = d.dataLatch
		//I don't think I care about phases being turned off?
	case 0xC0E3:
		d.phaseChange(1)
	case 0xC0E4:
		d.bus.data = d.dataLatch
		//I don't think I care about phases being turned off?
	case 0xC0E5:
		d.phaseChange(2)
	case 0xC0E6:
		d.bus.data = d.dataLatch
		//I don't think I care about phases being turned off?
	case 0xC0E7:
		d.phaseChange(3)
	case 0xC0E8:
		//fmt.Printf("MOTOR IS OFF (T:%d S:%d P:%d)\n", d.track, GetSector(d.pos), d.pos)
		d.bus.SetFastMode(false)
		d.bus.data = d.dataLatch
		d.motorOn = false
	case 0xC0E9:
		//fmt.Println("TURN THAT MOTOR ON!")
		if d.disk1 != nil || d.disk2 != nil {
			d.bus.SetFastMode(true)
		}
		d.motorOn = true
	case 0xC0EA:
		//fmt.Println("SLECT DRIVE 1")
		d.bus.data = d.dataLatch
		d.drive2 = false
	case 0xC0EB:
		//fmt.Println("SLECT DRIVE 2")
		d.drive2 = true
	case 0xC0EC:
		//fmt.Printf("READ BYTE 0x%x\n", d.dataLatch)
		d.bus.data = d.dataLatch
		d.updateData()
		d.q6 = false
	case 0xC0ED:
		d.q6 = true
	case 0xC0EE:
		d.q7 = false
	case 0xC0EF:
		d.q7 = true
	}
}

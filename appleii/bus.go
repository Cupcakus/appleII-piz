package appleii

/* bus.go -- A basic simulation of an AppleII Bus
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
	"log"
)

//Buser describes a device that sits on the bus in a specified address range
type Buser interface {
	busUpdate() //Called when the bus updates in an address range this object cares about
	Reset()
}

//Bus holds the current state of the BUS
type Bus struct {
	addr         uint16
	data         uint8
	cpuIRQ       bool
	cpuNMI       bool
	cpuReadWrite bool //True is read, False is write
	objects      []*BusObject
	fastMode     bool
}

//BusObject is an actual IC on the bus
type BusObject struct {
	object Buser
	start  uint16
	end    uint16
}

//NewBus creates and initializes a new Bus
func NewBus() *Bus {
	var bus Bus
	return &bus
}

//Add a new BusObject to this bus at specified address range...
//Order is important here! Objects added to the bus later will override those added earlier
func (b *Bus) Add(o Buser, startAddress uint16, endAddress uint16) {
	if endAddress < startAddress {
		log.Fatal("Object added to BUS must have an end address > start address")
	}
	obj := BusObject{object: o, start: startAddress, end: endAddress}
	b.objects = append([]*BusObject{&obj}, b.objects...)
}

//Set sets the lines on the bus, pass nil to a parameter to leave it as is
func (b *Bus) Set(aAddr *uint16, aData *uint8, aReadWrite *bool) {
	if aData != nil {
		b.data = *aData
	}
	if aReadWrite != nil {
		b.cpuReadWrite = *aReadWrite
	}
	if aAddr != nil {
		b.addr = *aAddr
		for _, o := range b.objects {
			if *aAddr >= o.start && *aAddr <= o.end {
				o.object.busUpdate()
			}
		}
	}
}

//Data gets the data currently on the bus
func (b *Bus) Data() uint8 {
	return b.data
}

//Reset the bus object when the CPU resets
func (b *Bus) Reset() {
	b.fastMode = false
	for _, o := range b.objects {
		o.object.Reset()
	}
}

//SetFastMode disable emulation throttles and go full speed (Used for speeding up diskette loading)
func (b *Bus) SetFastMode(mode bool) {
	b.fastMode = mode
}

//GetFastMode return the status of Fast Mode
func (b *Bus) GetFastMode() bool {
	return b.fastMode
}

package appleii

/* diskette.go -- Parses and expands a 35 Track/16 Sector DSK file into an actual diskette format
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

//Diskette DOS 3.3 -- 35 Tracks/16 Sector
type Diskette struct {
	//Tracks is the binary track data
	Tracks [35][]uint8
}

// DOS 3.3 used interleaved sectors
var sectorOrder = [16]int{
	0x0, 0x7, 0xE, 0x6, 0xD, 0x5, 0xC, 0x4, 0xB, 0x3, 0xA, 0x2, 0x9, 0x1, 0x8, 0xF,
}

var writeSectorOrder = [16]int{
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
}

var translateTable62 = [0x40]uint8{
	0x96, 0x97, 0x9a, 0x9b, 0x9d, 0x9e, 0x9f, 0xa6, 0xa7, 0xab, 0xac, 0xad, 0xae, 0xaf, 0xb2, 0xb3,
	0xb4, 0xb5, 0xb6, 0xb7, 0xb9, 0xba, 0xbb, 0xbc, 0xbd, 0xbe, 0xbf, 0xcb, 0xcd, 0xce, 0xcf, 0xd3,
	0xd6, 0xd7, 0xd9, 0xda, 0xdb, 0xdc, 0xdd, 0xde, 0xdf, 0xe5, 0xe6, 0xe7, 0xe9, 0xea, 0xeb, 0xec,
	0xed, 0xee, 0xef, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7, 0xf9, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff,
}

var volume uint8 = 254

//NewDiskette loads a diskette into one of the drives
func NewDiskette(filename string) *Diskette {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Failed to load diskette!")
	}

	if len(data) != 143360 {
		log.Fatalf("%s is not a valid Apple IIe diskette image", filename)
	}

	gap1 := make([]byte, 0x30)
	for i := range gap1 {
		gap1[i] = 0xff
	}
	gap2 := make([]byte, 5)
	for i := range gap2 {
		gap2[i] = 0xff
	}

	d := Diskette{}
	for t := uint8(0); t < 35; t++ { //35 Tracks
		for s := 15; s >= 0; s-- { //16 Sectors
			_s := writeSectorOrder[s]
			d.Tracks[t] = append(d.Tracks[t], gap1...)
			//Address Block
			d.Tracks[t] = append(d.Tracks[t], 0xD5, 0xAA, 0x96)
			d.Tracks[t] = append(d.Tracks[t], encode44(volume)...)
			d.Tracks[t] = append(d.Tracks[t], encode44(t)...)
			d.Tracks[t] = append(d.Tracks[t], encode44(uint8(_s))...)
			checksum := int(volume) ^ int(t) ^ int(_s)
			d.Tracks[t] = append(d.Tracks[t], encode44(uint8(checksum))...)
			d.Tracks[t] = append(d.Tracks[t], 0xDE, 0xAA, 0xEB)

			d.Tracks[t] = append(d.Tracks[t], gap2...)

			//Data Block
			pBuffer := make([]uint8, 256) //Primary Buffer
			sBuffer := make([]uint8, 86)  //Secondary Buffer
			d.Tracks[t] = append(d.Tracks[t], 0xD5, 0xAA, 0xAD)
			offset := (16*int(t) + sectorOrder[_s]) * 256
			sData := data[offset : offset+256]
			for b := 0; b < 256; b++ { //256 Bytes per sector
				pBuffer[b] = sData[b] >> 2
				lShift := uint((b / 86) * 2)
				bit0 := ((sData[b] & 0x01) << 1) << lShift
				bit1 := ((sData[b] & 0x02) >> 1) << lShift
				sOffset := b % 86
				sBuffer[sOffset] |= bit0 | bit1
			}
			last := byte(0)
			for _, v := range sBuffer {
				d.Tracks[t] = append(d.Tracks[t], translateTable62[v^last])
				last = v
			}
			for _, v := range pBuffer {
				d.Tracks[t] = append(d.Tracks[t], translateTable62[v^last])
				last = v
			}
			d.Tracks[t] = append(d.Tracks[t], translateTable62[last]) // Checksum
			d.Tracks[t] = append(d.Tracks[t], 0xDE, 0xAA, 0xEB)
		}
	}
	return &d
}

//Odd/Even encode a byte into two bytes
func encode44(b uint8) []uint8 {
	a := make([]uint8, 2)
	a[0] = ((b >> 1) & 0x55) | 0xaa
	a[1] = (b & 0x55) | 0xaa
	return a
}

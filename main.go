package main

/* main.go -- Main module for AppleII-PIZ Emulator
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

	"github.com/cupcakus/appleII-piz/sys"
)

func main() {
	runner := sys.NewRunner()
	runner.Init()
	err := runner.Run()
	if err != "" {
		log.Fatal(err)
	}
}

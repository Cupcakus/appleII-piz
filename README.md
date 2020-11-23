# AppleII-PIZ
Apple IIe Emulator specifically written in GO for the Raspberry PI Zero

## Intended Use
This emulator's intended use is for projects running in a limited resource environment.  
Namely, an environment with very small or hard to see video output.

If you are looking for a complete Apple II emulator with all the bells and whistles look elsewhere.

*However* if... like me, you want an Apple II emulator that runs full speed on a Raspberry PI Zero, and is intended to be viewed on displays as small as 3 inches 640x480. You have come to the right place!

## Running the Emulator
The emulator will run on Windows and on Raspberry PI OS Lite only.

To run the emulator type `go run main.go` on a command line

Special Keys are hard mapped as so:

`HOME -> RESET`

`ALT -> OPEN APPLE`

`END -> SOLID APPLE`

All other keys match 1:1 with a standard PC keyboard

## Emulated Features
* Apple IIe ONLY (No IIc/IIgs features)
* 80 Column Text
* Expanded Memory to 128k
* Mixed Graphics/Text in all modes
* RGB Color / Monochrome (No NTSC Filters)
* Low Resolution Graphics
* High Resolution Graphics
* Double High Resolution Graphics
* Disk ][ Controller Support (35 Track/16 Sector DOS 3.3 Disks Only) (Read Only for now)

## Still TODO
* Audio (Audio will be PI only, and will require a speaker on the GPIO)
* Rendering performance improvements
* Disk write support
* GUI for loading/ejecting disks
* Support for hardware disks (Special feature of my AppleIIe Mini project)


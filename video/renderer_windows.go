package video

import (
	"image"
	"image/draw"

	"github.com/nfnt/resize"
)

//RendererWindows is a windows specific renderer
type RendererWindows struct {
	blink  int
	width  int
	height int
}

var gImg draw.Image

//NewRenderer makes and returns a new renderer, specify the device width & height
func NewRenderer(w, h int) *RendererWindows {
	ren := RendererWindows{width: w, height: h}
	gImg = image.NewRGBA(image.Rect(0, 0, w, h))
	return &ren
}

//WindowsDraw Windows specific draw routine
func WindowsDraw(drw draw.Image) image.Rectangle {
	draw.Draw(drw, gImg.Bounds(), gImg, image.Point{0, 0}, draw.Src)
	return gImg.Bounds()
}

//Init set up the renderer
func (r *RendererWindows) Init() {

}

//Render renders the current display buffer
func (r *RendererWindows) Render(src draw.Image) {
	rect := image.Rect(0, 0, r.width, r.height)
	img := resize.Resize(uint(rect.Size().X), uint(rect.Size().Y), src, resize.Lanczos3)
	draw.Draw(gImg, rect, img, image.Point{0, 0}, draw.Src)
}

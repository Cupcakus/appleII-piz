package video

import (
	"image"
	"image/draw"

	"github.com/nfnt/resize"
)

//RendererWindows is a windows specific renderer
type RendererWindows struct {
	blink int
}

//NewRenderer makes and returns a new renderer
func NewRenderer() *RendererWindows {
	ren := RendererWindows{}
	return &ren
}

//Init set up the renderer
func (r *RendererWindows) Init() {

}

//Render renders the current display buffer
func (r *RendererWindows) Render(drw draw.Image, src draw.Image) image.Rectangle {
	rect := drw.Bounds()
	img := resize.Resize(uint(rect.Size().X), uint(rect.Size().Y), src, resize.Lanczos3)
	draw.Draw(drw, rect, img, image.Point{0, 0}, draw.Src)
	return rect
}

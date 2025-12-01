package tblock

import "fyne.io/fyne/v2"

type progressBarLayout struct {
	height float32
}

func (p *progressBarLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) > 0 {
		objects[0].Resize(fyne.NewSize(size.Width, p.height))
		objects[0].Move(fyne.NewPos(0, (size.Height-p.height)/2))
	}
}

func (p *progressBarLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) > 0 {
		min := objects[0].MinSize()
		return fyne.NewSize(min.Width, p.height)
	}
	return fyne.NewSize(0, p.height)
}

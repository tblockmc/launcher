package tblock

import "fyne.io/fyne/v2"

type RatioLayout struct {
	ratio float32
}

func NewRatioLayout(ratio float32) fyne.Layout {
	return &RatioLayout{ratio: ratio}
}

func (r *RatioLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) != 2 {
		return
	}

	totalParts := r.ratio + 1.0
	leftWidth := size.Width * r.ratio / totalParts
	rightWidth := size.Width - leftWidth*1.01

	leftMin := objects[0].MinSize()
	leftHeight := fyne.Max(leftMin.Height, size.Height)
	objects[0].Resize(fyne.NewSize(leftWidth, leftHeight))
	objects[0].Move(fyne.NewPos(0, 0))

	rightMin := objects[1].MinSize()
	rightHeight := fyne.Max(rightMin.Height, size.Height)
	objects[1].Resize(fyne.NewSize(rightWidth, rightHeight))
	objects[1].Move(fyne.NewPos(leftWidth*1.01, 0))
}

func (r *RatioLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) != 2 {
		return fyne.NewSize(0, 0)
	}

	leftMin := objects[0].MinSize()
	rightMin := objects[1].MinSize()

	minWidthFromRatio := fyne.Max(
		leftMin.Width*(1.0+1.0/r.ratio),
		rightMin.Width*(1.0+r.ratio),
	)

	totalWidth := fyne.Max(minWidthFromRatio, leftMin.Width+rightMin.Width)
	maxHeight := fyne.Max(leftMin.Height, rightMin.Height)

	return fyne.NewSize(totalWidth, maxHeight)
}

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

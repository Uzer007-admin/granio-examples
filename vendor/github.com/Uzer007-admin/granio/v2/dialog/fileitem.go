package dialog

import (
	"path/filepath"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/lang"
	"github.com/Uzer007-admin/granio/v2/storage"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"

	"github.com/Uzer007-admin/granio/v2/internal/foldermeta"
)

const (
	cellWidthIcon = iconSize * 1.25

	folderInsetBottom       = 25
	folderInsetBottomInline = 6
	folderInsetTop          = 20
	folderInsetTopInline    = 6
	folderInsetX            = 10
	folderInsetXInline      = 4

	iconSize       = 64
	iconSizeInline = 24
)

type fileDialogItem struct {
	widget.BaseWidget
	picker *fileDialog

	name     string
	id       int // id in the parent container
	choose   func(id int)
	open     func()
	location granio.URI
	dir      bool

	lastClick time.Time
}

func (i *fileDialogItem) CreateRenderer() granio.WidgetRenderer {
	text := widget.NewLabelWithStyle(i.name, granio.TextAlignCenter, granio.TextStyle{})
	text.Truncation = granio.TextTruncateEllipsis
	text.Wrapping = granio.TextWrapBreak
	icon := widget.NewFileIcon(i.location)
	over := &canvas.Image{}

	return &fileItemRenderer{
		item:         i,
		icon:         icon,
		text:         text,
		over:         over,
		objects:      []granio.CanvasObject{icon, text, over},
		fileTextSize: widget.NewLabel("M\nM").MinSize().Height, // cache two-line label height,
	}
}

func (i *fileDialogItem) setLocation(l granio.URI, dir, up bool) {
	i.dir = dir
	i.location = l
	i.name = l.Name()

	if i.picker.view == GridView {
		ext := filepath.Ext(i.name[1:])
		i.name = i.name[:len(i.name)-len(ext)]
	}

	if up {
		i.name = "(" + lang.X("file.parent", "Parent") + ")"
	}

	i.Refresh()
}

func (i *fileDialogItem) Tapped(*granio.PointEvent) {
	if i.choose != nil {
		i.choose(i.id)
	}
	now := time.Now()
	if !i.dir && now.Sub(i.lastClick) < granio.CurrentApp().Driver().DoubleTapDelay() && i.open != nil {
		// It is a double click, so we ask the dialog to open
		i.open()
	}
	i.lastClick = now
}

func (f *fileDialog) newFileItem(location granio.URI, dir, up bool) *fileDialogItem {
	item := &fileDialogItem{
		picker:   f,
		location: location,
		name:     location.Name(),
		dir:      dir,
	}

	if f.view == GridView {
		ext := filepath.Ext(item.name[1:])
		item.name = item.name[:len(item.name)-len(ext)]
	}

	if up {
		item.name = "(" + lang.X("file.parent", "Parent") + ")"
	}

	item.ExtendBaseWidget(item)
	return item
}

type fileItemRenderer struct {
	item         *fileDialogItem
	fileTextSize float32

	icon    *widget.FileIcon
	text    *widget.Label
	over    *canvas.Image
	objects []granio.CanvasObject
}

func (s *fileItemRenderer) Layout(size granio.Size) {
	if s.item.picker.view == GridView {
		s.icon.Resize(granio.NewSquareSize(iconSize))
		s.icon.Move(granio.NewPos((size.Width-iconSize)/2, 0))

		s.over.Resize(granio.NewSize(iconSize-folderInsetX*2, iconSize-folderInsetX-folderInsetBottom))
		s.over.Move(s.icon.Position().AddXY(folderInsetX, folderInsetTop))

		s.text.Alignment = granio.TextAlignCenter
		s.text.Resize(granio.NewSize(size.Width, s.fileTextSize))
		s.text.Move(granio.NewPos(0, size.Height-s.fileTextSize))
	} else {
		s.icon.Resize(granio.NewSquareSize(iconSizeInline))
		s.icon.Move(granio.NewPos(theme.Padding(), (size.Height-iconSizeInline)/2))

		s.over.Resize(granio.NewSize(iconSizeInline-folderInsetXInline*2, iconSizeInline-folderInsetXInline-folderInsetBottomInline))
		s.over.Move(s.icon.Position().AddXY(folderInsetXInline, folderInsetTopInline))

		s.text.Alignment = granio.TextAlignLeading
		textMin := s.text.MinSize()
		s.text.Resize(granio.NewSize(size.Width, textMin.Height))
		s.text.Move(granio.NewPos(iconSizeInline, (size.Height-textMin.Height)/2))
	}
}

func (s *fileItemRenderer) MinSize() granio.Size {
	if s.item.picker.view == GridView {
		return granio.NewSize(cellWidthIcon, iconSize+s.fileTextSize)
	}

	textMin := s.text.MinSize()
	return granio.NewSize(iconSizeInline+textMin.Width+theme.Padding(), textMin.Height)
}

func (s *fileItemRenderer) Refresh() {
	s.fileTextSize = widget.NewLabel("M\nM").MinSize().Height // cache two-line label height

	s.text.SetText(s.item.name)
	s.icon.SetURI(s.item.location)

	loc := s.item.location
	if loc.Path()[len(loc.Path())-1] == '/' {
		loc, _ = storage.ParseURI(loc.String()[:len(loc.String())-1])
	}

	if ff, _ := foldermeta.DetailsForFolder(loc); ff != nil {
		if ff.BackgroundURI != nil {
			s.over.File = ff.BackgroundURI.Path()
		} else {
			s.over.File = ""
		}
		if ff.BackgroundResource != nil {
			s.over.Resource = theme.NewColoredResource(ff.BackgroundResource, theme.ColorNameBackground)
		} else {
			s.over.Resource = nil
		}
		s.over.FillMode = ff.BackgroundFill
	} else {
		s.over.File = ""
		s.over.Resource = nil
		s.over.Image = nil
	}

	s.over.Refresh()
}

func (s *fileItemRenderer) Objects() []granio.CanvasObject {
	return s.objects
}

func (*fileItemRenderer) Destroy() {
}

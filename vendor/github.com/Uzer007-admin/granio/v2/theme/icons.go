package theme

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/svg"
)

const (
	// IconNameCancel is the name of theme lookup for cancel icon.
	//
	// Since: 2.0
	IconNameCancel granio.ThemeIconName = "cancel"

	// IconNameConfirm is the name of theme lookup for confirm icon.
	//
	// Since: 2.0
	IconNameConfirm granio.ThemeIconName = "confirm"

	// IconNameDelete is the name of theme lookup for delete icon.
	//
	// Since: 2.0
	IconNameDelete granio.ThemeIconName = "delete"

	// IconNameSearch is the name of theme lookup for search icon.
	//
	// Since: 2.0
	IconNameSearch granio.ThemeIconName = "search"

	// IconNameSearchReplace is the name of theme lookup for search and replace icon.
	//
	// Since: 2.0
	IconNameSearchReplace granio.ThemeIconName = "searchReplace"

	// IconNameMenu is the name of theme lookup for menu icon.
	//
	// Since: 2.0
	IconNameMenu granio.ThemeIconName = "menu"

	// IconNameMenuExpand is the name of theme lookup for menu expansion icon.
	//
	// Since: 2.0
	IconNameMenuExpand granio.ThemeIconName = "menuExpand"

	// IconNameCheckButton is the name of theme lookup for unchecked check button icon.
	//
	// Since: 2.0
	IconNameCheckButton granio.ThemeIconName = "unchecked"

	// IconNameCheckButtonChecked is the name of theme lookup for checked check button icon.
	//
	// Since: 2.0
	IconNameCheckButtonChecked granio.ThemeIconName = "checked"

	// IconNameCheckButtonFill is the name of theme lookup for filled check button icon.
	//
	// Since: 2.5
	IconNameCheckButtonFill granio.ThemeIconName = "iconNameCheckButtonFill"

	// IconNameCheckButtonPartial is the name of theme lookup for "partially" checked check button icon.
	//
	// Since: 2.6
	IconNameCheckButtonPartial granio.ThemeIconName = "partial"

	// IconNameRadioButton is the name of theme lookup for radio button unchecked icon.
	//
	// Since: 2.0
	IconNameRadioButton granio.ThemeIconName = "radioButton"

	// IconNameRadioButtonChecked is the name of theme lookup for radio button checked icon.
	//
	// Since: 2.0
	IconNameRadioButtonChecked granio.ThemeIconName = "radioButtonChecked"

	// IconNameRadioButtonFill is the name of theme lookup for filled radio button icon.
	//
	// Since: 2.5
	IconNameRadioButtonFill granio.ThemeIconName = "iconNameRadioButtonFill"

	// IconNameColorAchromatic is the name of theme lookup for greyscale color icon.
	//
	// Since: 2.0
	IconNameColorAchromatic granio.ThemeIconName = "colorAchromatic"

	// IconNameColorChromatic is the name of theme lookup for full color icon.
	//
	// Since: 2.0
	IconNameColorChromatic granio.ThemeIconName = "colorChromatic"

	// IconNameColorPalette is the name of theme lookup for color palette icon.
	//
	// Since: 2.0
	IconNameColorPalette granio.ThemeIconName = "colorPalette"

	// IconNameContentAdd is the name of theme lookup for content add icon.
	//
	// Since: 2.0
	IconNameContentAdd granio.ThemeIconName = "contentAdd"

	// IconNameContentRemove is the name of theme lookup for content remove icon.
	//
	// Since: 2.0
	IconNameContentRemove granio.ThemeIconName = "contentRemove"

	// IconNameContentCut is the name of theme lookup for content cut icon.
	//
	// Since: 2.0
	IconNameContentCut granio.ThemeIconName = "contentCut"

	// IconNameContentCopy is the name of theme lookup for content copy icon.
	//
	// Since: 2.0
	IconNameContentCopy granio.ThemeIconName = "contentCopy"

	// IconNameContentPaste is the name of theme lookup for content paste icon.
	//
	// Since: 2.0
	IconNameContentPaste granio.ThemeIconName = "contentPaste"

	// IconNameContentClear is the name of theme lookup for content clear icon.
	//
	// Since: 2.0
	IconNameContentClear granio.ThemeIconName = "contentClear"

	// IconNameContentRedo is the name of theme lookup for content redo icon.
	//
	// Since: 2.0
	IconNameContentRedo granio.ThemeIconName = "contentRedo"

	// IconNameContentUndo is the name of theme lookup for content undo icon.
	//
	// Since: 2.0
	IconNameContentUndo granio.ThemeIconName = "contentUndo"

	// IconNameInfo is the name of theme lookup for info icon.
	//
	// Since: 2.0
	IconNameInfo granio.ThemeIconName = "info"

	// IconNameQuestion is the name of theme lookup for question icon.
	//
	// Since: 2.0
	IconNameQuestion granio.ThemeIconName = "question"

	// IconNameWarning is the name of theme lookup for warning icon.
	//
	// Since: 2.0
	IconNameWarning granio.ThemeIconName = "warning"

	// IconNameError is the name of theme lookup for error icon.
	//
	// Since: 2.0
	IconNameError granio.ThemeIconName = "error"

	// IconNameBrokenImage is the name of the theme lookup for broken-image icon.
	//
	// Since: 2.4
	IconNameBrokenImage granio.ThemeIconName = "broken-image"

	// IconNameDocument is the name of theme lookup for document icon.
	//
	// Since: 2.0
	IconNameDocument granio.ThemeIconName = "document"

	// IconNameDocumentCreate is the name of theme lookup for document create icon.
	//
	// Since: 2.0
	IconNameDocumentCreate granio.ThemeIconName = "documentCreate"

	// IconNameDocumentPrint is the name of theme lookup for document print icon.
	//
	// Since: 2.0
	IconNameDocumentPrint granio.ThemeIconName = "documentPrint"

	// IconNameDocumentSave is the name of theme lookup for document save icon.
	//
	// Since: 2.0
	IconNameDocumentSave granio.ThemeIconName = "documentSave"

	// IconNameDragCornerIndicator is the name of the icon used in inner windows to indicate a draggable corner.
	//
	// Since: 2.5
	IconNameDragCornerIndicator granio.ThemeIconName = "dragCornerIndicator"

	// IconNameMoreHorizontal is the name of theme lookup for horizontal more.
	//
	// Since 2.0
	IconNameMoreHorizontal granio.ThemeIconName = "moreHorizontal"

	// IconNameMoreVertical is the name of theme lookup for vertical more.
	//
	// Since 2.0
	IconNameMoreVertical granio.ThemeIconName = "moreVertical"

	// IconNameMailAttachment is the name of theme lookup for mail attachment icon.
	//
	// Since: 2.0
	IconNameMailAttachment granio.ThemeIconName = "mailAttachment"

	// IconNameMailCompose is the name of theme lookup for mail compose icon.
	//
	// Since: 2.0
	IconNameMailCompose granio.ThemeIconName = "mailCompose"

	// IconNameMailForward is the name of theme lookup for mail forward icon.
	//
	// Since: 2.0
	IconNameMailForward granio.ThemeIconName = "mailForward"

	// IconNameMailReply is the name of theme lookup for mail reply icon.
	//
	// Since: 2.0
	IconNameMailReply granio.ThemeIconName = "mailReply"

	// IconNameMailReplyAll is the name of theme lookup for mail reply-all icon.
	//
	// Since: 2.0
	IconNameMailReplyAll granio.ThemeIconName = "mailReplyAll"

	// IconNameMailSend is the name of theme lookup for mail send icon.
	//
	// Since: 2.0
	IconNameMailSend granio.ThemeIconName = "mailSend"

	// IconNameMediaMusic is the name of theme lookup for media music icon.
	//
	// Since: 2.1
	IconNameMediaMusic granio.ThemeIconName = "mediaMusic"

	// IconNameMediaPhoto is the name of theme lookup for media photo icon.
	//
	// Since: 2.1
	IconNameMediaPhoto granio.ThemeIconName = "mediaPhoto"

	// IconNameMediaVideo is the name of theme lookup for media video icon.
	//
	// Since: 2.1
	IconNameMediaVideo granio.ThemeIconName = "mediaVideo"

	// IconNameMediaFastForward is the name of theme lookup for media fast-forward icon.
	//
	// Since: 2.0
	IconNameMediaFastForward granio.ThemeIconName = "mediaFastForward"

	// IconNameMediaFastRewind is the name of theme lookup for media fast-rewind icon.
	//
	// Since: 2.0
	IconNameMediaFastRewind granio.ThemeIconName = "mediaFastRewind"

	// IconNameMediaPause is the name of theme lookup for media pause icon.
	//
	// Since: 2.0
	IconNameMediaPause granio.ThemeIconName = "mediaPause"

	// IconNameMediaPlay is the name of theme lookup for media play icon.
	//
	// Since: 2.0
	IconNameMediaPlay granio.ThemeIconName = "mediaPlay"

	// IconNameMediaRecord is the name of theme lookup for media record icon.
	//
	// Since: 2.0
	IconNameMediaRecord granio.ThemeIconName = "mediaRecord"

	// IconNameMediaReplay is the name of theme lookup for media replay icon.
	//
	// Since: 2.0
	IconNameMediaReplay granio.ThemeIconName = "mediaReplay"

	// IconNameMediaSkipNext is the name of theme lookup for media skip next icon.
	//
	// Since: 2.0
	IconNameMediaSkipNext granio.ThemeIconName = "mediaSkipNext"

	// IconNameMediaSkipPrevious is the name of theme lookup for media skip previous icon.
	//
	// Since: 2.0
	IconNameMediaSkipPrevious granio.ThemeIconName = "mediaSkipPrevious"

	// IconNameMediaStop is the name of theme lookup for media stop icon.
	//
	// Since: 2.0
	IconNameMediaStop granio.ThemeIconName = "mediaStop"

	// IconNameMoveDown is the name of theme lookup for move down icon.
	//
	// Since: 2.0
	IconNameMoveDown granio.ThemeIconName = "arrowDown"

	// IconNameMoveUp is the name of theme lookup for move up icon.
	//
	// Since: 2.0
	IconNameMoveUp granio.ThemeIconName = "arrowUp"

	// IconNameNavigateBack is the name of theme lookup for navigate back icon.
	//
	// Since: 2.0
	IconNameNavigateBack granio.ThemeIconName = "arrowBack"

	// IconNameNavigateNext is the name of theme lookup for navigate next icon.
	//
	// Since: 2.0
	IconNameNavigateNext granio.ThemeIconName = "arrowForward"

	// IconNameArrowDropDown is the name of theme lookup for drop-down arrow icon.
	//
	// Since: 2.0
	IconNameArrowDropDown granio.ThemeIconName = "arrowDropDown"

	// IconNameArrowDropUp is the name of theme lookup for drop-up arrow icon.
	//
	// Since: 2.0
	IconNameArrowDropUp granio.ThemeIconName = "arrowDropUp"

	// IconNameFile is the name of theme lookup for file icon.
	//
	// Since: 2.0
	IconNameFile granio.ThemeIconName = "file"

	// IconNameFileApplication is the name of theme lookup for file application icon.
	//
	// Since: 2.0
	IconNameFileApplication granio.ThemeIconName = "fileApplication"

	// IconNameFileAudio is the name of theme lookup for file audio icon.
	//
	// Since: 2.0
	IconNameFileAudio granio.ThemeIconName = "fileAudio"

	// IconNameFileImage is the name of theme lookup for file image icon.
	//
	// Since: 2.0
	IconNameFileImage granio.ThemeIconName = "fileImage"

	// IconNameFileText is the name of theme lookup for file text icon.
	//
	// Since: 2.0
	IconNameFileText granio.ThemeIconName = "fileText"

	// IconNameFileVideo is the name of theme lookup for file video icon.
	//
	// Since: 2.0
	IconNameFileVideo granio.ThemeIconName = "fileVideo"

	// IconNameFolder is the name of theme lookup for folder icon.
	//
	// Since: 2.0
	IconNameFolder granio.ThemeIconName = "folder"

	// IconNameFolderNew is the name of theme lookup for folder new icon.
	//
	// Since: 2.0
	IconNameFolderNew granio.ThemeIconName = "folderNew"

	// IconNameFolderOpen is the name of theme lookup for folder open icon.
	//
	// Since: 2.0
	IconNameFolderOpen granio.ThemeIconName = "folderOpen"

	// IconNameHelp is the name of theme lookup for help icon.
	//
	// Since: 2.0
	IconNameHelp granio.ThemeIconName = "help"

	// IconNameHistory is the name of theme lookup for history icon.
	//
	// Since: 2.0
	IconNameHistory granio.ThemeIconName = "history"

	// IconNameHome is the name of theme lookup for home icon.
	//
	// Since: 2.0
	IconNameHome granio.ThemeIconName = "home"

	// IconNameSettings is the name of theme lookup for settings icon.
	//
	// Since: 2.0
	IconNameSettings granio.ThemeIconName = "settings"

	// IconNameStorage is the name of theme lookup for storage icon.
	//
	// Since: 2.0
	IconNameStorage granio.ThemeIconName = "storage"

	// IconNameUpload is the name of theme lookup for upload icon.
	//
	// Since: 2.0
	IconNameUpload granio.ThemeIconName = "upload"

	// IconNameViewFullScreen is the name of theme lookup for view fullscreen icon.
	//
	// Since: 2.0
	IconNameViewFullScreen granio.ThemeIconName = "viewFullScreen"

	// IconNameViewRefresh is the name of theme lookup for view refresh icon.
	//
	// Since: 2.0
	IconNameViewRefresh granio.ThemeIconName = "viewRefresh"

	// IconNameViewZoomFit is the name of theme lookup for view zoom fit icon.
	//
	// Since: 2.0
	IconNameViewZoomFit granio.ThemeIconName = "viewZoomFit"

	// IconNameViewZoomIn is the name of theme lookup for view zoom in icon.
	//
	// Since: 2.0
	IconNameViewZoomIn granio.ThemeIconName = "viewZoomIn"

	// IconNameViewZoomOut is the name of theme lookup for view zoom out icon.
	//
	// Since: 2.0
	IconNameViewZoomOut granio.ThemeIconName = "viewZoomOut"

	// IconNameViewRestore is the name of theme lookup for view restore icon.
	//
	// Since: 2.0
	IconNameViewRestore granio.ThemeIconName = "viewRestore"

	// IconNameVisibility is the name of theme lookup for visibility icon.
	//
	// Since: 2.0
	IconNameVisibility granio.ThemeIconName = "visibility"

	// IconNameVisibilityOff is the name of theme lookup for invisibility icon.
	//
	// Since: 2.0
	IconNameVisibilityOff granio.ThemeIconName = "visibilityOff"

	// IconNameVolumeDown is the name of theme lookup for volume down icon.
	//
	// Since: 2.0
	IconNameVolumeDown granio.ThemeIconName = "volumeDown"

	// IconNameVolumeMute is the name of theme lookup for volume mute icon.
	//
	// Since: 2.0
	IconNameVolumeMute granio.ThemeIconName = "volumeMute"

	// IconNameVolumeUp is the name of theme lookup for volume up icon.
	//
	// Since: 2.0
	IconNameVolumeUp granio.ThemeIconName = "volumeUp"

	// IconNameDownload is the name of theme lookup for download icon.
	//
	// Since: 2.0
	IconNameDownload granio.ThemeIconName = "download"

	// IconNameComputer is the name of theme lookup for computer icon.
	//
	// Since: 2.0
	IconNameComputer granio.ThemeIconName = "computer"

	// IconNameDesktop is the name of theme lookup for desktop icon.
	//
	// Since: 2.5
	IconNameDesktop granio.ThemeIconName = "desktop"

	// IconNameAccount is the name of theme lookup for account icon.
	//
	// Since: 2.1
	IconNameAccount granio.ThemeIconName = "account"

	// IconNameCalendar is the name of theme lookup for calendar icon.
	//
	// Since: 2.6
	IconNameCalendar granio.ThemeIconName = "calendar"

	// IconNameLogin is the name of theme lookup for login icon.
	//
	// Since: 2.1
	IconNameLogin granio.ThemeIconName = "login"

	// IconNameLogout is the name of theme lookup for logout icon.
	//
	// Since: 2.1
	IconNameLogout granio.ThemeIconName = "logout"

	// IconNameList is the name of theme lookup for list icon.
	//
	// Since: 2.1
	IconNameList granio.ThemeIconName = "list"

	// IconNameGrid is the name of theme lookup for grid icon.
	//
	// Since: 2.1
	IconNameGrid granio.ThemeIconName = "grid"

	// IconNameWindowClose is the name of theme lookup for window close icon.
	//
	// Since: 2.5
	IconNameWindowClose granio.ThemeIconName = "windowClose"

	// IconNameWindowMaximize is the name of theme lookup for window maximize icon.
	//
	// Since: 2.5
	IconNameWindowMaximize granio.ThemeIconName = "windowMaximize"

	// IconNameWindowMinimize is the name of theme lookup for window minimize icon.
	//
	// Since: 2.5
	IconNameWindowMinimize granio.ThemeIconName = "windowMinimize"
)

var icons = map[granio.ThemeIconName]granio.Resource{
	IconNameCancel:        NewThemedResource(cancelIconRes),
	IconNameConfirm:       NewThemedResource(checkIconRes),
	IconNameDelete:        NewThemedResource(deleteIconRes),
	IconNameSearch:        NewThemedResource(searchIconRes),
	IconNameSearchReplace: NewThemedResource(searchreplaceIconRes),
	IconNameMenu:          NewThemedResource(menuIconRes),
	IconNameMenuExpand:    NewThemedResource(menuexpandIconRes),

	IconNameCheckButton:        NewThemedResource(checkboxIconRes),
	IconNameCheckButtonChecked: NewThemedResource(checkboxcheckedIconRes),
	IconNameCheckButtonFill:    NewThemedResource(checkboxfillIconRes),
	IconNameCheckButtonPartial: NewThemedResource(checkboxpartialIconRes),
	IconNameRadioButton:        NewThemedResource(radiobuttonIconRes),
	IconNameRadioButtonChecked: NewThemedResource(radiobuttoncheckedIconRes),
	IconNameRadioButtonFill:    NewThemedResource(radiobuttonfillIconRes),

	IconNameContentAdd:    NewThemedResource(contentaddIconRes),
	IconNameContentClear:  NewThemedResource(cancelIconRes),
	IconNameContentRemove: NewThemedResource(contentremoveIconRes),
	IconNameContentCut:    NewThemedResource(contentcutIconRes),
	IconNameContentCopy:   NewThemedResource(contentcopyIconRes),
	IconNameContentPaste:  NewThemedResource(contentpasteIconRes),
	IconNameContentRedo:   NewThemedResource(contentredoIconRes),
	IconNameContentUndo:   NewThemedResource(contentundoIconRes),

	IconNameColorAchromatic: NewThemedResource(colorachromaticIconRes),
	IconNameColorChromatic:  NewThemedResource(colorchromaticIconRes),
	IconNameColorPalette:    NewThemedResource(colorpaletteIconRes),

	IconNameDocument:       NewThemedResource(documentIconRes),
	IconNameDocumentCreate: NewThemedResource(documentcreateIconRes),
	IconNameDocumentPrint:  NewThemedResource(documentprintIconRes),
	IconNameDocumentSave:   NewThemedResource(documentsaveIconRes),

	IconNameDragCornerIndicator: NewThemedResource(dragcornerindicatorIconRes),

	IconNameMoreHorizontal: NewThemedResource(morehorizontalIconRes),
	IconNameMoreVertical:   NewThemedResource(moreverticalIconRes),

	IconNameInfo:        NewThemedResource(infoIconRes),
	IconNameQuestion:    NewThemedResource(questionIconRes),
	IconNameWarning:     NewThemedResource(warningIconRes),
	IconNameError:       NewThemedResource(errorIconRes),
	IconNameBrokenImage: NewThemedResource(brokenimageIconRes),

	IconNameMailAttachment: NewThemedResource(mailattachmentIconRes),
	IconNameMailCompose:    NewThemedResource(mailcomposeIconRes),
	IconNameMailForward:    NewThemedResource(mailforwardIconRes),
	IconNameMailReply:      NewThemedResource(mailreplyIconRes),
	IconNameMailReplyAll:   NewThemedResource(mailreplyallIconRes),
	IconNameMailSend:       NewThemedResource(mailsendIconRes),

	IconNameMediaMusic:        NewThemedResource(mediamusicIconRes),
	IconNameMediaPhoto:        NewThemedResource(mediaphotoIconRes),
	IconNameMediaVideo:        NewThemedResource(mediavideoIconRes),
	IconNameMediaFastForward:  NewThemedResource(mediafastforwardIconRes),
	IconNameMediaFastRewind:   NewThemedResource(mediafastrewindIconRes),
	IconNameMediaPause:        NewThemedResource(mediapauseIconRes),
	IconNameMediaPlay:         NewThemedResource(mediaplayIconRes),
	IconNameMediaRecord:       NewThemedResource(mediarecordIconRes),
	IconNameMediaReplay:       NewThemedResource(mediareplayIconRes),
	IconNameMediaSkipNext:     NewThemedResource(mediaskipnextIconRes),
	IconNameMediaSkipPrevious: NewThemedResource(mediaskippreviousIconRes),
	IconNameMediaStop:         NewThemedResource(mediastopIconRes),

	IconNameNavigateBack:  NewThemedResource(arrowbackIconRes),
	IconNameMoveDown:      NewThemedResource(arrowdownIconRes),
	IconNameNavigateNext:  NewThemedResource(arrowforwardIconRes),
	IconNameMoveUp:        NewThemedResource(arrowupIconRes),
	IconNameArrowDropDown: NewThemedResource(arrowdropdownIconRes),
	IconNameArrowDropUp:   NewThemedResource(arrowdropupIconRes),

	IconNameFile:            NewThemedResource(fileIconRes),
	IconNameFileApplication: NewThemedResource(fileapplicationIconRes),
	IconNameFileAudio:       NewThemedResource(fileaudioIconRes),
	IconNameFileImage:       NewThemedResource(fileimageIconRes),
	IconNameFileText:        NewThemedResource(filetextIconRes),
	IconNameFileVideo:       NewThemedResource(filevideoIconRes),
	IconNameFolder:          NewThemedResource(folderIconRes),
	IconNameFolderNew:       NewThemedResource(foldernewIconRes),
	IconNameFolderOpen:      NewThemedResource(folderopenIconRes),
	IconNameHelp:            NewThemedResource(helpIconRes),
	IconNameHistory:         NewThemedResource(historyIconRes),
	IconNameHome:            NewThemedResource(homeIconRes),
	IconNameSettings:        NewThemedResource(settingsIconRes),

	IconNameViewFullScreen: NewThemedResource(viewfullscreenIconRes),
	IconNameViewRefresh:    NewThemedResource(viewrefreshIconRes),
	IconNameViewRestore:    NewThemedResource(viewzoomfitIconRes),
	IconNameViewZoomFit:    NewThemedResource(viewzoomfitIconRes),
	IconNameViewZoomIn:     NewThemedResource(viewzoominIconRes),
	IconNameViewZoomOut:    NewThemedResource(viewzoomoutIconRes),

	IconNameVisibility:    NewThemedResource(visibilityIconRes),
	IconNameVisibilityOff: NewThemedResource(visibilityoffIconRes),

	IconNameVolumeDown: NewThemedResource(volumedownIconRes),
	IconNameVolumeMute: NewThemedResource(volumemuteIconRes),
	IconNameVolumeUp:   NewThemedResource(volumeupIconRes),

	IconNameDownload: NewThemedResource(downloadIconRes),
	IconNameComputer: NewThemedResource(computerIconRes),
	IconNameDesktop:  NewThemedResource(desktopIconRes),
	IconNameStorage:  NewThemedResource(storageIconRes),
	IconNameUpload:   NewThemedResource(uploadIconRes),

	IconNameAccount:  NewThemedResource(accountIconRes),
	IconNameCalendar: NewThemedResource(calendarIconRes),
	IconNameLogin:    NewThemedResource(loginIconRes),
	IconNameLogout:   NewThemedResource(logoutIconRes),

	IconNameList: NewThemedResource(listIconRes),
	IconNameGrid: NewThemedResource(gridIconRes),

	IconNameWindowClose:    NewThemedResource(cancelIconRes),
	IconNameWindowMaximize: NewThemedResource(maximizeIconRes),
	IconNameWindowMinimize: NewThemedResource(minimizeIconRes),
}

// Icon looks up the specified icon for current theme.
//
// Since: 2.5
func Icon(name granio.ThemeIconName) granio.Resource {
	return safeIconLookup(name)
}

// IconForWidget looks up the specified icon for requested widget using the current theme.
// If the widget theme has been overridden that theme will be used.
//
// Since: 2.5
func IconForWidget(name granio.ThemeIconName, w granio.Widget) granio.Resource {
	return CurrentForWidget(w).Icon(name)
}

func (*builtinTheme) Icon(n granio.ThemeIconName) granio.Resource {
	return icons[n]
}

var _ granio.ThemedResource = (*ThemedResource)(nil)

// ThemedResource is a resource wrapper that will return a version of the resource with the main color changed
// for the currently selected theme.
type ThemedResource struct {
	source granio.Resource

	// ColorName specifies which theme colour should be used to theme the resource
	//
	// Since: 2.3
	ColorName granio.ThemeColorName
}

// NewColoredResource creates a resource that adapts to the current theme setting using
// the color named in the constructor.
//
// Since: 2.4
func NewColoredResource(src granio.Resource, name granio.ThemeColorName) *ThemedResource {
	return &ThemedResource{
		source:    src,
		ColorName: name,
	}
}

// NewSuccessThemedResource creates a resource that adapts to the current theme success color.
//
// Since: 2.4
func NewSuccessThemedResource(src granio.Resource) *ThemedResource {
	return &ThemedResource{
		source:    src,
		ColorName: ColorNameSuccess,
	}
}

// NewThemedResource creates a resource that adapts to the current theme setting.
// By default, this will match the foreground color, but it can be changed using the `ColorName` field.
func NewThemedResource(src granio.Resource) *ThemedResource {
	return &ThemedResource{
		source: src,
	}
}

// NewWarningThemedResource creates a resource that adapts to the current theme warning color.
//
// Since: 2.4
func NewWarningThemedResource(src granio.Resource) *ThemedResource {
	return &ThemedResource{
		source:    src,
		ColorName: ColorNameWarning,
	}
}

// Name returns the underlying resource name (used for caching).
func (res *ThemedResource) Name() string {
	return string(res.ThemeColorName()) + "_" + unwrapResource(res.source).Name()
}

// ThemeColorName returns the granio.ThemeColorName that is used as foreground color.
func (res *ThemedResource) ThemeColorName() granio.ThemeColorName {
	if res.ColorName != "" {
		return res.ColorName
	}

	return ColorNameForeground
}

// Content returns the underlying content of the resource adapted to the current text color.
func (res *ThemedResource) Content() []byte {
	return colorizeLogError(unwrapResource(res.source).Content(), Color(res.ThemeColorName()))
}

// Error returns a different resource for indicating an error.
func (res *ThemedResource) Error() *ErrorThemedResource {
	return NewErrorThemedResource(res)
}

var _ granio.ThemedResource = (*InvertedThemedResource)(nil)

// InvertedThemedResource is a resource wrapper that will return a version of the resource with the main color changed
// for use over highlighted elements.
type InvertedThemedResource struct {
	source granio.Resource
}

// NewInvertedThemedResource creates a resource that adapts to the current theme for use over highlighted elements.
func NewInvertedThemedResource(orig granio.Resource) *InvertedThemedResource {
	res := &InvertedThemedResource{source: orig}
	return res
}

// Name returns the underlying resource name (used for caching).
func (res *InvertedThemedResource) Name() string {
	return "inverted_" + unwrapResource(res.source).Name()
}

// Content returns the underlying content of the resource adapted to the current background color.
func (res *InvertedThemedResource) Content() []byte {
	clr := Color(ColorNameBackground)
	return colorizeLogError(unwrapResource(res.source).Content(), clr)
}

// ThemeColorName returns the granio.ThemeColorName that is used as foreground color.
func (*InvertedThemedResource) ThemeColorName() granio.ThemeColorName {
	return ColorNameBackground
}

// Original returns the underlying resource that this inverted themed resource was adapted from
func (res *InvertedThemedResource) Original() granio.Resource {
	return res.source
}

var _ granio.ThemedResource = (*ErrorThemedResource)(nil)

// ErrorThemedResource is a resource wrapper that will return a version of the resource with the main color changed
// to indicate an error.
type ErrorThemedResource struct {
	source granio.Resource
}

// NewErrorThemedResource creates a resource that adapts to the error color for the current theme.
func NewErrorThemedResource(orig granio.Resource) *ErrorThemedResource {
	res := &ErrorThemedResource{source: orig}
	return res
}

// Name returns the underlying resource name (used for caching).
func (res *ErrorThemedResource) Name() string {
	return "error_" + unwrapResource(res.source).Name()
}

// Content returns the underlying content of the resource adapted to the current background color.
func (res *ErrorThemedResource) Content() []byte {
	return colorizeLogError(unwrapResource(res.source).Content(), Color(ColorNameError))
}

// Original returns the underlying resource that this error themed resource was adapted from
func (res *ErrorThemedResource) Original() granio.Resource {
	return res.source
}

// ThemeColorName returns the granio.ThemeColorName that is used as foreground color.
//
// Since: 2.6
func (*ErrorThemedResource) ThemeColorName() granio.ThemeColorName {
	return ColorNameError
}

var _ granio.ThemedResource = (*PrimaryThemedResource)(nil)

// PrimaryThemedResource is a resource wrapper that will return a version of the resource with the main color changed
// to the theme primary color.
type PrimaryThemedResource struct {
	source granio.Resource
}

// NewPrimaryThemedResource creates a resource that adapts to the primary color for the current theme.
func NewPrimaryThemedResource(orig granio.Resource) *PrimaryThemedResource {
	return &PrimaryThemedResource{source: orig}
}

// Name returns the underlying resource name (used for caching).
func (res *PrimaryThemedResource) Name() string {
	return "primary_" + unwrapResource(res.source).Name()
}

// Content returns the underlying content of the resource adapted to the current background color.
func (res *PrimaryThemedResource) Content() []byte {
	return colorizeLogError(unwrapResource(res.source).Content(), Color(ColorNamePrimary))
}

// Original returns the underlying resource that this primary themed resource was adapted from
func (res *PrimaryThemedResource) Original() granio.Resource {
	return res.source
}

// ThemeColorName returns the granio.ThemeColorName that is used as foreground color.
//
// Since: 2.6
func (*PrimaryThemedResource) ThemeColorName() granio.ThemeColorName {
	return ColorNamePrimary
}

var _ granio.ThemedResource = (*DisabledResource)(nil)

// DisabledResource is a resource wrapper that will return an appropriate resource colorized by
// the current theme's `DisabledColor` color.
type DisabledResource struct {
	source granio.Resource
}

// Name returns the resource source name prefixed with `disabled_` (used for caching)
func (res *DisabledResource) Name() string {
	return "disabled_" + unwrapResource(res.source).Name()
}

// Content returns the disabled style content of the correct resource for the current theme
func (res *DisabledResource) Content() []byte {
	return colorizeLogError(unwrapResource(res.source).Content(), Color(ColorNameDisabled))
}

// ThemeColorName returns the granio.ThemeColorName that is used as foreground color.
//
// Since: 2.6
func (*DisabledResource) ThemeColorName() granio.ThemeColorName {
	return ColorNameDisabled
}

// NewDisabledResource creates a resource that adapts to the current theme's DisabledColor setting.
func NewDisabledResource(res granio.Resource) *DisabledResource {
	return &DisabledResource{
		source: res,
	}
}

// FyneLogo returns a resource containing the Fyne logo.
//
// Deprecated: Applications should use their own icon in most cases.
func FyneLogo() granio.Resource {
	return fynelogo
}

// CancelIcon returns a resource containing the standard cancel icon for the current theme
func CancelIcon() granio.Resource {
	return safeIconLookup(IconNameCancel)
}

// ConfirmIcon returns a resource containing the standard confirm icon for the current theme
func ConfirmIcon() granio.Resource {
	return safeIconLookup(IconNameConfirm)
}

// DeleteIcon returns a resource containing the standard delete icon for the current theme
func DeleteIcon() granio.Resource {
	return safeIconLookup(IconNameDelete)
}

// SearchIcon returns a resource containing the standard search icon for the current theme
func SearchIcon() granio.Resource {
	return safeIconLookup(IconNameSearch)
}

// SearchReplaceIcon returns a resource containing the standard search and replace icon for the current theme
func SearchReplaceIcon() granio.Resource {
	return safeIconLookup(IconNameSearchReplace)
}

// MenuIcon returns a resource containing the standard (mobile) menu icon for the current theme
func MenuIcon() granio.Resource {
	return safeIconLookup(IconNameMenu)
}

// MenuExpandIcon returns a resource containing the standard (mobile) expand "submenu icon for the current theme
func MenuExpandIcon() granio.Resource {
	return safeIconLookup(IconNameMenuExpand)
}

// CheckButtonIcon returns a resource containing the standard checkbox icon for the current theme
func CheckButtonIcon() granio.Resource {
	return safeIconLookup(IconNameCheckButton)
}

// CheckButtonCheckedIcon returns a resource containing the standard checkbox checked icon for the current theme
func CheckButtonCheckedIcon() granio.Resource {
	return safeIconLookup(IconNameCheckButtonChecked)
}

// CheckButtonFillIcon returns a resource containing the filled checkbox icon for the current theme.
//
// Since: 2.5
func CheckButtonFillIcon() granio.Resource {
	return safeIconLookup(IconNameCheckButtonFill)
}

// RadioButtonIcon returns a resource containing the standard radio button icon for the current theme
func RadioButtonIcon() granio.Resource {
	return safeIconLookup(IconNameRadioButton)
}

// RadioButtonCheckedIcon returns a resource containing the standard radio button checked icon for the current theme
func RadioButtonCheckedIcon() granio.Resource {
	return safeIconLookup(IconNameRadioButtonChecked)
}

// RadioButtonFillIcon returns a resource containing the filled checkbox icon for the current theme.
//
// Since: 2.5
func RadioButtonFillIcon() granio.Resource {
	return safeIconLookup(IconNameRadioButtonFill)
}

// ContentAddIcon returns a resource containing the standard content add icon for the current theme
func ContentAddIcon() granio.Resource {
	return safeIconLookup(IconNameContentAdd)
}

// ContentRemoveIcon returns a resource containing the standard content remove icon for the current theme
func ContentRemoveIcon() granio.Resource {
	return safeIconLookup(IconNameContentRemove)
}

// ContentClearIcon returns a resource containing the standard content clear icon for the current theme
func ContentClearIcon() granio.Resource {
	return safeIconLookup(IconNameContentClear)
}

// ContentCutIcon returns a resource containing the standard content cut icon for the current theme
func ContentCutIcon() granio.Resource {
	return safeIconLookup(IconNameContentCut)
}

// ContentCopyIcon returns a resource containing the standard content copy icon for the current theme
func ContentCopyIcon() granio.Resource {
	return safeIconLookup(IconNameContentCopy)
}

// ContentPasteIcon returns a resource containing the standard content paste icon for the current theme
func ContentPasteIcon() granio.Resource {
	return safeIconLookup(IconNameContentPaste)
}

// ContentRedoIcon returns a resource containing the standard content redo icon for the current theme
func ContentRedoIcon() granio.Resource {
	return safeIconLookup(IconNameContentRedo)
}

// ContentUndoIcon returns a resource containing the standard content undo icon for the current theme
func ContentUndoIcon() granio.Resource {
	return safeIconLookup(IconNameContentUndo)
}

// ColorAchromaticIcon returns a resource containing the standard achromatic color icon for the current theme
func ColorAchromaticIcon() granio.Resource {
	return safeIconLookup(IconNameColorAchromatic)
}

// ColorChromaticIcon returns a resource containing the standard chromatic color icon for the current theme
func ColorChromaticIcon() granio.Resource {
	return safeIconLookup(IconNameColorChromatic)
}

// ColorPaletteIcon returns a resource containing the standard color palette icon for the current theme
func ColorPaletteIcon() granio.Resource {
	return safeIconLookup(IconNameColorPalette)
}

// DocumentIcon returns a resource containing the standard document icon for the current theme
func DocumentIcon() granio.Resource {
	return safeIconLookup(IconNameDocument)
}

// DocumentCreateIcon returns a resource containing the standard document create icon for the current theme
func DocumentCreateIcon() granio.Resource {
	return safeIconLookup(IconNameDocumentCreate)
}

// DocumentPrintIcon returns a resource containing the standard document print icon for the current theme
func DocumentPrintIcon() granio.Resource {
	return safeIconLookup(IconNameDocumentPrint)
}

// DocumentSaveIcon returns a resource containing the standard document save icon for the current theme
func DocumentSaveIcon() granio.Resource {
	return safeIconLookup(IconNameDocumentSave)
}

// MoreHorizontalIcon returns a resource containing the standard horizontal more icon for the current theme
func MoreHorizontalIcon() granio.Resource {
	return Current().Icon(IconNameMoreHorizontal)
}

// MoreVerticalIcon returns a resource containing the standard vertical more icon for the current theme
func MoreVerticalIcon() granio.Resource {
	return Current().Icon(IconNameMoreVertical)
}

// InfoIcon returns a resource containing the standard dialog info icon for the current theme
func InfoIcon() granio.Resource {
	return safeIconLookup(IconNameInfo)
}

// QuestionIcon returns a resource containing the standard dialog question icon for the current theme
func QuestionIcon() granio.Resource {
	return safeIconLookup(IconNameQuestion)
}

// WarningIcon returns a resource containing the standard dialog warning icon for the current theme
func WarningIcon() granio.Resource {
	return safeIconLookup(IconNameWarning)
}

// ErrorIcon returns a resource containing the standard dialog error icon for the current theme
func ErrorIcon() granio.Resource {
	return safeIconLookup(IconNameError)
}

// BrokenImageIcon returns a resource containing an icon to specify a broken or missing image
//
// Since: 2.4
func BrokenImageIcon() granio.Resource {
	return safeIconLookup(IconNameBrokenImage)
}

// FileIcon returns a resource containing the appropriate file icon for the current theme
func FileIcon() granio.Resource {
	return safeIconLookup(IconNameFile)
}

// FileApplicationIcon returns a resource containing the file icon representing application files for the current theme
func FileApplicationIcon() granio.Resource {
	return safeIconLookup(IconNameFileApplication)
}

// FileAudioIcon returns a resource containing the file icon representing audio files for the current theme
func FileAudioIcon() granio.Resource {
	return safeIconLookup(IconNameFileAudio)
}

// FileImageIcon returns a resource containing the file icon representing image files for the current theme
func FileImageIcon() granio.Resource {
	return safeIconLookup(IconNameFileImage)
}

// FileTextIcon returns a resource containing the file icon representing text files for the current theme
func FileTextIcon() granio.Resource {
	return safeIconLookup(IconNameFileText)
}

// FileVideoIcon returns a resource containing the file icon representing video files for the current theme
func FileVideoIcon() granio.Resource {
	return safeIconLookup(IconNameFileVideo)
}

// FolderIcon returns a resource containing the standard folder icon for the current theme
func FolderIcon() granio.Resource {
	return safeIconLookup(IconNameFolder)
}

// FolderNewIcon returns a resource containing the standard folder creation icon for the current theme
func FolderNewIcon() granio.Resource {
	return safeIconLookup(IconNameFolderNew)
}

// FolderOpenIcon returns a resource containing the standard folder open icon for the current theme
func FolderOpenIcon() granio.Resource {
	return safeIconLookup(IconNameFolderOpen)
}

// HelpIcon returns a resource containing the standard help icon for the current theme
func HelpIcon() granio.Resource {
	return safeIconLookup(IconNameHelp)
}

// HistoryIcon returns a resource containing the standard history icon for the current theme
func HistoryIcon() granio.Resource {
	return safeIconLookup(IconNameHistory)
}

// HomeIcon returns a resource containing the standard home folder icon for the current theme
func HomeIcon() granio.Resource {
	return safeIconLookup(IconNameHome)
}

// SettingsIcon returns a resource containing the standard settings icon for the current theme
func SettingsIcon() granio.Resource {
	return safeIconLookup(IconNameSettings)
}

// MailAttachmentIcon returns a resource containing the standard mail attachment icon for the current theme
func MailAttachmentIcon() granio.Resource {
	return safeIconLookup(IconNameMailAttachment)
}

// MailComposeIcon returns a resource containing the standard mail compose icon for the current theme
func MailComposeIcon() granio.Resource {
	return safeIconLookup(IconNameMailCompose)
}

// MailForwardIcon returns a resource containing the standard mail forward icon for the current theme
func MailForwardIcon() granio.Resource {
	return safeIconLookup(IconNameMailForward)
}

// MailReplyIcon returns a resource containing the standard mail reply icon for the current theme
func MailReplyIcon() granio.Resource {
	return safeIconLookup(IconNameMailReply)
}

// MailReplyAllIcon returns a resource containing the standard mail reply all icon for the current theme
func MailReplyAllIcon() granio.Resource {
	return safeIconLookup(IconNameMailReplyAll)
}

// MailSendIcon returns a resource containing the standard mail send icon for the current theme
func MailSendIcon() granio.Resource {
	return safeIconLookup(IconNameMailSend)
}

// MediaMusicIcon returns a resource containing the standard media music icon for the current theme
//
// Since: 2.1
func MediaMusicIcon() granio.Resource {
	return safeIconLookup(IconNameMediaMusic)
}

// MediaPhotoIcon returns a resource containing the standard media photo icon for the current theme
//
// Since: 2.1
func MediaPhotoIcon() granio.Resource {
	return safeIconLookup(IconNameMediaPhoto)
}

// MediaVideoIcon returns a resource containing the standard media video icon for the current theme
//
// Since: 2.1
func MediaVideoIcon() granio.Resource {
	return safeIconLookup(IconNameMediaVideo)
}

// MediaFastForwardIcon returns a resource containing the standard media fast-forward icon for the current theme
func MediaFastForwardIcon() granio.Resource {
	return safeIconLookup(IconNameMediaFastForward)
}

// MediaFastRewindIcon returns a resource containing the standard media fast-rewind icon for the current theme
func MediaFastRewindIcon() granio.Resource {
	return safeIconLookup(IconNameMediaFastRewind)
}

// MediaPauseIcon returns a resource containing the standard media pause icon for the current theme
func MediaPauseIcon() granio.Resource {
	return safeIconLookup(IconNameMediaPause)
}

// MediaPlayIcon returns a resource containing the standard media play icon for the current theme
func MediaPlayIcon() granio.Resource {
	return safeIconLookup(IconNameMediaPlay)
}

// MediaRecordIcon returns a resource containing the standard media record icon for the current theme
func MediaRecordIcon() granio.Resource {
	return safeIconLookup(IconNameMediaRecord)
}

// MediaReplayIcon returns a resource containing the standard media replay icon for the current theme
func MediaReplayIcon() granio.Resource {
	return safeIconLookup(IconNameMediaReplay)
}

// MediaSkipNextIcon returns a resource containing the standard media skip next icon for the current theme
func MediaSkipNextIcon() granio.Resource {
	return safeIconLookup(IconNameMediaSkipNext)
}

// MediaSkipPreviousIcon returns a resource containing the standard media skip previous icon for the current theme
func MediaSkipPreviousIcon() granio.Resource {
	return safeIconLookup(IconNameMediaSkipPrevious)
}

// MediaStopIcon returns a resource containing the standard media stop icon for the current theme
func MediaStopIcon() granio.Resource {
	return safeIconLookup(IconNameMediaStop)
}

// MoveDownIcon returns a resource containing the standard down arrow icon for the current theme
func MoveDownIcon() granio.Resource {
	return safeIconLookup(IconNameMoveDown)
}

// MoveUpIcon returns a resource containing the standard up arrow icon for the current theme
func MoveUpIcon() granio.Resource {
	return safeIconLookup(IconNameMoveUp)
}

// NavigateBackIcon returns a resource containing the standard backward navigation icon for the current theme
func NavigateBackIcon() granio.Resource {
	return safeIconLookup(IconNameNavigateBack)
}

// NavigateNextIcon returns a resource containing the standard forward navigation icon for the current theme
func NavigateNextIcon() granio.Resource {
	return safeIconLookup(IconNameNavigateNext)
}

// MenuDropDownIcon returns a resource containing the standard menu drop down icon for the current theme
func MenuDropDownIcon() granio.Resource {
	return safeIconLookup(IconNameArrowDropDown)
}

// MenuDropUpIcon returns a resource containing the standard menu drop up icon for the current theme
func MenuDropUpIcon() granio.Resource {
	return safeIconLookup(IconNameArrowDropUp)
}

// ViewFullScreenIcon returns a resource containing the standard fullscreen icon for the current theme
func ViewFullScreenIcon() granio.Resource {
	return safeIconLookup(IconNameViewFullScreen)
}

// ViewRestoreIcon returns a resource containing the standard exit fullscreen icon for the current theme
func ViewRestoreIcon() granio.Resource {
	return safeIconLookup(IconNameViewRestore)
}

// ViewRefreshIcon returns a resource containing the standard refresh icon for the current theme
func ViewRefreshIcon() granio.Resource {
	return safeIconLookup(IconNameViewRefresh)
}

// ZoomFitIcon returns a resource containing the standard zoom fit icon for the current theme
func ZoomFitIcon() granio.Resource {
	return safeIconLookup(IconNameViewZoomFit)
}

// ZoomInIcon returns a resource containing the standard zoom in icon for the current theme
func ZoomInIcon() granio.Resource {
	return safeIconLookup(IconNameViewZoomIn)
}

// ZoomOutIcon returns a resource containing the standard zoom out icon for the current theme
func ZoomOutIcon() granio.Resource {
	return safeIconLookup(IconNameViewZoomOut)
}

// VisibilityIcon returns a resource containing the standard visibility icon for the current theme
func VisibilityIcon() granio.Resource {
	return safeIconLookup(IconNameVisibility)
}

// VisibilityOffIcon returns a resource containing the standard visibility off icon for the current theme
func VisibilityOffIcon() granio.Resource {
	return safeIconLookup(IconNameVisibilityOff)
}

// VolumeDownIcon returns a resource containing the standard volume down icon for the current theme
func VolumeDownIcon() granio.Resource {
	return safeIconLookup(IconNameVolumeDown)
}

// VolumeMuteIcon returns a resource containing the standard volume mute icon for the current theme
func VolumeMuteIcon() granio.Resource {
	return safeIconLookup(IconNameVolumeMute)
}

// VolumeUpIcon returns a resource containing the standard volume up icon for the current theme
func VolumeUpIcon() granio.Resource {
	return safeIconLookup(IconNameVolumeUp)
}

// ComputerIcon returns a resource containing the standard computer icon for the current theme
func ComputerIcon() granio.Resource {
	return safeIconLookup(IconNameComputer)
}

// DesktopIcon returns a resource containing the standard desktop icon for the current theme
func DesktopIcon() granio.Resource {
	return safeIconLookup(IconNameDesktop)
}

// DownloadIcon returns a resource containing the standard download icon for the current theme
func DownloadIcon() granio.Resource {
	return safeIconLookup(IconNameDownload)
}

// StorageIcon returns a resource containing the standard storage icon for the current theme
func StorageIcon() granio.Resource {
	return safeIconLookup(IconNameStorage)
}

// UploadIcon returns a resource containing the standard upload icon for the current theme
func UploadIcon() granio.Resource {
	return safeIconLookup(IconNameUpload)
}

// AccountIcon returns a resource containing the standard account icon for the current theme
func AccountIcon() granio.Resource {
	return safeIconLookup(IconNameAccount)
}

// CalendarIcon returns a resource containing the standard account icon for the current theme
//
// Since: 2.6
func CalendarIcon() granio.Resource {
	return safeIconLookup(IconNameCalendar)
}

// LoginIcon returns a resource containing the standard login icon for the current theme
func LoginIcon() granio.Resource {
	return safeIconLookup(IconNameLogin)
}

// LogoutIcon returns a resource containing the standard logout icon for the current theme
func LogoutIcon() granio.Resource {
	return safeIconLookup(IconNameLogout)
}

// ListIcon returns a resource containing the standard list icon for the current theme
func ListIcon() granio.Resource {
	return safeIconLookup(IconNameList)
}

// GridIcon returns a resource containing the standard grid icon for the current theme
func GridIcon() granio.Resource {
	return safeIconLookup(IconNameGrid)
}

// WindowCloseIcon returns a resource containing the window close icon for the current theme
//
// Since: 2.5
func WindowCloseIcon() granio.Resource {
	return safeIconLookup(IconNameWindowClose)
}

// WindowMaximizeIcon returns a resource containing the window maximize icon for the current theme
//
// Since: 2.5
func WindowMaximizeIcon() granio.Resource {
	return safeIconLookup(IconNameWindowMaximize)
}

// WindowMinimizeIcon returns a resource containing the window minimize icon for the current theme
//
// Since: 2.5
func WindowMinimizeIcon() granio.Resource {
	return safeIconLookup(IconNameWindowMinimize)
}

func safeIconLookup(n granio.ThemeIconName) granio.Resource {
	icon := Current().Icon(n)
	if icon == nil {
		granio.LogError("Loaded theme returned nil icon", nil)
		return fallbackIcon
	}
	return icon
}

// recursively "unwraps" the source of the given resource to extract
// the "base" resource - to avoid recolorizing SVG multiple times
// if we for example have a ThemedResource wrapped in an ErrorThemedResource
func unwrapResource(res granio.Resource) granio.Resource {
	for {
		switch typedRes := res.(type) {
		case *DisabledResource:
			res = typedRes.source
		case *ErrorThemedResource:
			res = typedRes.source
		case *InvertedThemedResource:
			res = typedRes.source
		case *PrimaryThemedResource:
			res = typedRes.source
		case *ThemedResource:
			res = typedRes.source
		default:
			return res
		}
	}
}

func colorizeLogError(src []byte, clr color.Color) []byte {
	content, err := svg.Colorize(src, clr)
	if err != nil {
		granio.LogError("", err)
	}
	return content
}

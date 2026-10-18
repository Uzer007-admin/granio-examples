//go:generate granio bundle -package data -name fynelogo -o bundled.go fyne_logo.png
//go:generate granio bundle -package data -o bundled.go --append Icon.png

package data

// FyneLogo contains the full fyne logo with background design
var FyneLogo = fynelogo

// Icon contains the app icon for use in window borders
var Icon = resourceIconPng

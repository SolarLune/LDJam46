package main

import (
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

var ImageResources = map[string]*ebiten.Image{}

func GetImage(filepath string) *ebiten.Image {

	res, exists := ImageResources[filepath]

	if !exists {
		res, _, _ = ebitenutil.NewImageFromFile(getPath(filepath), ebiten.FilterNearest)
		ImageResources[filepath] = res
	}

	return res

}

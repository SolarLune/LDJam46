package main

import (
	"path/filepath"

	"github.com/SolarLune/goaseprite"
	"github.com/hajimehoshi/ebiten"
)

const TypeAnimationComponent = "Animation"

type AnimationComponent struct {
	AnimPath  string
	Image     *ebiten.Image
	Ase       *goaseprite.File
	OnAnimEnd func(*AnimationComponent)
}

func NewAnimationComponent(animPath string) *AnimationComponent {
	a := &AnimationComponent{AnimPath: animPath}
	a.Ase = goaseprite.Open(a.AnimPath)
	a.Image = GetImage(filepath.Join(filepath.Dir(a.AnimPath), a.Ase.ImagePath))
	return a
}

func (a *AnimationComponent) OnAdd(g *GameObject) {}

func (a *AnimationComponent) OnRemove(g *GameObject) {}

func (a *AnimationComponent) Update(screen *ebiten.Image) {

	a.Ase.Update(1.0 / 60.0)

	if a.Ase.FinishedAnimation && a.OnAnimEnd != nil {
		a.OnAnimEnd(a)
	}

}

func (a *AnimationComponent) Type() string { return TypeAnimationComponent }

//

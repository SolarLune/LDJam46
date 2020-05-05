package main

import (
	"image/color"
	"math"

	"github.com/SolarLune/resolv"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/kvartborg/vector"
)

const TypeBodyComponent = "Body"

type BodyComponent struct {
	GameObject *GameObject
	Speed      vector.Vector
	Object     *resolv.Object
	OnBump     func(*BodyComponent)
}

func NewBodyComponent(x, y, w, h float64, space *resolv.Space) *BodyComponent {

	return &BodyComponent{
		Speed:  vector.Vector{0, 0},
		Object: resolv.NewObject(x, y, w, h, space),
	}

}

func (b *BodyComponent) OnAdd(g *GameObject) {
	b.GameObject = g
}

func (b *BodyComponent) OnRemove(g *GameObject) { b.Object.Remove() }

func (b *BodyComponent) Update(screen *ebiten.Image) {

	if col := b.Object.Check(b.Speed[0], 0, "solid"); col.Valid() {
		if b.OnBump != nil {
			b.OnBump(b)
		}
		b.Object.X += float64(col.ContactX)
		if col.CanSlide && math.Abs(col.SlideY) < 4 {
			b.Object.Y += col.SlideY
		} else {
			b.Speed[0] = 0
		}
	} else {
		b.Object.X += b.Speed[0]
	}

	if col := b.Object.Check(0, b.Speed[1], "solid"); col.Valid() {
		if b.OnBump != nil {
			b.OnBump(b)
		}
		b.Object.Y += float64(col.ContactY)
		if col.CanSlide && math.Abs(col.SlideX) < 4 {
			b.Object.X += col.SlideX
		} else {
			b.Speed[1] = 0
		}
	} else {
		b.Object.Y += b.Speed[1]
	}

	b.Object.Update()

	if b.GameObject.Level.Game.DebugMode {

		x, y := b.Object.X-b.GameObject.Level.CameraOffsetX, b.Object.Y-b.GameObject.Level.CameraOffsetY
		w, h := b.Object.W, b.Object.H

		red := color.RGBA{255, 0, 0, 192}
		ebitenutil.DrawLine(screen, x, y, x+w, y, red)
		ebitenutil.DrawLine(screen, x+w, y, x+w, y+h, red)
		ebitenutil.DrawLine(screen, x+w, y+h, x, y+h, red)
		ebitenutil.DrawLine(screen, x, y+h, x, y, red)
		// ebitenutil.DrawRect(screen, float64(rect.X)-level.CameraOffsetX, float64(rect.Y)-level.CameraOffsetY, float64(rect.W), float64(rect.H), color.RGBA{uintX, uintY, 0, 128})

	}

}

func (b *BodyComponent) Type() string { return TypeBodyComponent }

func (b *BodyComponent) Center() vector.Vector {
	return vector.Vector{
		b.Object.X + (b.Object.W / 2),
		b.Object.Y + (b.Object.H / 2),
	}
}

package main

import (
	"image"

	"github.com/hajimehoshi/ebiten"
	"github.com/kvartborg/vector"
)

const TypeDrawComponent = "Draw"

type DrawComponent struct {
	GameObject     *GameObject
	FlipHorizontal bool
	Rotation       float64
	Offset         vector.Vector
	Visible        bool
}

func NewDrawComponent(offsetX, offsetY float64) *DrawComponent {
	return &DrawComponent{Offset: vector.Vector{offsetX, offsetY}, Visible: true}
}

func (d *DrawComponent) OnAdd(g *GameObject) {
	d.GameObject = g
}

func (d *DrawComponent) OnRemove(g *GameObject) {}

func (d *DrawComponent) Update(screen *ebiten.Image) {

	if d.Visible && d.GameObject.GetComponent(TypeAnimationComponent) != nil {

		geoM := ebiten.GeoM{}
		anim := d.GameObject.GetComponent(TypeAnimationComponent).(*AnimationComponent)
		bodyX := d.Offset[0]
		bodyY := d.Offset[1]
		x, y := anim.Ase.GetFrameXY()
		img := anim.Image.SubImage(image.Rect(int(x), int(y), int(x+anim.Ase.FrameWidth), int(y+anim.Ase.FrameHeight))).(*ebiten.Image)
		srcW, srcH := img.Size()

		if d.GameObject.GetComponent(TypeBodyComponent) != nil {
			body := d.GameObject.GetComponent(TypeBodyComponent).(*BodyComponent)
			bodyX += body.Object.X - ((float64(srcW) - body.Object.W) / 2)
			bodyY += body.Object.Y - ((float64(srcH) - body.Object.H) / 2)
		}

		if d.Rotation != 0 {

			geoM.Translate(-float64(anim.Ase.FrameWidth/2), -float64(anim.Ase.FrameHeight/2))
			geoM.Rotate(d.Rotation)
			geoM.Translate(float64(anim.Ase.FrameWidth/2), float64(anim.Ase.FrameHeight/2))

		}

		colorM := ebiten.ColorM{}

		colorM.ChangeHSV(0, 1, 0)

		if d.FlipHorizontal {

			geoM.Translate(-float64(srcW/2), -float64(srcH/2))
			geoM.Scale(-1, 1)
			geoM.Translate(float64(srcW/2), float64(srcH/2))

		}

		geoM.Translate(bodyX-d.GameObject.Level.CameraOffsetX, bodyY-d.GameObject.Level.CameraOffsetY)

		screen.DrawImage(img, &ebiten.DrawImageOptions{GeoM: geoM})

	}

}

func (d *DrawComponent) Type() string { return TypeDrawComponent }

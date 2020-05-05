package main

import "github.com/hajimehoshi/ebiten"

const TypeCameraFollowComponent = "CameraFollow"

type CameraFollowComponent struct {
	GameObject       *GameObject
	OffsetX, OffsetY float64
	Softness         float64
}

func NewCameraFollowComponent() *CameraFollowComponent {
	return &CameraFollowComponent{Softness: 0.1}
}

func (cf *CameraFollowComponent) OnAdd(g *GameObject) {
	cf.GameObject = g
	cf.OffsetX = float64(cf.GameObject.Level.Game.Width / 2)
	cf.OffsetY = float64(cf.GameObject.Level.Game.Height / 2)

}
func (cf *CameraFollowComponent) OnRemove(g *GameObject) {}

func (cf *CameraFollowComponent) Update(screen *ebiten.Image) {

	b := cf.GameObject.GetComponent(TypeBodyComponent)

	if b != nil {

		body := b.(*BodyComponent)

		tx := body.Object.X - cf.OffsetX - cf.GameObject.Level.CameraOffsetX
		cf.GameObject.Level.CameraOffsetX += tx * cf.Softness

		ty := body.Object.Y - cf.OffsetY - cf.GameObject.Level.CameraOffsetY
		cf.GameObject.Level.CameraOffsetY += ty * cf.Softness

		if cf.GameObject.Level.CameraOffsetX < 0 {
			cf.GameObject.Level.CameraOffsetX = 0
		}
		if cf.GameObject.Level.CameraOffsetY < 0 {
			cf.GameObject.Level.CameraOffsetY = 0
		}

		if cf.GameObject.Level.CameraOffsetX+float64(cf.GameObject.Level.Game.Width) > float64(cf.GameObject.Level.Width()) {
			cf.GameObject.Level.CameraOffsetX = float64(cf.GameObject.Level.Width() - cf.GameObject.Level.Game.Width)
		}

		if cf.GameObject.Level.CameraOffsetY+float64(cf.GameObject.Level.Game.Height) > float64(cf.GameObject.Level.Height()) {
			cf.GameObject.Level.CameraOffsetY = float64(cf.GameObject.Level.Height() - cf.GameObject.Level.Game.Height)
		}

	}

}

func (cf *CameraFollowComponent) Type() string { return TypeCameraFollowComponent }

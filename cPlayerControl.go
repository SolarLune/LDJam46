package main

import (
	"math"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/kvartborg/vector"
)

const TypePlayerControlComponent = "PlayerControl"

type PlayerControlComponent struct {
	GameObject *GameObject
	Facing     vector.Vector
}

func NewPlayerControlComponent() *PlayerControlComponent {
	return &PlayerControlComponent{
		Facing: vector.Vector{0, 1},
	}
}

func (pc *PlayerControlComponent) OnAdd(g *GameObject) { pc.GameObject = g }

func (pc *PlayerControlComponent) OnRemove(g *GameObject) {}

func (pc *PlayerControlComponent) Update(screen *ebiten.Image) {

	b := pc.GameObject.GetComponent(TypeBodyComponent)

	if b != nil {

		body := b.(*BodyComponent)

		accel := float64(0.5)
		friction := float64(0.25)
		maxSpeed := float64(2)

		moveDir := vector.Vector{0, 0}

		if ebiten.IsKeyPressed(ebiten.KeyRight) {
			moveDir[0]++
		}
		if ebiten.IsKeyPressed(ebiten.KeyLeft) {
			moveDir[0]--
		}

		if ebiten.IsKeyPressed(ebiten.KeyUp) {
			moveDir[1]--
		}
		if ebiten.IsKeyPressed(ebiten.KeyDown) {
			moveDir[1]++
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyX) {
			if wc := pc.GameObject.GetComponent(TypeWeaponComponent); wc != nil {
				weapon := wc.(*WeaponComponent)
				weapon.FireDirection = pc.Facing.Clone()
				weapon.Fire()
			}
		}

		if body.Speed.Magnitude() < friction {
			body.Speed[0] = 0
			body.Speed[1] = 0
		} else {
			m := body.Speed.Magnitude()
			body.Speed.Unit().Scale(m - friction)
		}

		if moveDir.Magnitude() != 0 {

			body.Speed.Add(moveDir.Unit().Scale(accel))

			if body.Speed.Magnitude() > maxSpeed {
				body.Speed.Unit().Scale(maxSpeed)
			}

			pc.Facing = vector.Unit(body.Speed)

		}

		a := pc.GameObject.GetComponent(TypeAnimationComponent)

		if a != nil {

			anim := a.(*AnimationComponent)

			xc := 0
			yc := 0

			min := 0.3

			if math.Abs(body.Speed[0]) > min {
				xc = 1
			}

			if body.Speed[1] < -min {
				yc = -1
			} else if body.Speed[1] > min {
				yc = 1
			}

			if yc < 0 && xc != 0 {
				anim.Ase.Play("ur")
			} else if yc < 0 && xc == 0 {
				anim.Ase.Play("u")
			} else if yc == 0 && xc != 0 {
				anim.Ase.Play("r")
			} else if yc > 0 && xc != 0 {
				anim.Ase.Play("dr")
			} else if yc > 0 && xc == 0 {
				anim.Ase.Play("d")
			}

			if body.Speed.Magnitude() == 0 {
				anim.Ase.PlaySpeed = 0
				if anim.Ase.CurrentAnimation != nil {
					anim.Ase.CurrentFrame = anim.Ase.CurrentAnimation.Start
				}
			} else {
				anim.Ase.PlaySpeed = 1
			}

			d := pc.GameObject.GetComponent(TypeDrawComponent)

			if d != nil {

				DrawComponent := d.(*DrawComponent)

				if pc.Facing[0] < 0 {
					DrawComponent.FlipHorizontal = true
				} else if pc.Facing[0] > 0 {
					DrawComponent.FlipHorizontal = false
				}

			}

		}

	}

}

func (pc *PlayerControlComponent) Type() string { return TypePlayerControlComponent }

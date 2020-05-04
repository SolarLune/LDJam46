package main

import (
	"image"
	"image/color"
	"math"
	"path/filepath"

	"github.com/SolarLune/goaseprite"
	"github.com/SolarLune/paths"
	"github.com/SolarLune/resolv"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/kvartborg/vector"
)

type Component interface {
	OnAdd(*GameObject)
	OnRemove(*GameObject)
	Update(*ebiten.Image)
	Type() int
}

const (
	TypeDrawComponent = iota
	TypeAnimationComponent
	TypeBodyComponent
	TypePlayerControlComponent
	TypeAIControlComponent
	TypeCameraFollowComponent
	TypeDepthSortComponent
	TypeWeaponComponent
	TypeEndOnCollisionComponent
)

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

func (d *DrawComponent) Type() int { return TypeDrawComponent }

//

//

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

func (a *AnimationComponent) Type() int { return TypeAnimationComponent }

//

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

func (b *BodyComponent) Type() int { return TypeBodyComponent }

func (b *BodyComponent) Center() vector.Vector {
	return vector.Vector{
		b.Object.X + (b.Object.W / 2),
		b.Object.Y + (b.Object.H / 2),
	}
}

//

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

func (pc *PlayerControlComponent) Type() int { return TypePlayerControlComponent }

//

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

func (cf *CameraFollowComponent) Type() int { return TypeCameraFollowComponent }

//

type AIControlComponent struct {
	GameObject             *GameObject
	Facing                 vector.Vector
	TargetPos              vector.Vector
	Path                   *paths.Path
	PathRecalculationTimer int
}

func NewAIControlComponent() *AIControlComponent {
	return &AIControlComponent{
		Facing:    vector.Vector{0, 1},
		TargetPos: vector.Vector{0, 0},
	}
}

func (ai *AIControlComponent) OnAdd(g *GameObject) { ai.GameObject = g }

func (ai *AIControlComponent) OnRemove(g *GameObject) {}

func (ai *AIControlComponent) Update(screen *ebiten.Image) {

	if b := ai.GameObject.GetComponent(TypeBodyComponent); b != nil {

		body := b.(*BodyComponent)
		bodyCenterX, bodyCenterY := body.Object.Center()
		bodyPosition := vector.Vector{bodyCenterX, bodyCenterY}

		if ai.PathRecalculationTimer >= 30 {
			ai.PathRecalculationTimer = 0
			ai.RecalculatePath()
		}
		ai.PathRecalculationTimer++

		if ai.Path != nil {

			targetCell := ai.Path.Next()

			// DEBUG
			if ai.GameObject.Level.Game.DebugMode {

				for _, cell := range ai.Path.Cells {

					cx := float64(cell.X * 16)
					cy := float64(cell.Y * 16)
					cellColor := color.RGBA{0, 255, 0, 192}
					if ai.Path.Current() == cell {
						cellColor = color.RGBA{0, 0, 255, 192}
					}
					ebitenutil.DrawRect(screen, cx-ai.GameObject.Level.CameraOffsetX, cy-ai.GameObject.Level.CameraOffsetY, 16, 16, cellColor)

				}

			}

			if targetCell != nil {

				nextX := float64(targetCell.X*16) + 8
				nextY := float64(targetCell.Y*16) + 8

				dx := nextX - bodyPosition[0]
				dy := nextY - bodyPosition[1]

				dv := vector.Vector{float64(dx), float64(dy)}

				if dv.Magnitude() <= 4 && !ai.Path.AtEnd() {

					potential := ai.Path.Get(ai.Path.CurrentIndex + 3)

					if potential != nil {

						grid := ai.GameObject.Level.PathfindingGrid

						px, py := grid.GridToWorld(potential.X, potential.Y)
						px += float64(grid.CellWidth / 2)
						py += float64(grid.CellHeight / 2)

						diff := vector.Vector{px - bodyPosition[0], py - bodyPosition[1]}

						diff.Scale(0.5)

						tx := bodyPosition[0] + diff[0]
						ty := bodyPosition[1] + diff[1]

						tcx, tcy := body.Object.Space.WorldToSpace(tx, ty)

						if midPoint := body.Object.Space.Cell(tcx, tcy); midPoint != nil && !midPoint.ContainsTags("solid") {
							ai.Path.SetIndex(ai.Path.Index(potential))
						}

					}

				}

				friction := 0.25
				maxSpeed := 2.0
				accel := 0.25 + friction

				if body.Speed[0] > friction {
					body.Speed[0] -= friction
				} else if body.Speed[0] < -friction {
					body.Speed[0] += friction
				} else {
					body.Speed[0] = 0
				}

				if body.Speed[1] > friction {
					body.Speed[1] -= friction
				} else if body.Speed[1] < -friction {
					body.Speed[1] += friction
				} else {
					body.Speed[1] = 0
				}

				body.Speed.Add(dv.Unit().Scale(accel))

				if body.Speed.Magnitude() > maxSpeed {
					body.Speed.Unit().Scale(maxSpeed)
				}

				ai.Facing = body.Speed.Clone().Unit()

			} else {
				ai.RecalculatePath()
			}

		} else {
			ai.RecalculatePath()
		}

		if a := ai.GameObject.GetComponent(TypeAnimationComponent); a != nil {
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

			d := ai.GameObject.GetComponent(TypeDrawComponent)

			if d != nil {

				DrawComponent := d.(*DrawComponent)

				if ai.Facing[0] < 0 {
					DrawComponent.FlipHorizontal = true
				} else if ai.Facing[0] > 0 {
					DrawComponent.FlipHorizontal = false
				}

			}

		}

	}

}

func (ai *AIControlComponent) RecalculatePath() {

	if b := ai.GameObject.GetComponent(TypeBodyComponent); b != nil {

		body := b.(*BodyComponent)
		bodyCenterX, bodyCenterY := body.Object.Center()
		bodyPosition := vector.Vector{bodyCenterX, bodyCenterY}
		grid := ai.GameObject.Level.PathfindingGrid

		target := ai.GameObject.Level.GetGameObjectByComponent(TypePlayerControlComponent)

		if len(target) > 0 {
			targetBody := target[0].GetComponent(TypeBodyComponent).(*BodyComponent)
			ai.TargetPos = targetBody.Center()
		}

		ai.Path = grid.GetPath(bodyPosition[0], bodyPosition[1], ai.TargetPos[0], ai.TargetPos[1], false)

	}

}

func (ai *AIControlComponent) Type() int { return TypeAIControlComponent }

//

type DepthSortComponent struct {
	GameObject *GameObject
	Depth      float64
	YSort      bool
}

func NewDepthSortComponent(autoYSort bool) *DepthSortComponent {
	return &DepthSortComponent{YSort: autoYSort}
}

func (ds *DepthSortComponent) OnAdd(g *GameObject) { ds.GameObject = g }

func (ds *DepthSortComponent) OnRemove(g *GameObject) {}

func (ds *DepthSortComponent) Update(screen *ebiten.Image) {

	if ds.YSort {
		if b := ds.GameObject.GetComponent(TypeBodyComponent); b != nil {
			body := b.(*BodyComponent)
			ds.Depth = body.Center()[1]
		}
	}

}

func (ds *DepthSortComponent) Type() int { return TypeDepthSortComponent }

type WeaponComponent struct {
	GameObject    *GameObject
	Cooldown      int
	Projectile    string
	FireDirection vector.Vector
	SpawnOffset   vector.Vector
}

func NewWeaponComponent() *WeaponComponent {
	return &WeaponComponent{}
}

func (wp *WeaponComponent) OnAdd(g *GameObject) {
	wp.GameObject = g
}

func (wp *WeaponComponent) OnRemove(g *GameObject) {}

func (wp *WeaponComponent) Update(screen *ebiten.Image) {}

func (wp *WeaponComponent) Fire() {

	x, y := 0.0, 0.0

	if bp := wp.GameObject.GetComponent(TypeBodyComponent); bp != nil {
		goBody := bp.(*BodyComponent)
		x, y = goBody.Object.X, goBody.Object.Y
	}

	bullet := NewBullet(wp.GameObject.Level, x, y, wp.FireDirection)
	wp.GameObject.Level.Add(bullet)

}

func (wp *WeaponComponent) Type() int { return TypeWeaponComponent }

//

// type EndOnCollisionComponent struct {
// 	GameObject *GameObject
// }

// func NewEndOnCollisionComponent() *EndOnCollisionComponent {
// 	return &EndOnCollisionComponent{}
// }

// func (eoc *EndOnCollisionComponent) OnAdd(g *GameObject) {
// 	eoc.GameObject = g
// }

// func (eoc *EndOnCollisionComponent) OnRemove(g *GameObject) {}

// func (eoc *EndOnCollisionComponent) Update(screen *ebiten.Image) {

// 	if b := eoc.GameObject.GetComponent(TypeBodyComponent); b != nil {
// 		body := b.(*BodyComponent)
// 		if body.LastCollision != nil {
// 			eoc.GameObject.Level.Remove(eoc.GameObject)
// 		}
// 	}

// }

// func (eoc *EndOnCollisionComponent) Type() int { return TypeEndOnCollisionComponent }

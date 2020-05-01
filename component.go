package main

import (
	"image"
	"math"
	"path/filepath"

	"github.com/SolarLune/goaseprite"
	"github.com/SolarLune/paths"
	"github.com/SolarLune/resolv/resolv"

	"github.com/hajimehoshi/ebiten"
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
)

//

type DrawComponent struct {
	GameObject     *GameObject
	FlipHorizontal bool
}

func NewDrawComponent() *DrawComponent {
	return &DrawComponent{}
}

func (d *DrawComponent) OnAdd(g *GameObject) {
	d.GameObject = g
}

func (d *DrawComponent) OnRemove(g *GameObject) {}

func (d *DrawComponent) Update(screen *ebiten.Image) {

	if d.GameObject.Get(TypeAnimationComponent) != nil {

		geoM := ebiten.GeoM{}
		colorM := ebiten.ColorM{}

		colorM.ChangeHSV(0, 1, 0)

		bodyX := float64(0)
		bodyY := float64(0)

		anim := d.GameObject.Get(TypeAnimationComponent).(*AnimationComponent)
		x, y := anim.Ase.GetFrameXY()

		img := anim.Image.SubImage(image.Rect(int(x), int(y), int(x+anim.Ase.FrameWidth), int(y+anim.Ase.FrameHeight))).(*ebiten.Image)

		srcW, srcH := img.Size()

		if d.GameObject.Get(TypeBodyComponent) != nil {
			body := d.GameObject.Get(TypeBodyComponent).(*BodyComponent)
			bodyX = body.Position[0] - ((float64(srcW) - body.Size[0]) / 2)
			bodyY = body.Position[1] - ((float64(srcH) - body.Size[1]) / 2)
		}

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

type AnimationComponent struct {
	AnimPath string
	Image    *ebiten.Image
	Ase      *goaseprite.File
}

func NewAnimationComponent(animPath string) *AnimationComponent {
	a := &AnimationComponent{AnimPath: animPath}
	a.Ase = goaseprite.Open(a.AnimPath)
	a.Image = GetImage(filepath.Join(filepath.Dir(a.AnimPath), a.Ase.ImagePath))
	return a
}

func (a *AnimationComponent) OnAdd(g *GameObject) {}

func (a *AnimationComponent) OnRemove(g *GameObject) {}

func (a *AnimationComponent) Update(screen *ebiten.Image) { a.Ase.Update(1.0 / 60.0) }

func (a *AnimationComponent) Type() int { return TypeAnimationComponent }

//

type BodyComponent struct {
	GameObject *GameObject
	Position   vector.Vector
	Size       vector.Vector
	Speed      vector.Vector
	Shape      *resolv.Rectangle
}

func NewBodyComponent(x, y, w, h float64) *BodyComponent {

	return &BodyComponent{
		Position: vector.Vector{x, y},
		Size:     vector.Vector{w, h},
		Speed:    vector.Vector{0, 0},
		Shape:    resolv.NewRectangle(0, 0, 0, 0),
	}

}

func (b *BodyComponent) OnAdd(g *GameObject) {
	b.GameObject = g
	g.Level.Space.Add(b.Shape)
}

func (b *BodyComponent) OnRemove(g *GameObject) { g.Level.Space.Remove(b.Shape) }

func (b *BodyComponent) Update(screen *ebiten.Image) {

	solids := b.GameObject.Level.Space.FilterByTags("solid")

	roundUp := func(value float64) int32 {
		if value < 0 {
			return int32(math.Floor(value))
		} else {
			return int32(math.Ceil(value))
		}
	}

	b.Shape.X = int32(math.Round(b.Position[0]))
	b.Shape.Y = int32(math.Round(b.Position[1]))
	b.Shape.W = int32(math.Round(b.Size[0]))
	b.Shape.H = int32(math.Round(b.Size[1]))

	if res := solids.Resolve(b.Shape, roundUp(b.Speed[0]), 0); res.Colliding() {
		b.Position[0] += float64(res.ResolveX)
		b.Speed[0] = 0
	} else {
		b.Position[0] += b.Speed[0]
	}

	if res := solids.Resolve(b.Shape, 0, roundUp(b.Speed[1])); res.Colliding() {
		b.Position[1] += float64(res.ResolveY)
		b.Speed[1] = 0
	} else {
		b.Position[1] += b.Speed[1]
	}

}

func (b *BodyComponent) Type() int { return TypeBodyComponent }

func (b *BodyComponent) Center() vector.Vector {
	return vector.Vector{
		b.Position[0] + (b.Size[0] / 2),
		b.Position[1] + (b.Size[1] / 2),
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

	b := pc.GameObject.Get(TypeBodyComponent)

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

		a := pc.GameObject.Get(TypeAnimationComponent)

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

			d := pc.GameObject.Get(TypeDrawComponent)

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

func (pc *PlayerControlComponent) Type() int {
	return TypePlayerControlComponent
}

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

	b := cf.GameObject.Get(TypeBodyComponent)

	if b != nil {

		body := b.(*BodyComponent)

		tx := body.Position[0] - cf.OffsetX - cf.GameObject.Level.CameraOffsetX
		cf.GameObject.Level.CameraOffsetX += tx * cf.Softness

		ty := body.Position[1] - cf.OffsetY - cf.GameObject.Level.CameraOffsetY
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

	// cf.Level.CameraOffsetX =

}
func (cf *CameraFollowComponent) Type() int { return TypeCameraFollowComponent }

//

type AIControlComponent struct {
	GameObject             *GameObject
	Facing                 vector.Vector
	TargetPos              vector.Vector
	Path                   *paths.Path
	RayTest                *resolv.Line
	PathRecalculationTimer int
}

func NewAIControlComponent() *AIControlComponent {
	return &AIControlComponent{
		Facing:    vector.Vector{0, 1},
		TargetPos: vector.Vector{0, 0},
		RayTest:   resolv.NewLine(0, 0, 0, 0),
	}
}

func (ai *AIControlComponent) OnAdd(g *GameObject) { ai.GameObject = g }

func (ai *AIControlComponent) OnRemove(g *GameObject) {}

func (ai *AIControlComponent) Update(screen *ebiten.Image) {

	b := ai.GameObject.Get(TypeBodyComponent)

	grid := ai.GameObject.Level.PathfindingGrid

	if b != nil {

		body := b.(*BodyComponent)
		bodyPosition := body.Center()

		if ai.PathRecalculationTimer >= 30 {
			ai.PathRecalculationTimer = 0
			ai.Path = nil
		}
		ai.PathRecalculationTimer++

		if ai.Path != nil {

			targetCell := ai.Path.Next()
			if targetCell == nil {
				targetCell = ai.Path.Current()
			}

			// solids := ai.GameObject.Level.Space.FilterByTags("solid")

			// cancelPotential := false

			// for i := ai.Path.CurrentIndex; i < ai.Path.Length(); i++ {

			// 	potential := ai.Path.Get(i)

			// 	px, py := grid.GridToWorld(potential.X, potential.Y)
			// 	px += float64(grid.CellWidth / 2)
			// 	py += float64(grid.CellWidth / 2)

			// 	direction := vector.Vector{px - bodyPosition[0], py - bodyPosition[1]}.Unit()

			// 	rotations := []float64{
			// 		math.Pi / 2,
			// 		-math.Pi / 2,
			// 	}

			// 	freePotential := true

			// 	for _, rotation := range rotations {

			// 		bp := vector.Vector{bodyPosition[0], bodyPosition[1]}
			// 		offset := vector.Rotate(direction, rotation).Scale(float64(grid.CellWidth / 3))
			// 		ep := vector.Vector{px, py}
			// 		bp.Add(offset)
			// 		ep.Add(offset)

			// 		ai.RayTest.X = int32(bp[0])
			// 		ai.RayTest.Y = int32(bp[1])
			// 		ai.RayTest.X2 = int32(ep[0])
			// 		ai.RayTest.Y2 = int32(ep[1])

			// 		// DEBUG
			// 		if ai.GameObject.Level.Game.DebugMode {
			// 			for _, cell := range ai.Path.Cells {
			// 				cx := float64(cell.X * 16)
			// 				cy := float64(cell.Y * 16)
			// 				ebitenutil.DrawRect(screen, cx-ai.GameObject.Level.CameraOffsetX, cy-ai.GameObject.Level.CameraOffsetY, 16, 16, color.RGBA{0, 255, 0, 192})
			// 			}
			// 			ebitenutil.DrawLine(
			// 				screen,
			// 				float64(ai.RayTest.X)-ai.GameObject.Level.CameraOffsetX,
			// 				float64(ai.RayTest.Y)-ai.GameObject.Level.CameraOffsetY,
			// 				float64(ai.RayTest.X2)-ai.GameObject.Level.CameraOffsetX,
			// 				float64(ai.RayTest.Y2)-ai.GameObject.Level.CameraOffsetY,
			// 				color.White)
			// 		}

			// 		if solids.IsColliding(ai.RayTest) {
			// 			cancelPotential = true
			// 			freePotential = false
			// 			break
			// 		}

			// 	}

			// 	if freePotential {
			// 		targetCell = potential
			// 	}

			// 	if cancelPotential {
			// 		break
			// 	}

			// }

			if targetCell != nil {

				nextX := float64(targetCell.X*16) + 8
				nextY := float64(targetCell.Y*16) + 8

				dx := nextX - bodyPosition[0]
				dy := nextY - bodyPosition[1]

				dv := vector.Vector{float64(dx), float64(dy)}
				if dv.Magnitude() <= 4 {
					if ai.Path.AtEnd() {
						ai.Path = nil
					} else {
						ai.Path.SetIndex(ai.Path.Index(targetCell) + 1)
						// ai.Path.Advance()
					}
				}

				friction := 0.25
				maxSpeed := 2.0
				accel := 0.25 + friction

				if body.Speed.Magnitude() > friction {
					m := body.Speed.Magnitude()
					body.Speed.Unit().Scale(m - friction)
				} else {
					body.Speed = vector.Vector{0, 0}
				}

				body.Speed.Add(dv.Unit().Scale(accel))

				if body.Speed.Magnitude() > maxSpeed {
					body.Speed.Unit().Scale(maxSpeed)
				}

			}

		} else {

			// GENERATE PATH

			target := ai.GameObject.Level.GetGameObjectByComponent(TypePlayerControlComponent)

			if len(target) > 0 {
				targetBody := target[0].Get(TypeBodyComponent).(*BodyComponent)
				ai.TargetPos = targetBody.Center()
			}

			ai.Path = grid.GetPath(bodyPosition[0], bodyPosition[1], ai.TargetPos[0], ai.TargetPos[1], false)

		}

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
		if b := ds.GameObject.Get(TypeBodyComponent); b != nil {
			body := b.(*BodyComponent)
			ds.Depth = body.Center()[1]
		}
	}

}

func (ds *DepthSortComponent) Type() int { return TypeDepthSortComponent }

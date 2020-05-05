package main

import (
	"image/color"
	"math"

	"github.com/SolarLune/paths"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/kvartborg/vector"
)

const TypeAIControlComponent = "AIControl"

type AIControlComponent struct {
	GameObject             *GameObject
	Facing                 vector.Vector
	TargetPos              vector.Vector
	Path                   *paths.Path
	PathRecalculationTimer int
	TargetSpeed            vector.Vector
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

		if ai.PathRecalculationTimer >= 3000 {
			ai.PathRecalculationTimer = 0
			ai.RecalculatePath()
		}
		ai.PathRecalculationTimer++

		if ai.Path != nil {

			targetCell := ai.Path.Current()

			// DEBUG
			if ai.GameObject.Level.Game.DebugMode {

				for _, cell := range ai.Path.Cells {

					cx := float64(cell.X * 16)
					cy := float64(cell.Y * 16)
					cellColor := color.RGBA{0, 255, 0, 192}
					if ai.Path.Next() == cell {
						cellColor = color.RGBA{0, 0, 255, 192}
					}
					ebitenutil.DrawRect(screen, cx-ai.GameObject.Level.CameraOffsetX, cy-ai.GameObject.Level.CameraOffsetY, 16, 16, cellColor)

				}

			}

			if targetCell != nil {

				space := ai.GameObject.Level.Space

				nextX := float64((targetCell.X * space.CellWidth) + (space.CellWidth / 2))
				nextY := float64((targetCell.Y * space.CellHeight) + (space.CellHeight / 2))

				dx := nextX - bodyPosition[0]
				dy := nextY - bodyPosition[1]

				dv := vector.Vector{float64(dx), float64(dy)}

				if dv.Magnitude() <= 4 {

					if !ai.Path.AtEnd() {

						ai.Path.Advance()

						for i := len(ai.Path.Cells) - 1; i > 0; i-- {
							end := ai.Path.Cells[i]
							start := ai.Path.Cells[ai.Path.CurrentIndex]

							if start == end {
								break
							}

							free := true

							for _, c := range space.CellsInLine(start.X, start.Y, end.X, end.Y) {
								if c.ContainsTags("solid") {
									free = false
									break
								}
							}

							if free {
								ai.Path.SetIndex(i - 1)
								body.Speed = vector.Vector{0, 0}

								targetCell = ai.Path.Next()
								nextX = float64((targetCell.X * space.CellWidth) + (space.CellWidth / 2))
								nextY = float64((targetCell.Y * space.CellHeight) + (space.CellHeight / 2))

								dx = nextX - bodyPosition[0]
								dy = nextY - bodyPosition[1]

								dv = vector.Vector{float64(dx), float64(dy)}

								break
							}

						}

					} else {
						ai.RecalculatePath()
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

func (ai *AIControlComponent) Type() string { return TypeAIControlComponent }

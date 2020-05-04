package main

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"sort"

	"github.com/hajimehoshi/ebiten/ebitenutil"

	"github.com/SolarLune/dngn"
	"github.com/SolarLune/paths"
	"github.com/SolarLune/resolv"
	"github.com/hajimehoshi/ebiten"
)

// MAP CELL TYPES
const FLOOR = ' '
const WALL = 'x'

type Level struct {
	Game                         *Game
	Map                          *dngn.Room
	PathfindingGrid              *paths.Grid
	GameObjects                  []*GameObject
	ToRemove                     []*GameObject
	MapImageBG                   *ebiten.Image
	MapImageFG                   *ebiten.Image
	Space                        *resolv.Space
	CameraOffsetX, CameraOffsetY float64
}

func NewLevel(game *Game) *Level {

	cellW := 16
	cellH := 16

	level := &Level{
		Game:        game,
		Map:         dngn.NewRoom(60, 60),
		GameObjects: []*GameObject{},
		Space:       resolv.NewSpace(60, 60, cellW, cellH),
	}

	level.MapImageBG, _ = ebiten.NewImage(level.Map.Width*16, level.Map.Height*16, ebiten.FilterNearest)
	level.MapImageFG, _ = ebiten.NewImage(level.Map.Width*16, level.Map.Height*16, ebiten.FilterNearest)

	level.Init()

	return level

}

func (level *Level) Init() {

	level.Map.Select().Fill(WALL)

	// level.Map.GenerateRandomRooms(FLOOR, 8, 4, 4, 8, 8, true)
	level.Map.GenerateDrunkWalk(FLOOR, 0.5)

	// Border
	level.Map.Select().Shrink(true).Invert().Fill(WALL)

	// Spawn objects

	player := NewPlayer(level)
	level.Add(player)

	npc := NewNPC(level)
	level.Add(npc)

	pb := player.GetComponent(TypeBodyComponent).(*BodyComponent)
	npcBody := npc.GetComponent(TypeBodyComponent).(*BodyComponent)
	npcBody.Object.X = pb.Object.X
	npcBody.Object.Y = pb.Object.Y

	for _, ci := range level.Map.Select().ByRune(WALL).Cells {
		cx := float64(ci[0] * 16)
		cy := float64(ci[1] * 16)
		obj := resolv.NewObject(cx, cy, 16, 16, level.Space)
		obj.AddTag("solid")
	}

	level.RenderTiles()

	level.PathfindingGrid = paths.NewGridFromRuneArrays(level.Map.Data, 16, 16)
	level.PathfindingGrid.SetWalkable(WALL, false)

}

func (level *Level) Update(screen *ebiten.Image) {

	screen.Fill(color.RGBA{20, 18, 29, 255})

	geoM := ebiten.GeoM{}
	geoM.Translate(-level.CameraOffsetX, -level.CameraOffsetY-8)
	screen.DrawImage(level.MapImageBG, &ebiten.DrawImageOptions{GeoM: geoM})

	// Sort game objects by depth if they've got the component
	sort.Slice(level.GameObjects, func(i, j int) bool {

		if da := level.GameObjects[i].GetComponent(TypeDepthSortComponent); da != nil {

			if db := level.GameObjects[j].GetComponent(TypeDepthSortComponent); db != nil {

				depthA := da.(*DepthSortComponent).Depth
				depthB := db.(*DepthSortComponent).Depth
				return depthA < depthB

			}

		}

		return false
	})

	for _, g := range level.GameObjects {
		g.Update(screen)
	}

	screen.DrawImage(level.MapImageFG, &ebiten.DrawImageOptions{GeoM: geoM})

	for _, gameObject := range level.ToRemove {

		for i, g := range level.GameObjects {

			g.OnRemove()

			if g == gameObject {
				level.GameObjects = append(level.GameObjects[:i], level.GameObjects[i+1:]...)
			}

		}

	}

	level.ToRemove = []*GameObject{}

	if level.Game.DebugMode {

		for y := 0; y < level.Space.Height(); y++ {

			for x := 0; x < level.Space.Width(); x++ {

				cell := level.Space.Cell(x, y)

				cw := float64(level.Space.CellWidth)
				ch := float64(level.Space.CellHeight)
				cx := float64(cell.X) * cw
				cy := float64(cell.Y) * ch

				drawColor := color.RGBA{20, 20, 20, 255}

				if cell.Occupied() {
					drawColor = color.RGBA{255, 255, 0, 255}
				}

				camX := level.CameraOffsetX
				camY := level.CameraOffsetY

				ebitenutil.DrawLine(screen, cx-camX, cy-camY, cx-camX+cw, cy-camY, drawColor)

				ebitenutil.DrawLine(screen, cx+cw-camX, cy-camY, cx-camX+cw, cy+ch-camY, drawColor)

				ebitenutil.DrawLine(screen, cx+cw-camX, cy+ch-camY, cx-camX, cy+ch-camY, drawColor)

				ebitenutil.DrawLine(screen, cx-camX, cy+ch-camY, cx-camX, cy-camY, drawColor)
			}

		}

	}

}

func (level *Level) RenderTiles() {

	level.MapImageBG.Fill(color.Transparent)

	tileset := GetImage("assets/tileset.png")

	for y := 0; y < level.Map.Height; y++ {

		for x := 0; x < level.Map.Width; x++ {

			srcX := 0
			srcY := 0
			rotation := float64(0)

			value := level.Map.Get(x, y)

			switch value {
			case FLOOR:
				srcX = 0
				srcY = 16
				if rand.Float32() < 0.1 {
					srcY = 32
				}
				if level.Map.Get(x, y-1) == WALL {
					srcY = 0
				}
			case WALL:

				left := level.Map.Get(x-1, y) == value
				right := level.Map.Get(x+1, y) == value
				up := level.Map.Get(x, y-1) == value
				down := level.Map.Get(x, y+1) == value

				if x == 0 {
					left = true
				} else if x == level.Map.Width-1 {
					right = true
				}

				if y == 0 {
					up = true
				} else if y == level.Map.Height-1 {
					down = true
				}

				num := 0
				if left {
					num++
				}
				if right {
					num++
				}
				if up {
					num++
				}
				if down {
					num++
				}

				if num == 4 {
					srcX = 32
					srcY = 16

				} else if num == 3 {
					srcX = 32
					srcY = 0

					if !right {
						rotation = math.Pi / 2
					} else if !down {
						rotation = math.Pi
					} else if !left {
						rotation = -math.Pi / 2
					}

				} else if num == 2 {

					if right && left {
						srcX = 48
						srcY = 16
						rotation = math.Pi / 2
					} else if up && down {
						srcX = 48
						srcY = 16
					} else if down && left {
						// Corners
						srcX = 48
						srcY = 0
					} else if up && left {
						// Corners
						srcX = 48
						srcY = 0
						rotation = math.Pi / 2
					} else if up && right {
						// Corners
						srcX = 48
						srcY = 0
						rotation = math.Pi
					} else if right && down {
						// Corners
						srcX = 48
						srcY = 0
						rotation = -math.Pi / 2
					}

				} else if num == 1 {

					srcX = 16
					srcY = 0

					if down {
						rotation = math.Pi / 2
					} else if left {
						rotation = math.Pi
					} else if up {
						rotation = -math.Pi / 2
					}

				} else {
					srcX = 16
					srcY = 16
				}
			}

			sub := tileset.SubImage(image.Rect(srcX, srcY, srcX+16, srcY+16)).(*ebiten.Image)

			geoM := ebiten.GeoM{}

			geoM.Translate(-8, -8)
			geoM.Rotate(rotation)
			geoM.Translate(8, 8)

			geoM.Translate(float64(x*16), float64(y*16))

			if value == FLOOR {
				level.MapImageBG.DrawImage(sub, &ebiten.DrawImageOptions{GeoM: geoM})
			} else {
				level.MapImageFG.DrawImage(sub, &ebiten.DrawImageOptions{GeoM: geoM})
			}

		}

	}

}

func (level *Level) Add(g *GameObject) {
	level.GameObjects = append(level.GameObjects, g)
}

func (level *Level) Remove(g *GameObject) {
	level.ToRemove = append(level.ToRemove, g)
}

func (level *Level) GetGameObjectByComponent(componentTypeConstant int) []*GameObject {

	goList := []*GameObject{}

	for _, g := range level.GameObjects {
		if g.GetComponent(componentTypeConstant) != nil {
			goList = append(goList, g)
		}
	}

	return goList

}

func (level *Level) Width() int {
	return level.Map.Width * 16
}

func (level *Level) Height() int {
	return level.Map.Height * 16
}

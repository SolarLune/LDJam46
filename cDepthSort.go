package main

import "github.com/hajimehoshi/ebiten"

const TypeDepthSortComponent = "DepthSort"

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

func (ds *DepthSortComponent) Type() string { return TypeDepthSortComponent }

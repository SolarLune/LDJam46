package main

import "github.com/hajimehoshi/ebiten"

type GameObject struct {
	Level      *Level
	Components []Component
	ToRemove   []Component
}

func NewGameObject(level *Level) *GameObject {
	g := &GameObject{
		Level:      level,
		Components: []Component{},
		ToRemove:   []Component{},
	}
	return g
}

func (g *GameObject) Update(screen *ebiten.Image) {

	for _, c := range g.Components {
		c.Update(screen)
	}

	for _, component := range g.ToRemove {

		for i, c := range g.Components {

			if c == component {
				g.Components = append(g.Components[:i], g.Components[i+1:]...)
			}

		}

	}

	g.ToRemove = []Component{}

}

func (g *GameObject) Add(components ...Component) {
	for _, component := range components {
		component.OnAdd(g)
		g.Components = append(g.Components, component)
	}
}

func (g *GameObject) Remove(components ...Component) {

	for _, component := range components {

		for i, c := range g.Components {
			if c == component {
				component.OnRemove(g)
				g.Components = append(g.Components[:i], g.Components[i+1:]...)
				break
			}
		}

	}

}

func (g *GameObject) Get(componentTypeConstant int) Component {

	for _, c := range g.Components {
		if c.Type() == componentTypeConstant {
			return c
		}
	}

	return nil

}

func (g *GameObject) Clear() {
	for _, c := range g.Components {
		c.OnRemove(g)
	}
	g.Components = []Component{}
}
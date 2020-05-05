package main

import (
	"github.com/hajimehoshi/ebiten"
	"github.com/kvartborg/vector"
)

const TypeWeaponComponent = "Weapon"

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

func (wp *WeaponComponent) Type() string { return TypeWeaponComponent }

package main

import (
	"math"

	"github.com/kvartborg/vector"
)

func NewPlayer(level *Level) *GameObject {

	player := NewGameObject(level)

	spawnPoint := level.Map.Select().ByRune(FLOOR).ByPercentage(0.1).Cells[0]

	player.AddComponent(
		NewBodyComponent(float64(spawnPoint[0]*16), float64(spawnPoint[1]*16), 8, 8, level.Space),
		NewDrawComponent(0, 0),
		NewDepthSortComponent(true),
		NewAnimationComponent("assets/npc.json"),
		NewPlayerControlComponent(),
		NewCameraFollowComponent(),
		NewWeaponComponent(),
	)

	return player

}

func NewNPC(level *Level) *GameObject {

	npc := NewGameObject(level)

	npc.AddComponent(
		NewBodyComponent(0, 0, 8, 8, level.Space),
		NewDrawComponent(0, 0),
		NewDepthSortComponent(true),
		NewAnimationComponent("assets/npc.json"),
		NewAIControlComponent(),
	)

	return npc

}

func NewBullet(level *Level, x, y float64, movementDirection vector.Vector) *GameObject {

	bullet := NewGameObject(level)
	body := NewBodyComponent(x, y, 4, 4, level.Space)
	body.Speed = movementDirection.Scale(4)

	body.OnBump = func(b *BodyComponent) { // Remove bullet and spawn explosion
		level := b.GameObject.Level
		level.Remove(b.GameObject)
		level.Add(NewExplosionParticle(level, b.Object.X, b.Object.Y))
	}

	anim := NewAnimationComponent("assets/shot.json")
	anim.Ase.Play("Anim")
	draw := NewDrawComponent(0, 0)
	draw.Rotation, _ = vector.Vector{1, 0}.Angle(movementDirection)
	draw.Rotation += math.Pi / 2

	bullet.AddComponent(
		anim,
		draw,
		body,
	)

	return bullet

}

func NewExplosionParticle(level *Level, x, y float64) *GameObject {

	particle := NewGameObject(level)

	draw := NewDrawComponent(x-8, y-8)

	ds := NewDepthSortComponent(false)
	ds.Depth = -1000000

	anim := NewAnimationComponent("assets/small_explosion.json")
	anim.Ase.Play("Anim")
	anim.OnAnimEnd = func(anim *AnimationComponent) {
		draw.Visible = false
		level.Remove(particle)
	}

	particle.AddComponent(
		ds,
		anim,
		draw,
	)

	return particle

}

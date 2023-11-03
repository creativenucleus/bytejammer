package server

import (
	"math/rand"
)

var machineNameRand = rand.Int()

func GetFunName(index int) string {
	var names = [...]string{
		"Sunseed Berry",
		"Face Wrinkler",
		"Cupid's Grenade",
		"Dapper Blob",
		"Velvety Dreamdrop",
		"Citrus Lump",
		"Dusk Pustules",
		"Heroine's Tear",
		"Pocked Airhead",
		"Crimson Banquet",
		"Delectable Bouquet",
		"Lesser Mock Bottom",
		"Portable Sunset",
		"Blonde Imposter",
		"Insect Condo",
		"Seed Hive",
		"Searing Acidshock",
		"Firebreathing Feast",
		"Dragon Fruit",
		"Zest Bomb",
		"Golden Sunseed",
		"Stellar Extrusion",
		"Scaly Custard",
		"Wayward Moon",
		"Astringent Clump",
		"Mock Bottom",
		"Slapstick Crescent",
		"Golden Grenade",
		"Dawn Pustules",
		"Disguised Delicacy",
		"Juicy Gaggle",
		"Crunchy Deluge",
		"Tremendous Sniffer",
	}

	nameIndex := (machineNameRand + index) % len(names)
	return names[nameIndex]
}

package server

func GetFunName(id int) string {
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

	return names[id%len(names)]
}

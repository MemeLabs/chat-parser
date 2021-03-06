// +build gofuzz

package parser

// Fuzz is used by go-fuzz
func Fuzz(data []byte) int {
	ctx := NewParserContext(ParserContextValues{
		Emotes:         emotes,
		EmoteModifiers: modifiers,
		Nicks:          names,
		Tags:           []string{"nsfw", "weeb", "nsfl", "spoiler"},
	})
	p := NewParser(ctx, NewLexer(string(data)))

	ast := p.ParseMessage()
	if ast.Nodes != nil {
		// increase weight of "interesting" content
		return 1
	}
	return 0
}

var modifiers = []string{"wide",
	"rustle",
	"spin",
	"banned",
	"flip",
	"hyper",
	"lag",
	"love",
	"mirror",
	"rain",
	"snow",
	"worth",
	"virus",
}

var emotes = []string{"INFESTOR",
	"FIDGETLOL",
	"Hhhehhehe",
	"GameOfThrows",
	"Abathur",
	"LUL",
	"SURPRISE",
	"NoTears",
	"OverRustle",
	"DuckerZ",
	"Kappa",
	"Klappa",
	"DappaKappa",
	"BibleThump",
	"AngelThump",
	"BasedGod",
	"OhKrappa",
	"SoDoge",
	"WhoahDude",
	"MotherFuckinGame",
	"DaFeels",
	"UWOTM8",
	"DatGeoff",
	"FerretLOL",
	"Sippy",
	"Nappa",
	"DAFUK",
	"HEADSHOT",
	"DANKMEMES",
	"MLADY",
	"MASTERB8",
	"NOTMYTEMPO",
	"LeRuse",
	"YEE",
	"SWEATY",
	"PEPE",
	"SpookerZ",
	"WEEWOO",
	"ASLAN",
	"TRUMPED",
	"BASEDWATM8",
	"BERN",
	"Hmmm",
	"PepoThink",
	"FeelsAmazingMan",
	"FeelsBadMan",
	"FeelsGoodMan",
	"OhMyDog",
	"Wowee",
	"haHAA",
	"POTATO",
	"NOBULLY",
	"gachiGASM",
	"REE",
	"monkaS",
	"RaveDoge",
	"CuckCrab",
	"MiyanoHype",
	"ECH",
	"NiceMeMe",
	"ITSRAWWW",
	"Riperino",
	"4Head",
	"BabyRage",
	"EleGiggle",
	"Kreygasm",
	"PogChamp",
	"SMOrc",
	"NotLikeThis",
	"POGGERS",
	"AYAYA",
	"PepOk",
	"PepoComfy",
	"PepoWant",
	"PepeHands",
	"BOGGED",
	"ComfyApe",
	"ApeHands",
	"OMEGALUL",
	"COGGERS",
	"PepoWant",
	"Clap",
	"FeelsWeirdMan",
	"monkaMEGA",
	"ComfyDog",
	"GIMI",
	"MOOBERS",
	"PepoBan",
	"ComfyAYA",
	"ComfyFerret",
	"BOOMER",
	"ZOOMER",
	"SOY",
	"FeelsPepoMan",
	"ComfyCat",
	"ComfyPOTATO",
	"SUGOI",
	"DJPepo",
	"CampFire",
	"ComfyYEE",
	"weSmart",
	"PepoG",
	"OBJECTION",
	"ComfyWeird",
	"umaruCry",
	"OsKrappa",
	"monkaHmm",
	"PepoHmm",
	"PepeComfy",
	"SUGOwO",
	"EZ",
	"Pepega",
	"shyLurk",
	"FeelsOkayMan",
	"POKE",
	"PepoDance",
	"ORDAH",
	"SPY",
	"PepoGood",
	"PepeJam",
	"LAG",
	"billyWeird",
	"SOTRIGGERED",
	"OnlyPretending",
	"cmonBruh",
	"VroomVroom",
	"mikuDance",
	"WAG",
	"PepoFight",
	"NeneLaugh",
	"PepeLaugh",
	"PeepoS",
	"SLEEPY",
	"GODMAN",
	"NOM",
	"FeelsDumbMan",
	"SEMPAI",
	"OSTRIGGERED",
	"MiyanoBird",
	"KING",
	"PIKOHH",
	"PepoPirate",
	"PepeMods",
	"OhISee",
	"WeirdChamp",
	"RedCard",
	"illyaTriggered",
	"SadBenis",
	"PeepoHappy",
	"ComfyWAG",
	"MiyanoComfy",
	"sataniaLUL",
	"DELUSIONAL",
	"GREED",
	"AYAWeird",
	"FeelsCountryMan",
	"SNAP",
	"PeepoRiot",
	"HiHi",
	"ComfyFeels",
	"MiyanoSip",
	"PeepoWeird",
	"JimFace",
	"HACKER",
	"monkaVirus",
	"DOUBT",
	"KEKW",
	"SHOCK",
}

var names = []string{"Readysetzerg",
	"MorDDDka",
	"Sine",
	"Alex4",
	"ProstateFondler",
	"BlackBolt",
	"Voiture",
	"TsoChicken",
	"SoMuchForSubtlety",
	"vergobret",
	"Char",
	"Stanlot",
	"Tensei",
	"ceele",
	"Mannekino",
	"CosmicPanda",
	"chihayalove72",
	"UltimaTheSeventh",
	"ehu",
	"TheFamousJohnyD",
	"abeous",
	"emmelnem",
	"Haxxy",
	"IchiFi",
	"Cinder",
	"Vaan99",
	"goodguy",
	"Feopachi",
	"bartrand96",
	"zapnuk",
	"mossad",
	"moogy",
	"thiswilldestroyu",
	"keyno",
	"Raskolnikow",
	"anon",
	"techtilla",
	"Zenxbear",
	"dysonia",
	"yardhin",
	"Astraea",
	"Nept",
	"mcdoudles",
	"dna666",
	"jansoon",
	"Silverpikachu1",
	"wishingskeleton",
	"Deadwing",
	"Jope",
	"mentions",
	"jirio",
	"7mango_",
	"n0se13",
	"flyingsausage",
	"gzo852",
	"Gehirnchirurg",
	"palkess",
	"KartoffelKopf",
	"not_johnny420",
	"stin1",
	"arkzats",
	"laetus",
	"Of_Odin",
	"Nako",
	"suvaacc",
	"gball",
	"microburger",
	"Ftwpala",
	"dekezander",
	"blankspaceblank",
	"Cracksi",
	"fuyocouch",
	"Slugalisk",
	"ritten",
	"Gramol",
	"hightalian",
	"jolantru",
	"frustrated_nerd",
	"Bot",
	"jbpratt",
	"Xymos",
	"decoiii",
	"RightToBearArmsLOL",
	"wuforever",
	"darkasmysoul",
	"w1nter",
	"jamcrackers",
	"bngw",
	"cosm",
	"Versicarius",
	"box",
	"Juzmo",
	"syunfox",
	"Valerion",
	"ArmandoEnjoysMen",
	"Rippig",
	"butterflytechnique",
	"biscophan",
	"bingobongo33",
	"j2y",
	"Josh",
	"ronsbro",
	"Nox",
	"SlothIsASin",
	"Noctale",
	"D0UD",
	"Greythorn",
	"Sufflelol",
	"Stefono",
	"sirrenitee",
	"Kirby",
	"toh_",
	"41b",
	"whenis",
	"mokuhazushi",
	"Zimveo",
	"plentifulSwag",
	"Feenamabob",
	"Daenda",
	"xartemisx",
	"Guava",
	"Robby12320",
	"exempteel",
	"dragopie",
	"cheatmasterxii",
	"badsync",
}

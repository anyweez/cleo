package libcleo

import (
	"log"
	"proto"
	"strings"
)

// Some basic data structures.

// LivePCGL and LivePCGLRecord are both used at query runtime.
type LivePCGL struct {
	Champions map[proto.ChampionType]LivePCGLRecord
	All       []GameId
}

type LivePCGLRecord struct {
	Winning []GameId
	Losing  []GameId
}

// RecordContainer describes what a row in the MongoDB is composed of.
type RecordContainer struct {
	GameData  []byte
	GameId    uint64		// 64-bit because it's used pre-packing.
	Timestamp uint64
}

type GameId uint32

// Based on version 4.7.16 from Riot API.
var champion_map = map[uint32]proto.ChampionType{
	1:   proto.ChampionType_ANNIE,
	2:   proto.ChampionType_OLAF,
	3:   proto.ChampionType_GALIO,
	4:   proto.ChampionType_TWISTED_FATE,
	5:   proto.ChampionType_XIN_ZHAO,
	6:   proto.ChampionType_URGOT,
	7:   proto.ChampionType_LEBLANC,
	8:   proto.ChampionType_VLADMIR,
	9:   proto.ChampionType_FIDDLESTICKS,
	10:  proto.ChampionType_KAYLE,
	11:  proto.ChampionType_MASTER_YI,
	12:  proto.ChampionType_ALISTAR,
	13:  proto.ChampionType_RYZE,
	14:  proto.ChampionType_SION,
	15:  proto.ChampionType_SIVIR,
	16:  proto.ChampionType_SORAKA,
	17:  proto.ChampionType_TEEMO,
	18:  proto.ChampionType_TRISTANA,
	19:  proto.ChampionType_WARWICK,
	20:  proto.ChampionType_NUNU,
	21:  proto.ChampionType_MISS_FORTUNE,
	22:  proto.ChampionType_ASHE,
	23:  proto.ChampionType_TRYNDAMERE,
	24:  proto.ChampionType_JAX,
	25:  proto.ChampionType_MORGANA,
	26:  proto.ChampionType_ZILEAN,
	27:  proto.ChampionType_SINGED,
	28:  proto.ChampionType_EVELYN,
	29:  proto.ChampionType_TWITCH,
	30:  proto.ChampionType_KARTHUS,
	31:  proto.ChampionType_CHOGATH,
	32:  proto.ChampionType_AMUMU,
	33:  proto.ChampionType_RAMMUS,
	34:  proto.ChampionType_ANIVIA,
	35:  proto.ChampionType_SHACO,
	36:  proto.ChampionType_DR_MUNDO,
	37:  proto.ChampionType_SONA,
	38:  proto.ChampionType_KASSADIN,
	39:  proto.ChampionType_IRELIA,
	40:  proto.ChampionType_JANNA,
	41:  proto.ChampionType_GANGPLANK,
	42:  proto.ChampionType_CORKI,
	43:  proto.ChampionType_KARMA,
	44:  proto.ChampionType_TARIC,
	45:  proto.ChampionType_VEIGAR,
	48:  proto.ChampionType_TRUNDLE,
	50:  proto.ChampionType_SWAIN,
	51:  proto.ChampionType_CAITLYN,
	53:  proto.ChampionType_BLITZCRANK,
	54:  proto.ChampionType_MALPHITE,
	55:  proto.ChampionType_KATARINA,
	56:  proto.ChampionType_NOCTURNE,
	57:  proto.ChampionType_MAOKAI,
	58:  proto.ChampionType_RENEKTON,
	59:  proto.ChampionType_JARVAN_IV,
	60:  proto.ChampionType_ELISE,
	61:  proto.ChampionType_ORIANNA,
	62:  proto.ChampionType_WUKONG,
	63:  proto.ChampionType_BRAND,
	64:  proto.ChampionType_LEE_SIN,
	67:  proto.ChampionType_VAYNE,
	68:  proto.ChampionType_RUMBLE,
	69:  proto.ChampionType_CASSIOPEIA,
	72:  proto.ChampionType_SKARNER,
	74:  proto.ChampionType_HEIMERDINGER,
	75:  proto.ChampionType_NASUS,
	76:  proto.ChampionType_NIDALEE,
	77:  proto.ChampionType_UDYR,
	78:  proto.ChampionType_POPPY,
	79:  proto.ChampionType_GRAGAS,
	80:  proto.ChampionType_PANTHEON,
	81:  proto.ChampionType_EZREAL,
	82:  proto.ChampionType_MORDEKAISER,
	83:  proto.ChampionType_YORICK,
	84:  proto.ChampionType_AKALI,
	85:  proto.ChampionType_KENNEN,
	86:  proto.ChampionType_GAREN,
	89:  proto.ChampionType_LEONA,
	90:  proto.ChampionType_MALZAHAR,
	91:  proto.ChampionType_TALON,
	92:  proto.ChampionType_RIVEN,
	96:  proto.ChampionType_KOGMAW,
	98:  proto.ChampionType_SHEN,
	99:  proto.ChampionType_LUX,
	101: proto.ChampionType_XERATH,
	102: proto.ChampionType_SHYVANA,
	103: proto.ChampionType_AHRI,
	104: proto.ChampionType_GRAVES,
	105: proto.ChampionType_FIZZ,
	106: proto.ChampionType_VOLIBEAR,
	107: proto.ChampionType_RENGAR,
	110: proto.ChampionType_VARUS,
	111: proto.ChampionType_NAUTILUS,
	112: proto.ChampionType_VIKTOR,
	113: proto.ChampionType_SEJUANI,
	114: proto.ChampionType_FIORA,
	115: proto.ChampionType_ZIGGS,
	117: proto.ChampionType_LULU,
	119: proto.ChampionType_DRAVEN,
	120: proto.ChampionType_HECARIM,
	121: proto.ChampionType_KHAZIX,
	122: proto.ChampionType_DARIUS,
	126: proto.ChampionType_JAYCE,
	127: proto.ChampionType_LISSANDRA,
	131: proto.ChampionType_DIANA,
	133: proto.ChampionType_QUINN,
	134: proto.ChampionType_SYNDRA,
	143: proto.ChampionType_ZYRA,
	154: proto.ChampionType_ZAC,
	157: proto.ChampionType_YASUO,
	161: proto.ChampionType_VELKOZ,
	201: proto.ChampionType_BRAUM,
	222: proto.ChampionType_JINX,
	236: proto.ChampionType_LUCIAN,
	238: proto.ChampionType_ZED,
	254: proto.ChampionType_VI,
	266: proto.ChampionType_AATROX,
	267: proto.ChampionType_NAMI,
	412: proto.ChampionType_THRESH,
}

func String2ChampionType(instr string) proto.ChampionType {
	for id, str := range proto.ChampionType_name {
		if strings.ToLower(str) == strings.ToLower(instr) {
			return proto.ChampionType(id)
		}
	}

	return proto.ChampionType_UNKNOWN
}

// This function converts a Riot champion ID into an internal Cleo
// representation.
func Rid2Cleo(rid uint32) proto.ChampionType {
	ct, exists := champion_map[rid]

	if exists {
		return ct
	} else {
		log.Println("Unknown champion ID detected:", rid)
		return proto.ChampionType_UNKNOWN
	}
}

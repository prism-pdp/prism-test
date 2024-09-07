package types

type EntityType int

const (
	SM EntityType = iota
	SP
	TPA
	SU1
	SU2
	SU3
)
var EntityList [6]EntityType = [6]EntityType{
	SM,
	SP,
	TPA,
	SU1,
	SU2,
	SU3,
}

func GetEntityName(_entity EntityType) string {
	switch _entity {
	case SM: return "SM"
	case SP: return "SP"
	case TPA: return "TPA"
	case SU1: return "SU1"
	case SU2: return "SU2"
	case SU3: return "SU3"
	}
	return "NA"
}
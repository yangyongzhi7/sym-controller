package constant

const (
	ConditionTypePrefix = "OK_"

	BlueGroup  = "blue"
	GreenGroup = "green"

	ConditionStatusTrue  = "True"
	ConditionStatusFalse = "False"

	AppLabel   = "app"
	GroupLabel = "sym-group"
)

func ConcatConditionType(group string) string {
	return ConditionTypePrefix + group
}

package constant

const (
	ConditionTypePrefix  = "OK_"
	BlueGroup            = "blue"
	GreenGroup           = "green"
	ConditionStatusTrue  = "True"
	ConditionStatusFalse = "False"
)

func ConcatConditionType(group string) string {
	return ConditionTypePrefix + group
}

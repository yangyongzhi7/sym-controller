package constant

const (
	ConditionTypePrefix = "OK_"
	BlueGroup           = "blue"
	GreenGroup          = "green"
)

func ConcatConditionType(group string) string {
	return ConditionTypePrefix + group
}

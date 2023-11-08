package batch

const (
	JobDefinitionStatusInactive string = "INACTIVE"
	JobDefinitionStatusActive   string = "ACTIVE"
)

func JobDefinitionStatus_Values() []string {
	return []string{
		JobDefinitionStatusInactive,
		JobDefinitionStatusActive,
	}
}

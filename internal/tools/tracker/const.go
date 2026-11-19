package tracker

const (
	entityTypeGoal      = "goal"
	entityTypePortfolio = "portfolio"
	entityTypeProject   = "project"
	maxPerScroll        = 1000
	maxEntitiesPerPage  = 50
)

func emptyObjectInputSchema() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

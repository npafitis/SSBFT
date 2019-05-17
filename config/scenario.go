package config

type Scenario int

const (
	NORMAL Scenario = iota
	STALE_VIEWS
	STALE_STATES
	BYZANTINE_PRIM
	STALE_REQUESTS
)

var TestCase Scenario = NORMAL

func InitialiseScenario(scenario Scenario) {
	TestCase = scenario
}

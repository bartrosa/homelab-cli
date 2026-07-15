package services

// ResolveServiceOrderForTest exposes dependency ordering for tests.
func ResolveServiceOrderForTest(ids []string) ([]string, error) {
	return resolveServiceOrder(ids)
}

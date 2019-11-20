package v1

func (g *Gang) Validate() bool {
	if g.Status.Total != g.Status.Running+g.Status.Pending+g.Status.Succeeded+g.Status.Failed+g.Status.Unknown {
		return false
	}
	return true
}

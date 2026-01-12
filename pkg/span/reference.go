package span

type Reference struct {
	TraceID string
	SpanID  string
	RefType string
}

const (
	ChildOf    = "CHILD_OF"
	FollowFrom = "FOLLOW_FROm"
)

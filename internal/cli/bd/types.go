package bd

type StatusSummary struct {
	TotalIssues             int     `json:"total_issues"`
	OpenIssues              int     `json:"open_issues"`
	InProgressIssues        int     `json:"in_progress_issues"`
	ClosedIssues            int     `json:"closed_issues"`
	BlockedIssues           int     `json:"blocked_issues"`
	DeferredIssues          int     `json:"deferred_issues"`
	ReadyIssues             int     `json:"ready_issues"`
	TombstoneIssues         int     `json:"tombstone_issues"`
	PinnedIssues            int     `json:"pinned_issues"`
	EpicsEligibleForClosure int     `json:"epics_eligible_for_closure"`
	AverageLeadTimeHours    float64 `json:"average_lead_time_hours"`
}

type RecentActivity struct {
	HoursTracked   int `json:"hours_tracked"`
	CommitCount    int `json:"commit_count"`
	IssuesCreated  int `json:"issues_created"`
	IssuesClosed   int `json:"issues_closed"`
	IssuesUpdated  int `json:"issues_updated"`
	IssuesReopened int `json:"issues_reopened"`
	TotalChanges   int `json:"total_changes"`
}

type Status struct {
	Summary        StatusSummary  `json:"summary"`
	RecentActivity RecentActivity `json:"recent_activity"`
}

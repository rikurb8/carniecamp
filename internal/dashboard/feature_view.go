package dashboard

import "sort"

func filterOpenFeatures(data dataState) []Issue {
	openIssues := make([]Issue, 0, len(data.Ready)+len(data.InProgress)+len(data.Blocked))
	openIssues = append(openIssues, data.Ready...)
	openIssues = append(openIssues, data.InProgress...)
	openIssues = append(openIssues, data.Blocked...)

	features := make([]Issue, 0, len(openIssues))
	for _, issue := range openIssues {
		if issue.IssueType != "feature" {
			continue
		}
		if !isOpenStatus(issue.Status) {
			continue
		}
		features = append(features, issue)
	}

	sort.SliceStable(features, func(i, j int) bool {
		if features[i].Priority == features[j].Priority {
			return features[i].ID < features[j].ID
		}
		return features[i].Priority < features[j].Priority
	})

	return features
}

func buildFeatureChildren(issues []Issue) map[string][]Issue {
	issueByID := make(map[string]Issue, len(issues))
	for _, issue := range issues {
		issueByID[issue.ID] = issue
	}

	childrenByFeature := make(map[string][]Issue)
	for _, issue := range issues {
		if issue.IssueType != "task" {
			continue
		}
		for _, dep := range issue.Dependencies {
			if dep.Type != "parent-child" {
				continue
			}
			parentID := dep.DependsOnID
			parent, ok := issueByID[parentID]
			if !ok || parent.IssueType != "feature" {
				continue
			}
			childrenByFeature[parentID] = append(childrenByFeature[parentID], issue)
		}
	}

	for featureID := range childrenByFeature {
		children := childrenByFeature[featureID]
		sort.SliceStable(children, func(i, j int) bool {
			if children[i].Priority == children[j].Priority {
				return children[i].ID < children[j].ID
			}
			return children[i].Priority < children[j].Priority
		})
		childrenByFeature[featureID] = children
	}

	return childrenByFeature
}

func isOpenStatus(status string) bool {
	switch status {
	case "open", "ready", "in_progress", "blocked":
		return true
	default:
		return false
	}
}

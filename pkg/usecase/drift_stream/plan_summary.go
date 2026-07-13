package drift_stream

import (
	"regexp"
	"strconv"
)

var planSummaryRegex = regexp.MustCompile(`Plan:\s+(\d+) to add,\s+(\d+) to change,\s+(\d+) to destroy`)

// ParsePlanSummary extracts the add/change/destroy counts from a plan's summary
// line. Returns nil pointers when the line is absent (e.g. init output or "No changes").
func ParsePlanSummary(planOutput string) (added, changed, destroyed *int32) {
	m := planSummaryRegex.FindStringSubmatch(planOutput)
	if m == nil {
		return nil, nil, nil
	}
	return atoi32(m[1]), atoi32(m[2]), atoi32(m[3])
}

func atoi32(s string) *int32 {
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	v := int32(n)
	return &v
}

package dto

import "time"

// DriftRateDataPoint represents a single day's drift rate data
type DriftRateDataPoint struct {
	Date             string  `json:"date"`
	TotalRuns        int64   `json:"total_runs"`
	RunsWithDrift    int64   `json:"runs_with_drift"`
	DriftRatePercent float64 `json:"drift_rate_percent"`
}

// FrequentlyDriftedProject represents a project that drifts frequently
type FrequentlyDriftedProject struct {
	Dir              string  `json:"dir"`
	Type             string  `json:"type"`
	DriftCount       int64   `json:"drift_count"`
	TotalAppearances int64   `json:"total_appearances"`
	DriftPercentage  float64 `json:"drift_percentage"`
}

// DriftFreeStreakDTO represents the current drift-free streak
type DriftFreeStreakDTO struct {
	StreakCount int64      `json:"streak_count"`
	LastRunAt   *time.Time `json:"last_run_at"`
}

// ResolutionTimeDataPoint represents daily resolution time data
type ResolutionTimeDataPoint struct {
	Date              string  `json:"date"`
	ResolutionsCount  int64   `json:"resolutions_count"`
	AvgHoursToResolve float64 `json:"avg_hours_to_resolve"`
}

// TrendsSummaryDTO provides a high-level summary of trends
type TrendsSummaryDTO struct {
	TotalRuns        int64   `json:"total_runs"`
	RunsWithDrift    int64   `json:"runs_with_drift"`
	DriftRatePercent float64 `json:"drift_rate_percent"`
	StreakCount      int64   `json:"streak_count"`
}

// RepositoryTrendsDTO is the main response for the trends endpoint
type RepositoryTrendsDTO struct {
	Summary                   TrendsSummaryDTO           `json:"summary"`
	DriftRateOverTime         []DriftRateDataPoint       `json:"drift_rate_over_time"`
	FrequentlyDriftedProjects []FrequentlyDriftedProject `json:"frequently_drifted_projects"`
	DriftFreeStreak           DriftFreeStreakDTO         `json:"drift_free_streak"`
	ResolutionTimes           []ResolutionTimeDataPoint  `json:"resolution_times"`
	DaysBack                  int                        `json:"days_back"`
}

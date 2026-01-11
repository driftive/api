package parsing

import (
	"driftive.cloud/api/pkg/model/dto"
	"driftive.cloud/api/pkg/repository/queries"
)

func ToDriftRateDataPoints(rows []queries.GetDriftRateOverTimeRow) []dto.DriftRateDataPoint {
	result := make([]dto.DriftRateDataPoint, 0, len(rows))
	for _, row := range rows {
		driftRatePercent := float64(0)
		if row.TotalRuns > 0 {
			driftRatePercent = float64(row.RunsWithDrift) / float64(row.TotalRuns) * 100
		}
		result = append(result, dto.DriftRateDataPoint{
			Date:             row.Date.Time.Format("2006-01-02"),
			TotalRuns:        row.TotalRuns,
			RunsWithDrift:    row.RunsWithDrift,
			DriftRatePercent: driftRatePercent,
		})
	}
	return result
}

func ToFrequentlyDriftedProjects(rows []queries.GetMostFrequentlyDriftedProjectsRow) []dto.FrequentlyDriftedProject {
	result := make([]dto.FrequentlyDriftedProject, 0, len(rows))
	for _, row := range rows {
		driftPercentage := float64(0)
		if row.TotalAppearances > 0 {
			driftPercentage = float64(row.DriftCount) / float64(row.TotalAppearances) * 100
		}
		result = append(result, dto.FrequentlyDriftedProject{
			Dir:              row.Dir,
			Type:             row.Type,
			DriftCount:       row.DriftCount,
			TotalAppearances: row.TotalAppearances,
			DriftPercentage:  driftPercentage,
		})
	}
	return result
}

func ToDriftFreeStreakDTO(row queries.GetDriftFreeStreakRow) dto.DriftFreeStreakDTO {
	return dto.DriftFreeStreakDTO{
		StreakCount: row.StreakCount,
		LastRunAt:   &row.LastRunAt,
	}
}

func ToResolutionTimeDataPoints(rows []queries.GetMeanTimeToResolutionRow) []dto.ResolutionTimeDataPoint {
	result := make([]dto.ResolutionTimeDataPoint, 0, len(rows))
	for _, row := range rows {
		result = append(result, dto.ResolutionTimeDataPoint{
			Date:              row.Date.Time.Format("2006-01-02"),
			ResolutionsCount:  row.ResolutionsCount,
			AvgHoursToResolve: row.AvgHoursToResolve,
		})
	}
	return result
}

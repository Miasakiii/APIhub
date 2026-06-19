package service

import (
	"apihub/internal/repository"
	"errors"
	"testing"
)

// mockFrequencyRepo is a mock implementation of frequency repository.
type mockFrequencyRepo struct {
	heatmap    [][]int64
	peakQPS    *repository.PeakQPS
	hourly     []int64
	date       string
	heatmapErr error
	peakErr    error
	hourlyErr  error
}

func (m *mockFrequencyRepo) GetHourlyHeatmap(days int) ([][]int64, error) {
	return m.heatmap, m.heatmapErr
}

func (m *mockFrequencyRepo) GetPeakQPS(days int) (*repository.PeakQPS, error) {
	return m.peakQPS, m.peakErr
}

func (m *mockFrequencyRepo) GetHourlyDistribution() ([]int64, string, error) {
	return m.hourly, m.date, m.hourlyErr
}

func TestFrequencyService_GetHourlyHeatmap(t *testing.T) {
	heatmap := make([][]int64, 7)
	for i := range heatmap {
		heatmap[i] = make([]int64, 24)
	}
	heatmap[0][10] = 5 // Monday at 10am

	repo := &mockFrequencyRepo{
		heatmap: heatmap,
	}
	svc := NewFrequencyService(repo)

	result, err := svc.GetHourlyHeatmap(7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Days != 7 {
		t.Errorf("expected days=7, got %d", result.Days)
	}
	if result.Heatmap[0][10] != 5 {
		t.Errorf("expected heatmap[0][10]=5, got %d", result.Heatmap[0][10])
	}
}

func TestFrequencyService_GetHourlyHeatmap_Error(t *testing.T) {
	repo := &mockFrequencyRepo{
		heatmapErr: errors.New("database error"),
	}
	svc := NewFrequencyService(repo)

	_, err := svc.GetHourlyHeatmap(7)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFrequencyService_GetPeakQPS(t *testing.T) {
	repo := &mockFrequencyRepo{
		peakQPS: &repository.PeakQPS{
			PeakQPS:    10.5,
			PeakMinute: "2024-01-15 14:30",
			AvgQPS:     2.3,
			PeakCount:  630,
		},
	}
	svc := NewFrequencyService(repo)

	result, err := svc.GetPeakQPS(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PeakQPS != 10.5 {
		t.Errorf("expected peak QPS 10.5, got %f", result.PeakQPS)
	}
	if result.PeakMinute != "2024-01-15 14:30" {
		t.Errorf("expected peak minute '2024-01-15 14:30', got '%s'", result.PeakMinute)
	}
}

func TestFrequencyService_GetTodayDistribution(t *testing.T) {
	hourly := make([]int64, 24)
	hourly[9] = 100
	hourly[14] = 200

	repo := &mockFrequencyRepo{
		hourly: hourly,
		date:   "2024-01-15",
	}
	svc := NewFrequencyService(repo)

	result, err := svc.GetTodayDistribution()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Date != "2024-01-15" {
		t.Errorf("expected date '2024-01-15', got '%s'", result.Date)
	}
	if result.Hourly[9] != 100 {
		t.Errorf("expected hourly[9]=100, got %d", result.Hourly[9])
	}
	if result.Hourly[14] != 200 {
		t.Errorf("expected hourly[14]=200, got %d", result.Hourly[14])
	}
}

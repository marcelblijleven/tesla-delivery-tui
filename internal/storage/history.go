package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/marcelblijleven/tesla-delivery-tui/internal/model"
)

const (
	historyDirName       = "history"
	maxHistoryEntries    = 20
)

// History manages order history persistence
type History struct {
	baseDir string
}

// NewHistory creates a new History instance
func NewHistory(configDir string) (*History, error) {
	historyDir := filepath.Join(configDir, historyDirName)
	if err := os.MkdirAll(historyDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	return &History{baseDir: historyDir}, nil
}

// historyFilePath returns the path to the history file for an order
func (h *History) historyFilePath(referenceNumber string) string {
	return filepath.Join(h.baseDir, referenceNumber+".json")
}

// LoadHistory loads the history for a specific order
func (h *History) LoadHistory(referenceNumber string) (*model.OrderHistory, error) {
	filePath := h.historyFilePath(referenceNumber)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &model.OrderHistory{
				ReferenceNumber: referenceNumber,
				Snapshots:       []model.HistoricalSnapshot{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	var history model.OrderHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to parse history file: %w", err)
	}

	return &history, nil
}

// SaveHistory saves the history for a specific order
func (h *History) SaveHistory(history *model.OrderHistory) error {
	// Prune to max entries
	if len(history.Snapshots) > maxHistoryEntries {
		history.Snapshots = history.Snapshots[len(history.Snapshots)-maxHistoryEntries:]
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	filePath := h.historyFilePath(history.ReferenceNumber)
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

// AddSnapshot adds a new snapshot to the history, returning any changes detected
func (h *History) AddSnapshot(order model.CombinedOrder) ([]model.OrderDiff, error) {
	history, err := h.LoadHistory(order.Order.ReferenceNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to load history: %w", err)
	}

	var diffs []model.OrderDiff

	// Compare with last snapshot if exists
	if len(history.Snapshots) > 0 {
		lastSnapshot := history.Snapshots[len(history.Snapshots)-1]
		diffs = compareOrders(lastSnapshot.Data, order)
	}

	// Only add snapshot if there are changes or it's the first one
	if len(diffs) > 0 || len(history.Snapshots) == 0 {
		history.Snapshots = append(history.Snapshots, model.HistoricalSnapshot{
			Timestamp: time.Now(),
			Data:      order,
		})

		if err := h.SaveHistory(history); err != nil {
			return nil, fmt.Errorf("failed to save history: %w", err)
		}
	}

	return diffs, nil
}

// GetLatestSnapshot returns the most recent snapshot for an order
func (h *History) GetLatestSnapshot(referenceNumber string) (*model.HistoricalSnapshot, error) {
	history, err := h.LoadHistory(referenceNumber)
	if err != nil {
		return nil, err
	}

	if len(history.Snapshots) == 0 {
		return nil, nil
	}

	return &history.Snapshots[len(history.Snapshots)-1], nil
}

// compareOrders delegates to the canonical model.CompareOrders
func compareOrders(old, new model.CombinedOrder) []model.OrderDiff {
	return model.CompareOrders(old, new)
}

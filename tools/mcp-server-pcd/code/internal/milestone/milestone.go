// Package milestone implements the set_milestone_status file editing logic.
// SPDX-License-Identifier: GPL-2.0-only

package milestone

import (
	"fmt"
	"strings"
)

// Status represents the pipeline state of a milestone.
type Status string

const (
	StatusPending  Status = "pending"
	StatusActive   Status = "active"
	StatusFailed   Status = "failed"
	StatusReleased Status = "released"
)

// IsValidStatus returns true if s is a valid MilestoneStatus value.
func IsValidStatus(s string) bool {
	switch Status(s) {
	case StatusPending, StatusActive, StatusFailed, StatusReleased:
		return true
	}
	return false
}

// SetMilestoneResult is returned by SetStatus on success.
type SetMilestoneResult struct {
	SpecPath       string
	MilestoneName  string
	PreviousStatus Status
	NewStatus      Status
}

// Filesystem is the interface used to read and write spec files.
type Filesystem interface {
	ReadFile(path string) (string, error)
	WriteFile(path string, content string) error
}

// OSFilesystem implements Filesystem using the real OS.
type OSFilesystem struct{}

func (OSFilesystem) ReadFile(path string) (string, error) {
	data, err := readFileOS(path)
	return string(data), err
}

func (OSFilesystem) WriteFile(path string, content string) error {
	return writeFileOS(path, []byte(content))
}

// FakeFilesystem implements Filesystem for tests.
type FakeFilesystem struct {
	Files    map[string]string
	ReadErr  map[string]error
	WriteErr map[string]error
	Written  map[string]string // tracks what was written
}

func (f *FakeFilesystem) ReadFile(path string) (string, error) {
	if f.ReadErr != nil {
		if err, ok := f.ReadErr[path]; ok {
			return "", err
		}
	}
	if f.Files == nil {
		return "", fmt.Errorf("file not found: %s", path)
	}
	content, ok := f.Files[path]
	if !ok {
		return "", fmt.Errorf("file not found: %s", path)
	}
	return content, nil
}

func (f *FakeFilesystem) WriteFile(path string, content string) error {
	if f.WriteErr != nil {
		if err, ok := f.WriteErr[path]; ok {
			return err
		}
	}
	if f.Written == nil {
		f.Written = make(map[string]string)
	}
	f.Written[path] = content
	return nil
}

// ── SetStatus ─────────────────────────────────────────────────────────────────

// SetStatus implements the set_milestone_status BEHAVIOR.
//
// Steps (per spec):
//  1. Read spec_path via fs.ReadFile; on error → error "cannot open file: {path}"
//  2. Locate ## MILESTONE: {milestone_name}; on not found → error
//  3. If new_status=active: check no other milestone already active
//  4. Record previous_status
//  5. Replace/insert Status: line as first non-blank line after ## MILESTONE: header
//  6. Write modified content via fs.WriteFile
//  7. Return SetMilestoneResult
func SetStatus(fs Filesystem, specPath, milestoneName, newStatusStr string) (SetMilestoneResult, error) {
	// Step 1: read file
	content, err := fs.ReadFile(specPath)
	if err != nil {
		return SetMilestoneResult{}, fmt.Errorf("cannot open file: %s", specPath)
	}

	newStatus := Status(newStatusStr)

	lines := strings.Split(content, "\n")

	// Step 2: locate ## MILESTONE: {milestone_name}
	milestoneHeader := "## MILESTONE: " + milestoneName
	milestoneIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == milestoneHeader {
			milestoneIdx = i
			break
		}
	}
	if milestoneIdx < 0 {
		return SetMilestoneResult{}, fmt.Errorf("MILESTONE '%s' not found in %s", milestoneName, specPath)
	}

	// Find the extent of this milestone section (until next ## heading or EOF)
	milestoneEnd := len(lines)
	for i := milestoneIdx + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			milestoneEnd = i
			break
		}
	}

	// Step 3: if new_status=active, check no other milestone already active
	if newStatus == StatusActive {
		otherActive := findOtherActiveMillestone(lines, milestoneIdx, milestoneEnd)
		if otherActive != "" {
			return SetMilestoneResult{}, fmt.Errorf(
				"Cannot set MILESTONE '%s' to active: MILESTONE '%s' is already active. Set it to released or failed first.",
				milestoneName, otherActive)
		}
	}

	// Step 4: record previous_status
	// Find Status: line within the milestone section
	statusIdx := -1
	var prevStatus Status = StatusPending
	for i := milestoneIdx + 1; i < milestoneEnd; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "Status:") {
			statusIdx = i
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, "Status:"))
			prevStatus = Status(val)
			break
		}
	}

	// Step 5: replace or insert Status: line
	// MECHANISM: Status: must be the first non-blank line after ## MILESTONE: header.
	newStatusLine := "Status: " + string(newStatus)

	if statusIdx >= 0 {
		// Replace existing Status: line
		lines[statusIdx] = newStatusLine
	} else {
		// Insert after ## MILESTONE: header as first non-blank line
		// Find the first non-blank line after the header
		insertAt := milestoneIdx + 1
		for insertAt < milestoneEnd && strings.TrimSpace(lines[insertAt]) == "" {
			insertAt++
		}
		// Insert before insertAt
		newLines := make([]string, 0, len(lines)+1)
		newLines = append(newLines, lines[:insertAt]...)
		newLines = append(newLines, newStatusLine)
		newLines = append(newLines, lines[insertAt:]...)
		lines = newLines
	}

	// Step 6: write back
	newContent := strings.Join(lines, "\n")
	if err := fs.WriteFile(specPath, newContent); err != nil {
		return SetMilestoneResult{}, fmt.Errorf("write failed: %s", specPath)
	}

	// Step 7: return result
	return SetMilestoneResult{
		SpecPath:       specPath,
		MilestoneName:  milestoneName,
		PreviousStatus: prevStatus,
		NewStatus:      newStatus,
	}, nil
}

// findOtherActiveMillestone scans all milestone sections outside [milestoneIdx, milestoneEnd)
// and returns the name of any milestone with Status: active, or "" if none.
func findOtherActiveMillestone(lines []string, thisStart, thisEnd int) string {
	inOtherMilestone := false
	currentName := ""
	currentStart := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## MILESTONE: ") {
			name := strings.TrimPrefix(trimmed, "## MILESTONE: ")
			inOtherMilestone = (i < thisStart || i >= thisEnd)
			currentName = name
			currentStart = i
			_ = currentStart
			continue
		}
		if strings.HasPrefix(trimmed, "## ") && trimmed != "## MILESTONE: " {
			// End of any milestone section
			if !strings.HasPrefix(trimmed, "## MILESTONE:") {
				inOtherMilestone = false
			}
		}
		if inOtherMilestone && (i < thisStart || i >= thisEnd) {
			if strings.HasPrefix(trimmed, "Status:") {
				val := strings.TrimSpace(strings.TrimPrefix(trimmed, "Status:"))
				if val == "active" {
					return currentName
				}
			}
		}
	}
	return ""
}

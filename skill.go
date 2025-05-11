package ultima

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"iter"
	"strings"
)

var (
	// ErrInvalidSkillIndex is returned when an invalid skill index is requested
	ErrInvalidSkillIndex = errors.New("invalid skill index")
	// ErrInvalidSkillGroupIndex is returned when an invalid skill group index is requested
	ErrInvalidSkillGroupIndex = errors.New("invalid skill group index")
)

const (
	skillUnicodeFlag = -1 // Flag in skillgrp.mul to indicate unicode encoding
	miscGroupName    = "Misc"
)

// Skill defines a single character skill in the game
type Skill struct {
	ID       int    // ID of the skill
	Name     string // Name of the skill
	IsAction bool   // True if the skill is an action (button), false if passive
}

// SkillGroup defines a group of related skills
type SkillGroup struct {
	ID     int    // ID of the skill group
	Name   string // Name of the skill group
	Skills []int  // IDs of skills that belong to this group
}

// Skill retrieves a specific skill by its ID
func (s *SDK) Skill(id int) (*Skill, error) {
	// Load the skills file
	file, err := s.loadSkills()
	if err != nil {
		return nil, fmt.Errorf("failed to load skills: %w", err)
	}

	// Check for valid index (can't check count directly, so we try to read)
	if id < 0 {
		return nil, fmt.Errorf("%w: %d", ErrInvalidSkillIndex, id)
	}

	// Read the skill data
	data, err := file.Read(uint64(id))
	if err != nil {
		return nil, fmt.Errorf("%w: %d", ErrInvalidSkillIndex, id)
	}

	// Skip empty entries
	if len(data) == 0 {
		return nil, fmt.Errorf("%w: no data for skill %d", ErrInvalidSkillIndex, id)
	}

	// Create a reader for the data
	reader := bytes.NewReader(data)

	// Read the action flag (first byte)
	isAction, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read skill action flag: %w", err)
	}

	// Read the skill name (rest of the data, null-terminated)
	nameBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read skill name: %w", err)
	}

	// Find the null terminator
	nullTermPos := bytes.IndexByte(nameBytes, 0)
	if nullTermPos != -1 {
		nameBytes = nameBytes[:nullTermPos]
	}

	// Convert to string and clean up
	name := strings.TrimSpace(string(nameBytes))
	if name == "" {
		name = fmt.Sprintf("Skill %d", id)
	}

	return &Skill{
		ID:       id,
		Name:     name,
		IsAction: isAction != 0,
	}, nil
}

// Skills returns an iterator over all defined skills
func (s *SDK) Skills() iter.Seq[*Skill] {
	return func(yield func(*Skill) bool) {
		file, err := s.loadSkills()
		if err != nil {
			return
		}

		// Use the file.Entries() iterator to go through all entries
		for id := range file.Entries() {
			skill, err := s.Skill(int(id))
			if err != nil {
				// Skip invalid entries and continue
				continue
			}

			if !yield(skill) {
				break
			}
		}
	}
}

// SkillGroup retrieves a specific skill group by its ID
func (s *SDK) SkillGroup(id int) (*SkillGroup, error) {
	// Get all skill groups
	groups, skillMap, err := s.loadSkillGroupData()
	if err != nil {
		return nil, err
	}

	// Check for valid index
	if id < 0 || id >= len(groups) {
		return nil, fmt.Errorf("%w: %d", ErrInvalidSkillGroupIndex, id)
	}

	// Find all skills that belong to this group
	var skills []int
	for skillID, groupID := range skillMap {
		if groupID == id {
			skills = append(skills, skillID)
		}
	}

	return &SkillGroup{
		ID:     id,
		Name:   groups[id],
		Skills: skills,
	}, nil
}

// SkillGroups returns an iterator over all defined skill groups
func (s *SDK) SkillGroups() iter.Seq[*SkillGroup] {
	return func(yield func(*SkillGroup) bool) {
		groups, skillMap, err := s.loadSkillGroupData()
		if err != nil {
			return
		}

		for id, name := range groups {
			// Find all skills that belong to this group
			var skills []int
			for skillID, groupID := range skillMap {
				if groupID == id {
					skills = append(skills, skillID)
				}
			}

			if !yield(&SkillGroup{
				ID:     id,
				Name:   name,
				Skills: skills,
			}) {
				break
			}
		}
	}
}

// loadSkillGroupData loads all skill group data from skillgrp.mul
func (s *SDK) loadSkillGroupData() ([]string, map[int]int, error) {
	file, err := s.loadSkillGroups()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load skill groups: %w", err)
	}

	// Read the entire file content
	data, err := file.Read(0)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read skill group data: %w", err)
	}

	// Create a reader for the data
	reader := bytes.NewReader(data)

	// Read the first int to check for Unicode encoding
	var count int32
	err = binary.Read(reader, binary.LittleEndian, &count)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read skill group count: %w", err)
	}

	isUnicode := count == skillUnicodeFlag
	if isUnicode {
		// Read the actual count value that follows the Unicode flag
		err = binary.Read(reader, binary.LittleEndian, &count)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read Unicode skill group count: %w", err)
		}
	}

	// Calculate the string length and position
	strLen := 17
	bytesPerChar := 1
	if isUnicode {
		bytesPerChar = 2
	}
	strLen *= bytesPerChar

	// Initialize the groups list with the "Misc" group at index 0
	groups := make([]string, count)
	groups[0] = miscGroupName

	// Read each group name
	for i := 1; i < int(count); i++ {
		// Position the reader at the beginning of the string
		offset := 4 + (i * strLen)
		if isUnicode {
			offset += 4 // For Unicode, add 4 bytes for the extra int32
		}

		if offset >= len(data) {
			return nil, nil, fmt.Errorf("invalid skill group data: offset %d exceeds data length %d", offset, len(data))
		}

		reader.Seek(int64(offset), io.SeekStart)

		// Read the string
		name := ""
		if isUnicode {
			var buf [2]byte
			builder := strings.Builder{}
			for {
				_, err := reader.Read(buf[:])
				if err != nil {
					return nil, nil, fmt.Errorf("failed to read Unicode skill group name: %w", err)
				}
				char := binary.LittleEndian.Uint16(buf[:])
				if char == 0 {
					break
				}
				builder.WriteRune(rune(char))
			}
			name = builder.String()
		} else {
			var buf [1]byte
			builder := strings.Builder{}
			for {
				_, err := reader.Read(buf[:])
				if err != nil {
					return nil, nil, fmt.Errorf("failed to read skill group name: %w", err)
				}
				if buf[0] == 0 {
					break
				}
				builder.WriteByte(buf[0])
			}
			name = builder.String()
		}

		groups[i] = strings.TrimSpace(name)
	}

	// Read the skill to group mappings
	skillMap := make(map[int]int)

	// Position the reader at the beginning of the skill list
	// The skill list starts after all the group name entries
	startOfSkillList := 4 + (int(count) * strLen)
	if isUnicode {
		startOfSkillList += 4 // For Unicode, add 4 bytes for the extra int32
	}

	if startOfSkillList >= len(data) {
		// If we've reached the end of the data, return what we have so far
		return groups, skillMap, nil
	}

	reader.Seek(int64(startOfSkillList), io.SeekStart)

	// Read all skill group assignments until the end of the file
	skillID := 0
	for {
		var groupID int32
		err = binary.Read(reader, binary.LittleEndian, &groupID)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, nil, fmt.Errorf("failed to read skill group assignment: %w", err)
		}

		// Assign this skill to its group
		skillMap[skillID] = int(groupID)
		skillID++
	}

	return groups, skillMap, nil
}

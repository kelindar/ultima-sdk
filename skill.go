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
	file, err := s.loadSkills()
	if err != nil {
		return nil, fmt.Errorf("failed to load skills: %w", err)
	}

	// Check for valid index (can't check count directly, so we try to read)
	if id < 0 {
		return nil, fmt.Errorf("%w: %d", ErrInvalidSkillIndex, id)
	}

	// Read the skill data
	data, _, err := file.Read(uint32(id))
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
func (s *SDK) loadSkillGroupData() (groups []string, skillMap map[int]int, err error) {
	file, err := s.loadSkillGroups()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load skillgrp.mul: %w", err)
	}

	data, _, err := file.Read(0) // Entry 0 is the whole file content
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read data from skillgrp.mul: %w", err)
	}

	if len(data) == 0 { // Empty file
		return []string{}, make(map[int]int), nil
	}

	reader := bytes.NewReader(data)

	var count int32
	isUnicode := false
	initialOffset := 4 // Base offset assuming one int32 is read for count or flag

	var tempRead int32
	err = binary.Read(reader, binary.LittleEndian, &tempRead)
	if err != nil {
		switch {
		case errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
			return nil, nil, fmt.Errorf("skillgrp.mul data too short for initial read (count/flag): %w", err)
		default:
			return nil, nil, fmt.Errorf("failed to read initial data (count/flag) from skillgrp.mul: %w", err)
		}
	}

	switch tempRead {
	case skillUnicodeFlag:
		isUnicode = true
		initialOffset += 4 // Add 4 because skillUnicodeFlag itself is an int32, and count follows
		err = binary.Read(reader, binary.LittleEndian, &count)
		if err != nil {
			switch {
			case errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
				return nil, nil, fmt.Errorf("skillgrp.mul data too short for unicode count: %w", err)
			default:
				return nil, nil, fmt.Errorf("failed to read unicode count from skillgrp.mul: %w", err)
			}
		}
	default:
		count = tempRead // The first value read was the count itself
	}

	if count < 0 {
		return nil, nil, fmt.Errorf("invalid skill group count %d in skillgrp.mul (after flag processing)", count)
	}
	if count == 0 { // No groups defined in the file
		return []string{}, make(map[int]int), nil
	}

	parsedGroups := make([]string, count)
	parsedGroups[0] = miscGroupName // Group 0 is always "Misc"

	nameSlotSize := 17
	if isUnicode {
		nameSlotSize *= 2 // 2 bytes per character for Unicode
	}

	numStoredNames := int(count) - 1

	for i := 0; i < numStoredNames; i++ {
		groupIndexInSlice := i + 1

		nameDataStart := initialOffset + (i * nameSlotSize)
		nameDataEnd := nameDataStart + nameSlotSize

		if nameDataEnd > len(data) {
			return nil, nil, fmt.Errorf("skillgrp.mul data ended prematurely reading name for group %d (expected %d bytes, got %d)", groupIndexInSlice, nameSlotSize, len(data)-nameDataStart)
		}

		nameBytes := data[nameDataStart:nameDataEnd]
		var groupName string
		if isUnicode {
			runes := make([]rune, 0, len(nameBytes)/2)
			for charIdx := 0; charIdx+1 < len(nameBytes); charIdx += 2 {
				u16char := binary.LittleEndian.Uint16(nameBytes[charIdx : charIdx+2])
				if u16char == 0 { // Null terminator
					break
				}
				runes = append(runes, rune(u16char))
			}
			groupName = string(runes)
		} else {
			nullIdx := bytes.IndexByte(nameBytes, 0)
			if nullIdx != -1 {
				groupName = string(nameBytes[:nullIdx])
			} else {
				groupName = string(nameBytes) // No null terminator found in slot
			}
		}
		parsedGroups[groupIndexInSlice] = strings.TrimSpace(groupName)
	}

	parsedSkillMap := make(map[int]int)
	offsetToSkillMappings := initialOffset + (numStoredNames * nameSlotSize)

	if offsetToSkillMappings > len(data) {
		return nil, nil, fmt.Errorf("skill map offset %d is beyond data length %d; skillgrp.mul may be corrupt or truncated", offsetToSkillMappings, len(data))
	}

	if _, err = reader.Seek(int64(offsetToSkillMappings), io.SeekStart); err != nil {
		return nil, nil, fmt.Errorf("unexpected error seeking to skill mappings at offset %d: %w", offsetToSkillMappings, err)
	}

	currentSkillID := 0
	for {
		if reader.Len() < 4 {
			if reader.Len() == 0 {
				break
			}
			return nil, nil, fmt.Errorf("corrupt skillgrp.mul: partial skill mapping data for skill ID %d (have %d bytes, need 4)", currentSkillID, reader.Len())
		}

		var groupIDFromFile int32
		if err = binary.Read(reader, binary.LittleEndian, &groupIDFromFile); err != nil {
			switch {
			case errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
				return nil, nil, fmt.Errorf("unexpected EOF/ErrUnexpectedEOF reading group ID for skill %d after length check: %w", currentSkillID, err)
			default:
				return nil, nil, fmt.Errorf("error reading group ID for skill %d from skillgrp.mul: %w", currentSkillID, err)
			}
		}
		parsedSkillMap[currentSkillID] = int(groupIDFromFile)
		currentSkillID++
	}

	return parsedGroups, parsedSkillMap, nil
}

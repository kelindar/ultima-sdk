package ultima

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSkill tests retrieving individual skills
func TestSkill(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		// Test retrieving a valid skill
		skill, err := sdk.Skill(1) // Skill ID 1 (typically "Animal Taming")
		assert.NoError(t, err)
		assert.Equal(t, 1, skill.ID)
		assert.NotEmpty(t, skill.Name)

		// Should have either IsAction true or false, but we can't predict which
		// without knowing the exact client files

		// Test invalid skill ID - negative
		_, err = sdk.Skill(-1)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidSkillIndex)

		// Test skill at higher index - may be valid depending on the client files
		// but we should at least not crash
		skill, err = sdk.Skill(100)
		if err == nil {
			assert.Equal(t, 100, skill.ID)
		}
	})
}

// TestSkills tests iterating through all skills
func TestSkills(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		// Count the number of skills and check for basic validity
		count := 0
		var firstSkill *Skill

		for skill := range sdk.Skills() {
			// Record the first skill we see
			if count == 0 {
				firstSkill = skill
			}

			// Basic validity checks for all skills
			assert.Equal(t, count, skill.ID)
			assert.NotEmpty(t, skill.Name)

			count++
		}

		// Should have found at least a few skills
		assert.Greater(t, count, 10)

		// First skill should have ID 0
		assert.Equal(t, 0, firstSkill.ID)
	})
}

// TestSkillGroup tests retrieving individual skill groups
func TestSkillGroup(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		// Test retrieving the "Misc" group (always group 0)
		group, err := sdk.SkillGroup(0)
		require.NoError(t, err)
		assert.Equal(t, 0, group.ID)
		assert.Equal(t, "Misc", group.Name)

		// Test an invalid group
		_, err = sdk.SkillGroup(-1)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidSkillGroupIndex)
	})
}

// TestSkillGroups tests iterating through all skill groups
func TestSkillGroups(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		groups := make([]SkillGroup, 0)
		for group := range sdk.SkillGroups() {
			groups = append(groups, *group)
		}

		// Should have at least a few groups
		assert.GreaterOrEqual(t, len(groups), 1)
		assert.Equal(t, miscGroupName, groups[0].Name)
	})
}

// TestSkillAndGroupRelationship tests the relationship between skills and groups
func TestSkillAndGroupRelationship(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		// Keep track of which skill IDs we've seen in any group
		skillsInGroups := make(map[int]bool)

		// Iterate through all skill groups
		for group := range sdk.SkillGroups() {
			// Each group should have a valid relationship reference
			for _, skillID := range group.Skills {
				// Mark this skill as having been assigned to a group
				skillsInGroups[skillID] = true

				// Verify we can access the skill
				skill, err := sdk.Skill(skillID)
				if assert.NoError(t, err) {
					assert.Equal(t, skillID, skill.ID)
				}
			}
		}

		// We should have found at least a few skills in groups
		assert.NotEmpty(t, skillsInGroups)
	})
}

// TestSkillFileManipulation tests low-level file operations for skills
func TestSkillFileManipulation(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		// Test loading the skills file
		file, err := sdk.loadSkills()
		assert.NoError(t, err)
		assert.NotNil(t, file)

		// Test loading the skill groups file
		groupFile, err := sdk.loadSkillGroups()
		assert.NoError(t, err)
		assert.NotNil(t, groupFile)

		// Test reading group data
		groups, skillMap, err := sdk.loadSkillGroupData()
		assert.NoError(t, err)
		assert.NotNil(t, groups)
		assert.NotNil(t, skillMap)

		// We should have at least the "Misc" group
		assert.GreaterOrEqual(t, len(groups), 1)
		assert.Equal(t, miscGroupName, groups[0])
	})
}

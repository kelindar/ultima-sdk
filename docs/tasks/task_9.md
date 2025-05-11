# Task 9: Port Skills.cs and SkillGroups.cs to skill.go

## Objective

Implement support for loading and accessing skill and skill group data from `skills.idx`, `skills.mul`, and `skillgrp.mul` files, which define the character skills and their grouping in the UO client.

## C# Reference Implementation Analysis

The C# implementation consists of two main classes:

- `Skills` - Handles loading skill definitions from `skills.idx` and `skills.mul`
- `SkillGroups` - Handles loading skill group definitions from `skillgrp.mul`

These files contain information about all character skills, whether they're action-based or passive, their names, and how they're grouped in the UI. The skill system is integral to character progression in UO.

## Work Items

1. Create a new file `skill.go` in the root package.

2. Define the `Skill` struct that represents a single skill entry:

   ```go
   type Skill struct {
       ID        int
       IsAction  bool   // True if the skill is an action, false if passive
       Name      string // Name of the skill
   }
   ```

3. Define the `SkillGroup` struct that represents a group of skills:

   ```go
   type SkillGroup struct {
       ID     int
       Name   string
       Skills []int
   }
   ```

4. Add methods to the SDK struct for accessing skills and skill groups:

   ```go
   // Skill retrieves a specific skill by its ID
   func (s *SDK) Skill(id int) (*Skill, error) {
       // Implementation for retrieving a specific skill
   }

   // Skills returns an iterator over all defined skills
   func (s *SDK) Skills() iter.Seq[*Skill] {
       // Implementation for iterating over skills
   }

   // SkillGroup retrieves a specific skill group by its ID
   func (s *SDK) SkillGroup(id int) (*SkillGroup, error) {
       // Implementation for retrieving a specific skill group
   }

   // SkillGroups returns an iterator over all defined skill groups
   func (s *SDK) SkillGroups() iter.Seq[*SkillGroup] {
       // Implementation for iterating over skill groups
   }
   ```

5. Write comprehensive unit tests in `skill_test.go`:
   - Test loading skill data
   - Test loading skill group data
   - Test accessing skills by ID
   - Test accessing skill groups by ID

## Key Considerations

- The `skills.idx` file contains index information mapping to entries in `skills.mul`
- The `skills.mul` file contains the actual skill data (name, action flag)
- The `skillgrp.mul` file contains group definitions and their member skills
- Character encoding in these files is typically Windows-1252, not UTF-8
- Some skills may have empty names (especially at the end of the file)
- The relationship between skills and skill groups is many-to-one (a skill belongs to exactly one group)
- Consider lazy loading to avoid unnecessary file operations
- Ensure proper error handling for all file operations and index bounds checking

## Expected Output

A complete implementation that allows:

- Loading skill and skill group data
- Accessing individual skills and skill groups by ID
- Iterating over all available skills and skill groups
- Finding all skills within a specific group

## Verification

- Compare loaded skill and group names with the C# implementation
- Verify the action flag is correctly set for action-oriented skills
- Ensure all skills are properly assigned to their correct groups
- Test with edge cases like skills with empty names or the last skill in the file

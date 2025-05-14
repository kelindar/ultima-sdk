# Ultima SDK Go Porting Plan

This document outlines the plan for porting the C# Ultima SDK to Go.

## Overall Goal

To create an idiomatic Go library that provides the same functionality as the original C# Ultima SDK, allowing Go applications to read and interact with Ultima Online data files. The port should be verified against the C# implementation for correctness.

**Before starting, please read and adhere to the guidelines outlined in `docs/coding-standards.md`.**

**Goal:** Create a functionally equivalent Go library for reading and accessing Ultima Online data files. The resulting library should have an **idiomatic Go public API**, as defined in `api.md`. For each piece of functionality ported, the implementer **must verify its behavior against the C# reference implementation and write comprehensive unit tests** to ensure correctness. While the internal structure of the C# version can serve as a guide, the primary focus for internals is correctness and testability, not a line-by-line port if a more idiomatic or efficient Go approach exists for the same logic. The implementation **must be correct and verified** against the original functionality.

**Target Go Version:** 1.22+ (to leverage features like `iter.Seq` and other modern Go capabilities)

**Execution Workflow:**

1.  **One Task at a Time:** Address only _one_ checklist item marked with ❌ at a time.
2.  **Implement & Test:** Implement the required functionality. **Crucially, verify its behavior against the C# reference implementation and write corresponding unit tests using `testify/assert` and the specified test data.**
3.  **Mark for Review:** Once a task and its tests are complete, change its status from ❌ to ❓.
4.  **Wait for Review:** Do not proceed to the next task until the current one marked with ❓ has been manually reviewed and marked as ✅ (Done) or reverted to ❌ (Needs Rework).
5.  **Ask for Clarification:** If a task description is unclear, ambiguous, or seems too large/complex for a single step, ask for clarification before proceeding.
6.  **Use the Test Data and `runWith` Helper:** All unit tests that require an initialized SDK instance **must** use the provided `runWith(t, func(sdk *SDK) ` helper function located in `sdk_test.go`. This function ensures the SDK is correctly initialized with the test data from `d:\Workspace\Go\src\github.com\kelindar\ultima-sdk-testdata\`. This is crucial for verifying the correctness of the ported code by comparing outputs or behavior with the C# reference.

## Porting Instructions & Checklist

Follow these steps sequentially. Each step involves translating the corresponding C# code into idiomatic Go, placing it in the proposed structure, writing tests, verifying against the C# reference, and marking for review (❌ -> ❓).

1.  **[✅] Setup Go Project:**
    - Create the root directory `ultima-sdk-go`.
    - Run `go mod init github.com/kelindar/ultima-sdk-go` (or your preferred module path).
    - Create the `internal/` subdirectory for implementation details.
2.  **[✅] Implement `SDK` Struct and Lifecycle (`sdk.go`):**
    - Define `SDK` struct.
    - Implement `Open(directory string) (*SDK, error)` and `sdk.Close()` methods.
    - _Sub-task: Write basic tests for Open/Close._
    - \_Sub-task: Verify against C# `Ultima.Client` constructor and `Files.Initialize`/`Files.Dispose` patterns for resource management.
3.  **[✅] Implement MUL File Reading Utilities (`internal/mul`):**
    - Create `internal/mul` package.
    - Define and expose a `Reader` struct for handling MUL file specific reading logic.
    - Port relevant functionality from `C# BinaryExtensions.cs` for reading primitive types and structures from MUL files.
    - _Sub-task: Write unit tests for `mul.Reader` methods._
    - _Sub-task: Verify against C# `BinaryReader` extensions in the context of MUL file parsing._
4.  **[✅] Implement UOP File Handling (`internal/uop`):**
    - Create `internal/uop` package.
    - Define and expose a `Reader` struct for handling UOP file format parsing.
    - Port UOP file format parsing logic from `C# Helpers/UopUtils.cs`.
    - Implement functions to extract specific file entries from UOP archives (compression will be handled by `internal/file`).
    - _Sub-task: Write tests for parsing UOP files and `uop.Reader` methods._
    - _Sub-task: Verify UOP parsing logic against C# implementation._
5.  **[✅] Implement Unified File Access Logic (`internal/uofile`):**
    - Create `internal/uofile` package.
    - Define a `Reader` interface that `internal/mul.Reader` and `internal/uop.Reader` will implement.
    - Define and expose a `File` struct that uses the `Reader` interface for accessing file data.
    - Port `C# FileIndex.cs` logic for representing file entries (MUL/IDX, UOP), loading index files (`*.idx`), and providing unified access through the `File` struct.
    - `internal/file` will depend on `internal/mul` and `internal/uop` for their respective `Reader` implementations.
    - _Sub-task: Write unit tests for `File` struct methods, index loading, and compression utilities._
    - _Sub-task: Verify against C# `FileIndex` behavior and Zlib decompression results._
6.  **[✅] Implement Top-Level File Accessors (integrate into `sdk.go`):**
    - Port C# `Files` class concept: methods on `SDK` to get access to specific game files (e.g., `sdk.Art()`) using the `File` struct from `internal/uofile`.
    - Make sure it is done using lazy loading to avoid opening too many or unused files at once.
    - _Sub-task: Verify correct file sources (MUL/UOP) are accessed via `internal/file.File`._
7.  **[✅] Port `Hues.cs` -> `hue.go`:**
    - Define `Hue` struct. Implement loading for `hues.mul`.
    - _Sub-task: Write tests to load hues and verify data against C# output._
8.  **[✅] Port `RadarCol.cs` -> `radarcol.go`:**
    - Implement loading for `radarcol.mul`.
    - _Sub-task: Write tests to load radar colors and verify data._
9.  **[✅] Port `Skills.cs`, `SkillGroups.cs` -> `skill.go`:**
    - Define `Skill`, `SkillGroup` structs. Implement loading for `skills.idx`, `skills.mul`, `skillgrp.mul`.
    - _Sub-task: Write tests for loading skills/groups and verify data._
10. **[✅] Port `SpeechList.cs` -> `speech.go`:**
    - Define `Speech` struct. Implement loading for `speech.mul`.
    - _Sub-task: Write tests for loading speech entries._
11. **[✅] Port `StringList.cs`, `StringEntry.cs` -> `cliloc.go`:**
    - Implement `Cliloc` loading (e.g., `Cliloc.enu`). Provide `SDK.GetString(id int) string`.
    - _Sub-task: Write tests for loading cliloc files and retrieving strings._
12. **[✅] Implement Internal Image Handling Utilities (`internal/bitmap`):**
    - Handle 16-bit ARGB1555 pixel format to `image.Image` conversion.
    - Implement hue application logic.
    - _Sub-task: Write tests for pixel format conversion and hue application._
    - _Sub-task: Verify image output against C# rendering or pixel data._
13. **[✅] Port `TileData.cs` & `Helpers/TileDataHelpers.cs` -> `tiledata.go`:**
    - Define `LandTileData`, `StaticItemData` structs. Implement loading for `tiledata.mul`.
    - Integrate `TileDataHelpers.cs` logic.
    - _Sub-task: Write tests for loading tile data and verify properties._
14. **[✅] Port `Light.cs` -> `light.go`:**
    - Define `Light` struct. Implement loading for `light.idx`, `light.mul`.
    - _Sub-task: Write tests for loading light data._
15. **[❌] Port `Art.cs` -> `art.go`:**
    - Define `ArtTile` struct. Use `internal/bitmap`.
    - _Sub-task: Write tests to load art tiles, verify dimensions/image data._
16. **[❌] Port `Gumps.cs` -> `gump.go`:**
    - Define `Gump`, `GumpInfo` structs. Use `internal/bitmap`.
    - _Sub-task: Write tests for loading gumps, verify dimensions/image data._
17. **[❌] Port `Textures.cs` -> `texture.go`:**
    - Define `Texture` struct. Use `internal/bitmap`.
    - _Sub-task: Write tests for loading textures and verify image data._
18. **[❌] Port `Sound.cs`, `WaveFormat.cs` -> `sound.go`:**
    - Define `Sound` struct. Implement loading for `soundidx.mul`, `sound.mul`. Handle WAV/PCM.
    - _Sub-task: Write tests for loading sound entries and verifying data properties._
19. **[❌] Port `ASCIIFont.cs`, `UnicodeFont.cs` -> `font.go`:**
    - Define `Font` interface, `FontCharacterInfo`. Implement loading for `fonts.mul`.
    - _Sub-task: Write tests for loading fonts and character data._
20. **[❌] Port `Map.cs` & `Helpers/MapHelper.cs` -> `map.go`:**
    - Define `GameMap`, `MapTileInfo`, `LandTile`, `StaticTile`. Implement loading for map files, statics, patches.
    - Integrate `MapHelper.cs` logic.
    - _Sub-task: Write extensive tests for map/static reading, including patches._
    - _Sub-task: Verify against C# tile/static details at specific coordinates._
21. **[❌] Port Animations (`anim.go`, `internal/anim`):**
    - Define `Animation`, `AnimationFrame`. Port logic from `Animations.cs`, `Animdata.cs`, `BodyConverter.cs`, `BodyTable.cs`.
    - Read `anim.idx`/`.mul`, `animdata.mul`, `bodyconv.def`, `body.def`. Use `internal/bitmap`.
    - _Sub-task: Write tests for loading animations, verify frame counts/image data._
22. **[❌] Port `Multis.cs` & `Helpers/MultiHelpers.cs` -> `multi.go`:**
    - Define `Multi`, `MultiItem`. Implement loading for `multi.idx`, `multi.mul`.
    - Integrate `MultiHelpers.cs` logic.
    - _Sub-task: Write tests for loading multis and verifying structure._
23. **[❌] Port `Verdata.cs` -> `verdata.go`:**
    - Define `VerdataPatch` struct. Implement loading for `verdata.mul`.
    - _Sub-task: Write tests for loading verdata._

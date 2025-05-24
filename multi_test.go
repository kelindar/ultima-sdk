// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMulti_Load(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		multi, err := sdk.Multi(0)
		assert.NoError(t, err)
		assert.NotNil(t, multi)
		assert.Greater(t, len(multi.Items), 0)
		item := multi.Items[0]
		assert.NotZero(t, item.Item)

		img, err := multi.Image()
		assert.NoError(t, err)
		assert.NotNil(t, img)
		//savePng(img, "multi.png")
	})
}

func TestMulti_MarshalCSV(t *testing.T) {
	// Create a test Multi with sample data
	multi := &Multi{
		Items: []MultiItem{
			{Item: 100, X: -10, Y: 5, Z: 0, Flags: 1, Cliloc: 0},
			{Item: 200, X: 0, Y: 0, Z: 10, Flags: 2, Cliloc: 1},
			{Item: 300, X: 15, Y: -20, Z: -5, Flags: 4, Cliloc: 2},
		},
	}

	// Marshal to CSV
	csvData, err := multi.ToCSV()
	assert.NoError(t, err)

	// Parse the CSV output
	reader := csv.NewReader(bytes.NewReader(csvData))
	records, err := reader.ReadAll()
	assert.NoError(t, err)

	// Verify header
	expectedHeader := []string{"item", "x", "y", "z", "flags", "cliloc"}
	assert.Equal(t, expectedHeader, records[0])

	// Verify data rows
	assert.Len(t, records, 4) // header + 3 data rows

	// Verify first data row
	expectedRow1 := []string{"100", "-10", "5", "0", "1", "0"}
	assert.Equal(t, expectedRow1, records[1])

	// Verify second data row
	expectedRow2 := []string{"200", "0", "0", "10", "2", "1"}
	assert.Equal(t, expectedRow2, records[2])

	// Verify third data row
	expectedRow3 := []string{"300", "15", "-20", "-5", "4", "2"}
	assert.Equal(t, expectedRow3, records[3])
}

func TestMulti_MarshalCSV_EmptyMulti(t *testing.T) {
	// Test with empty Multi
	multi := &Multi{Items: []MultiItem{}}

	csvData, err := multi.ToCSV()
	assert.NoError(t, err)

	// Parse the CSV output
	reader := csv.NewReader(bytes.NewReader(csvData))
	records, err := reader.ReadAll()
	assert.NoError(t, err)

	// Should only have header
	assert.Len(t, records, 1)
	expectedHeader := []string{"item", "x", "y", "z", "flags", "cliloc"}
	assert.Equal(t, expectedHeader, records[0])
}

func TestSDK_UnmarshalMultiCSV(t *testing.T) {
	sdk := &SDK{} // Create a minimal SDK for testing

	// Test CSV data with the expected format
	csvData := `item,x,y,z,flags,cliloc
100,-10,5,0,1,0
200,0,0,10,2,1
300,15,-20,-5,4,2`

	multi, err := sdk.MultiFromCSV([]byte(csvData))
	assert.NoError(t, err)
	assert.NotNil(t, multi)
	assert.Equal(t, sdk, multi.sdk)
	assert.Len(t, multi.Items, 3)

	// Verify first item
	item1 := multi.Items[0]
	assert.Equal(t, uint16(100), item1.Item)
	assert.Equal(t, int16(-10), item1.X)
	assert.Equal(t, int16(5), item1.Y)
	assert.Equal(t, int16(0), item1.Z)
	assert.Equal(t, uint32(1), item1.Flags)
	assert.Equal(t, uint32(0), item1.Cliloc)

	// Verify second item
	item2 := multi.Items[1]
	assert.Equal(t, uint16(200), item2.Item)
	assert.Equal(t, int16(0), item2.X)
	assert.Equal(t, int16(0), item2.Y)
	assert.Equal(t, int16(10), item2.Z)
	assert.Equal(t, uint32(2), item2.Flags)
	assert.Equal(t, uint32(1), item2.Cliloc)

	// Verify third item
	item3 := multi.Items[2]
	assert.Equal(t, uint16(300), item3.Item)
	assert.Equal(t, int16(15), item3.X)
	assert.Equal(t, int16(-20), item3.Y)
	assert.Equal(t, int16(-5), item3.Z)
	assert.Equal(t, uint32(4), item3.Flags)
	assert.Equal(t, uint32(2), item3.Cliloc)
}

func TestSDK_UnmarshalMultiCSV_OnlyHeader(t *testing.T) {
	sdk := &SDK{}

	// Test CSV with only header
	csvData := `item,x,y,z,flags,cliloc`

	multi, err := sdk.MultiFromCSV([]byte(csvData))
	assert.NoError(t, err)
	assert.NotNil(t, multi)
	assert.Len(t, multi.Items, 0) // No items, just header
}

func TestSDK_UnmarshalMultiCSV_MinimalColumns(t *testing.T) {
	sdk := &SDK{}

	// Test CSV with only required columns (item, x, y, z)
	csvData := `item,x,y,z
100,-10,5,0
200,0,0,10`

	multi, err := sdk.MultiFromCSV([]byte(csvData))
	assert.NoError(t, err)
	assert.NotNil(t, multi)
	assert.Len(t, multi.Items, 2)

	// Verify first item (flags and unk1 should default to 0)
	item1 := multi.Items[0]
	assert.Equal(t, uint16(100), item1.Item)
	assert.Equal(t, int16(-10), item1.X)
	assert.Equal(t, int16(5), item1.Y)
	assert.Equal(t, int16(0), item1.Z)
	assert.Equal(t, uint32(0), item1.Flags)  // Default
	assert.Equal(t, uint32(0), item1.Cliloc) // Default

	// Verify second item
	item2 := multi.Items[1]
	assert.Equal(t, uint16(200), item2.Item)
	assert.Equal(t, int16(0), item2.X)
	assert.Equal(t, int16(0), item2.Y)
	assert.Equal(t, int16(10), item2.Z)
	assert.Equal(t, uint32(0), item2.Flags)  // Default
	assert.Equal(t, uint32(0), item2.Cliloc) // Default
}

func TestSDK_UnmarshalMultiCSV_PartialColumns(t *testing.T) {
	sdk := &SDK{}

	// Test CSV with 5 columns (item, x, y, z, flags) - missing cliloc
	csvData := `item,x,y,z,flags
100,-10,5,0,1
200,0,0,10,2`

	multi, err := sdk.MultiFromCSV([]byte(csvData))
	assert.NoError(t, err)
	assert.NotNil(t, multi)
	assert.Len(t, multi.Items, 2)

	// Verify first item (unk1 should default to 0)
	item1 := multi.Items[0]
	assert.Equal(t, uint16(100), item1.Item)
	assert.Equal(t, int16(-10), item1.X)
	assert.Equal(t, int16(5), item1.Y)
	assert.Equal(t, int16(0), item1.Z)
	assert.Equal(t, uint32(1), item1.Flags)  // From CSV
	assert.Equal(t, uint32(0), item1.Cliloc) // Default

	// Verify second item
	item2 := multi.Items[1]
	assert.Equal(t, uint16(200), item2.Item)
	assert.Equal(t, int16(0), item2.X)
	assert.Equal(t, int16(0), item2.Y)
	assert.Equal(t, int16(10), item2.Z)
	assert.Equal(t, uint32(2), item2.Flags)  // From CSV
	assert.Equal(t, uint32(0), item2.Cliloc) // Default
}

func TestSDK_UnmarshalMultiCSV_FlexibleHeaders(t *testing.T) {
	sdk := &SDK{}

	// Test CSV with different header names (should still work)
	csvData := `Item ID,X Offset,Y Offset,Z Offset,Flags,Cliloc
100,-10,5,0,1,0
200,0,0,10,2,1`

	multi, err := sdk.MultiFromCSV([]byte(csvData))
	assert.NoError(t, err)
	assert.NotNil(t, multi)
	assert.Len(t, multi.Items, 2)

	// Verify the data is parsed correctly despite different headers
	item1 := multi.Items[0]
	assert.Equal(t, uint16(100), item1.Item)
	assert.Equal(t, int16(-10), item1.X)
	assert.Equal(t, int16(5), item1.Y)
	assert.Equal(t, int16(0), item1.Z)
	assert.Equal(t, uint32(1), item1.Flags)
	assert.Equal(t, uint32(0), item1.Cliloc)
}

func TestSDK_UnmarshalMultiCSV_InvalidData(t *testing.T) {
	sdk := &SDK{}

	// Test CSV with invalid data types
	csvData := `item,x,y,z,flags,cliloc
invalid,-10,5,0,1,0`

	multi, err := sdk.MultiFromCSV([]byte(csvData))
	assert.Error(t, err)
	assert.Nil(t, multi)
	assert.Contains(t, err.Error(), "invalid ItemID")
}

func TestSDK_UnmarshalMultiCSV_EmptyData(t *testing.T) {
	sdk := &SDK{}

	// Test with empty CSV data
	multi, err := sdk.MultiFromCSV([]byte(""))
	assert.Error(t, err)
	assert.Nil(t, multi)
	assert.Contains(t, err.Error(), "CSV data is empty")
}

func TestMulti_MarshalUnmarshalCSV_RoundTrip(t *testing.T) {
	// Test that we can marshal and unmarshal and get the same data back
	original := &Multi{
		Items: []MultiItem{
			{Item: 100, X: -10, Y: 5, Z: 0, Flags: 1, Cliloc: 0},
			{Item: 200, X: 0, Y: 0, Z: 10, Flags: 2, Cliloc: 1},
			{Item: 300, X: 15, Y: -20, Z: -5, Flags: 4, Cliloc: 2},
		},
	}

	// Marshal to CSV
	csvData, err := original.ToCSV()
	assert.NoError(t, err)

	// Unmarshal back
	sdk := &SDK{}
	restored, err := sdk.MultiFromCSV(csvData)
	assert.NoError(t, err)

	// Compare items (ignore SDK pointer)
	assert.Len(t, restored.Items, len(original.Items))
	for i, originalItem := range original.Items {
		restoredItem := restored.Items[i]
		assert.Equal(t, originalItem.Item, restoredItem.Item)
		assert.Equal(t, originalItem.X, restoredItem.X)
		assert.Equal(t, originalItem.Y, restoredItem.Y)
		assert.Equal(t, originalItem.Z, restoredItem.Z)
		assert.Equal(t, originalItem.Flags, restoredItem.Flags)
		assert.Equal(t, originalItem.Cliloc, restoredItem.Cliloc)
	}
}

func savePng(img image.Image, name string) error {
	file, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	return png.Encode(file, img)
}

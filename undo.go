package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"slices"

	c "finance-planner-tui/constants"

	"gopkg.in/yaml.v3"
)

// sets the FP.SelectedProfile & config to the value specified by the current
// undo buffer
//
// warning: naively assumes that the FP.UndoBufferPos has already been set to a
// valid value and updates the currently selected config & profile accordingly
func pushUndoBufferChange() {
	n := FP.SelectedProfile.Name

	err := yaml.Unmarshal(FP.UndoBuffer[FP.UndoBufferPos], &FP.Config)
	if err != nil {
		FP.ProfileStatusText.SetText(fmt.Sprintf(
			"%v%v%v",
			FP.Colors["ProfileStatusTextError"],
			FP.T["UndoBufferPushValueConfigUnmarshalFailure"],
			c.ResetStyle,
		))
	}

	// set the FP.SelectedProfile to the latest FP.UndoBuffer's config
	for i := range FP.Config.Profiles {
		if FP.Config.Profiles[i].Name == n {
			FP.SelectedProfile = &(FP.Config.Profiles[i])
		}
	}
}

// moves 1 step backward in the FP.UndoBuffer
func undo() {
	undoBufferLen := len(FP.UndoBuffer)
	newUndoBufferPos := FP.UndoBufferPos - 1

	if newUndoBufferPos < 0 {
		// nothing to undo - at beginning of FP.UndoBuffer
		FP.ProfileStatusText.SetText(fmt.Sprintf(
			"%v%v [%v/%v]%v",
			FP.Colors["ProfileStatusTextPassive"],
			FP.T["UndoBufferNothingToUndo"],
			FP.UndoBufferPos+1,
			undoBufferLen,
			c.ResetStyle,
		))

		return
	}

	FP.UndoBufferPos = newUndoBufferPos

	pushUndoBufferChange()

	FP.ProfileStatusText.SetText(fmt.Sprintf(
		"%v%v: [%v/%v]%v",
		FP.Colors["ProfileStatusTextPassive"],
		FP.T["UndoBufferUndoAction"],
		FP.UndoBufferPos+1,
		undoBufferLen,
		c.ResetStyle,
	))

	populateProfilesPage()
	getTransactionsTable()
	FP.TransactionsTable.Select(FP.SelectedProfile.SelectedRow, FP.SelectedProfile.SelectedColumn)
	FP.App.SetFocus(FP.TransactionsTable)
}

// moves 1 step forward in the FP.UndoBuffer
func redo() {
	undoBufferLen := len(FP.UndoBuffer)
	undoBufferLastPos := undoBufferLen - 1
	newUndoBufferPos := FP.UndoBufferPos + 1

	if newUndoBufferPos > undoBufferLastPos {
		// nothing to redo - at end of FP.UndoBuffer
		FP.ProfileStatusText.SetText(fmt.Sprintf(
			"%v%v [%v/%v]%v",
			FP.Colors["ProfileStatusTextPassive"],
			FP.T["UndoBufferNothingToRedo"],
			FP.UndoBufferPos+1,
			undoBufferLen,
			c.ResetStyle,
		))

		return
	}

	FP.UndoBufferPos = newUndoBufferPos

	pushUndoBufferChange()

	FP.ProfileStatusText.SetText(fmt.Sprintf(
		"%v%v: [%v/%v]%v",
		FP.Colors["ProfileStatusTextPassive"],
		FP.T["UndoBufferRedoAction"],
		FP.UndoBufferPos+1,
		undoBufferLen,
		c.ResetStyle,
	))

	populateProfilesPage()
	getTransactionsTable()
	FP.TransactionsTable.Select(FP.SelectedProfile.SelectedRow, FP.SelectedProfile.SelectedColumn)
	FP.App.SetFocus(FP.TransactionsTable)
}

// Uses gzip to compress bytes.
func compress(input []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)

	if _, err := gz.Write(input); err != nil {
		return []byte{}, fmt.Errorf("%v: %w", FP.T["UndoBufferCompressionWriteError"], err)
	}

	if err := gz.Close(); err != nil {
		return []byte{}, fmt.Errorf("%v: %w", FP.T["UndoBufferCompressionCloseError"], err)
	}

	return b.Bytes(), nil
}

// Uses gzip to decompress bytes.
func decompress(input []byte) ([]byte, error) {
	var b bytes.Buffer

	gz, err := gzip.NewReader(&b)
	if err != nil {
		return []byte{}, fmt.Errorf("%v: %w", FP.T["UndoBufferDecompressionReaderInitError"], err)
	}

	if _, err := gz.Read(input); err != nil {
		return []byte{}, fmt.Errorf("%v: %w", FP.T["UndoBufferDecompressionWriteError"], err)
	}

	if err := gz.Close(); err != nil {
		return []byte{}, fmt.Errorf("%v: %w", FP.T["UndoBufferDecompressionCloseError"], err)
	}

	return b.Bytes(), nil
}

// attempts to place the current config at FP.UndoBuffer[FP.UndoBufferPos+1] but
// only if there were actual changes.
//
// also updates the status text accordingly
//
// TODO: This needs to be refactored and it needs to have better error handling.
// Specifically, it needs to better alert the user when saving fails in a way
// that is extremely invasive. Currently the small status text cannot show the
// entire error. As for the refactoring - since this function is run very often,
// the fewer operations the better.
func modified() {
	if FP.SelectedProfile == nil {
		return
	}

	FP.SelectedProfile.Modified = true
	cr, cc := FP.TransactionsTable.GetSelection()
	FP.SelectedProfile.SelectedColumn = cc
	FP.SelectedProfile.SelectedRow = cr

	// marshal to detect differences between this config and the latest
	// config in the undo buffer
	if len(FP.UndoBuffer) >= 1 {
		b, err := yaml.Marshal(FP.Config)
		if err != nil {
			FP.ProfileStatusText.SetText(fmt.Sprintf(
				"%v%v%v",
				FP.Colors["ProfileStatusTextError"],
				FP.T["UndoBufferCannotMarshalConfigError"],
				c.ResetStyle,
			))
		}

		var bo []byte

		if FP.Config.DisableGzipCompressionInUndoBuffer {
			bo = FP.UndoBuffer[FP.UndoBufferPos]
		} else {
			bo, err = decompress(FP.UndoBuffer[FP.UndoBufferPos])
			if err != nil {
				FP.ProfileStatusText.SetText(fmt.Sprintf(
					"%v%v%v",
					FP.Colors["ProfileStatusTextError"],
					FP.T["UndoBufferConfigDecompressionError"],
					c.ResetStyle,
				))
			}
		}

		if string(bo) == string(b) {
			// no difference between this config and previous one
			FP.ProfileStatusText.SetText(fmt.Sprintf(
				"%v%v [%v/%v]%v",
				FP.Colors["ProfileStatusTextError"],
				FP.T["UndoBufferNoChange"],
				FP.UndoBufferPos+1,
				len(FP.UndoBuffer),
				c.ResetStyle,
			))

			return
		}
	}

	// if the FP.UndoBufferPos is not at the end of the FP.UndoBuffer, then all
	// values after FP.UndoBufferPos need to be deleted
	if FP.UndoBufferPos != len(FP.UndoBuffer)-1 {
		FP.UndoBuffer = slices.Delete(FP.UndoBuffer, FP.UndoBufferPos, len(FP.UndoBuffer))
	}

	getTransactionsTable()

	// now that we've ensured that we are actually at the end of the buffer,
	// proceed to insert this config into the FP.UndoBuffer
	b, err := yaml.Marshal(FP.Config)
	if err != nil {
		FP.ProfileStatusText.SetText(fmt.Sprintf(
			"%v%v%v",
			FP.Colors["ProfileStatusTextError"],
			FP.T["UndoBufferCannotMarshalConfigError"],
			c.ResetStyle,
		))
	}

	var bgz []byte

	if FP.Config.DisableGzipCompressionInUndoBuffer {
		bgz = b
	} else {
		// push compressed bytes into the undo buffer to save on RAM :)
		bgz, err = compress(b)
		if err != nil {
			FP.ProfileStatusText.SetText(fmt.Sprintf(
				"%v%v%v",
				FP.Colors["ProfileStatusTextError"],
				FP.T["UndoBufferConfigCompressionError"],
				c.ResetStyle,
			))
		}
	}

	FP.UndoBuffer = append(FP.UndoBuffer, bgz)
	FP.UndoBufferPos = len(FP.UndoBuffer) - 1

	totalUndoBufferSize := 0
	for i := range FP.UndoBuffer {
		totalUndoBufferSize += len(FP.UndoBuffer[i])
	}

	// TODO: restrict the length of the buffer based on the configured max

	pushUndoBufferChange()
	FP.ProfileStatusText.SetText(fmt.Sprintf(
		"%v%v*%v[%v/%v %vkB]%v",
		FP.Colors["ProfileStatusTextModifiedMarker"],
		c.ResetStyle,
		FP.Colors["ProfileStatusTextPassive"],
		// FP.FlagConfigFile,
		FP.UndoBufferPos+1,
		len(FP.UndoBuffer),
		// float64(len(bgz)/1000),
		// float64(len(b)/1000),
		float64(totalUndoBufferSize/1000),
		c.ResetStyle,
	))
}

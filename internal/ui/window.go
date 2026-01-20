package ui

import (
	"fmt"
	"strings"

	"appy/internal/appimage"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ShowMetadata displays the AppImage metadata in a GUI window
func ShowMetadata(metadata *appimage.Metadata) {
	a := app.New()
	w := a.NewWindow("AppImage Info - " + metadata.Name)

	content := buildMetadataLayout(metadata)
	w.SetContent(content)
	w.Resize(fyne.NewSize(500, 400))
	w.ShowAndRun()
}

// ShowError displays an error dialog
func ShowError(err error) {
	a := app.New()
	w := a.NewWindow("Error")

	errorMsg := err.Error()
	label := widget.NewLabel(errorMsg)
	label.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(
		widget.NewIcon(theme.ErrorIcon()),
		label,
		widget.NewButton("OK", func() {
			a.Quit()
		}),
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(400, 200))
	w.ShowAndRun()
}

// buildMetadataLayout creates the layout for displaying metadata
func buildMetadataLayout(metadata *appimage.Metadata) fyne.CanvasObject {
	// Icon
	var iconWidget fyne.CanvasObject
	if metadata.Icon != nil && len(metadata.Icon.Data) > 0 {
		iconResource := fyne.NewStaticResource("icon", metadata.Icon.Data)
		iconImage := canvas.NewImageFromResource(iconResource)
		iconImage.FillMode = canvas.ImageFillContain
		iconImage.SetMinSize(fyne.NewSize(128, 128))
		iconWidget = iconImage
	} else {
		// Use default application icon if no icon available
		iconWidget = widget.NewIcon(theme.ComputerIcon())
	}

	// Name (using RichText for larger text)
	nameLabel := widget.NewRichTextFromMarkdown("# " + metadata.Name)

	// Description
	descriptionLabel := widget.NewLabel(metadata.Description)
	descriptionLabel.Wrapping = fyne.TextWrapWord
	if metadata.Description == "" {
		descriptionLabel.SetText("No description available")
	}

	// Version
	versionLabel := widget.NewLabel(fmt.Sprintf("Version: %s", metadata.Version))

	// Categories
	categoriesStr := strings.Join(metadata.Categories, ", ")
	categoriesLabel := widget.NewLabel(fmt.Sprintf("Categories: %s", categoriesStr))

	// File path
	pathLabel := widget.NewLabel(fmt.Sprintf("File: %s", metadata.FilePath))
	pathLabel.Wrapping = fyne.TextWrapWord
	pathLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Layout
	iconContainer := container.NewCenter(iconWidget)

	infoContainer := container.NewVBox(
		nameLabel,
		widget.NewSeparator(),
		descriptionLabel,
		widget.NewLabel(""),
		versionLabel,
		categoriesLabel,
		widget.NewLabel(""),
		pathLabel,
	)

	mainContainer := container.NewBorder(
		iconContainer,
		nil,
		nil,
		nil,
		container.NewPadded(infoContainer),
	)

	return mainContainer
}

// ShowErrorDialog displays an error dialog on an existing window
func ShowErrorDialog(w fyne.Window, err error) {
	dialog.ShowError(err, w)
}

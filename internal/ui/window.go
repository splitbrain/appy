package ui

import (
	"fmt"
	"strings"

	"appy/internal/appimage"
	"appy/internal/installer"

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
	w := a.NewWindow("Appy - " + metadata.Name)

	content := buildMetadataLayout(metadata, w)
	w.SetContent(content)
	w.Resize(fyne.NewSize(500, 450))
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
func buildMetadataLayout(metadata *appimage.Metadata, w fyne.Window) fyne.CanvasObject {
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

	// Categories
	categoriesStr := strings.Join(metadata.Categories, ", ")
	categoriesLabel := widget.NewLabel(fmt.Sprintf("Categories: %s", categoriesStr))

	// File path
	pathLabel := widget.NewLabel(fmt.Sprintf("File: %s", metadata.FilePath))
	pathLabel.Wrapping = fyne.TextWrapWord
	pathLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Layout
	iconContainer := container.NewCenter(iconWidget)

	// Build info container, conditionally including description
	infoItems := []fyne.CanvasObject{
		nameLabel,
		widget.NewSeparator(),
	}

	// Only add description if it exists
	if metadata.Description != "" {
		descriptionLabel := widget.NewLabel(metadata.Description)
		descriptionLabel.Wrapping = fyne.TextWrapWord
		infoItems = append(infoItems, descriptionLabel, widget.NewLabel(""))
	}

	infoItems = append(infoItems,
		categoriesLabel,
		widget.NewLabel(""),
		pathLabel,
	)

	infoContainer := container.NewVBox(infoItems...)

	// Check for existing installation
	sanitizedName := installer.SanitizeAppName(metadata.Name)
	isInstalledPath := installer.IsInstalledPath(metadata.Name, metadata.FilePath)
	willOverwrite := !isInstalledPath && installer.IsInstalled(metadata.Name)

	// Show warning if installation will overwrite
	if willOverwrite {
		info, err := installer.GetInstalledInfo(sanitizedName)
		if err == nil {
			warningBox := container.NewVBox(
				widget.NewLabel(""),
				widget.NewSeparator(),
			)

			// Warning header with icon
			warningHeader := container.NewHBox(
				widget.NewIcon(theme.WarningIcon()),
				widget.NewLabel("Installation will overwrite existing version"),
			)
			warningBox.Add(warningHeader)

			// Show original filename if available
			if info.OriginalFilename != "" {
				originalLabel := widget.NewLabel(fmt.Sprintf("Currently installed: %s", info.OriginalFilename))
				originalLabel.TextStyle = fyne.TextStyle{Italic: true}
				warningBox.Add(originalLabel)
			}

			infoContainer.Add(warningBox)
		}
	}

	// Install/Uninstall button
	var actionButton *widget.Button
	var updateButton func()

	updateButton = func() {
		// Check if the current file is the installed version
		isInstalledPath := installer.IsInstalledPath(metadata.Name, metadata.FilePath)
		if isInstalledPath {
			actionButton.SetText("Uninstall")
			actionButton.OnTapped = func() {
				handleUninstall(w, metadata, updateButton)
			}
		} else {
			actionButton.SetText("Install")
			actionButton.OnTapped = func() {
				handleInstall(w, metadata, updateButton)
			}
		}
		actionButton.Refresh()
	}

	actionButton = widget.NewButton("", func() {})
	updateButton()

	buttonContainer := container.NewCenter(actionButton)

	mainContainer := container.NewBorder(
		iconContainer,
		buttonContainer,
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

// handleInstall performs installation with confirmation
func handleInstall(w fyne.Window, metadata *appimage.Metadata, updateButton func()) {
	// Check if already installed (will overwrite)
	if installer.IsInstalled(metadata.Name) {
		// Show simple confirmation for overwrite
		dialog.ShowConfirm(
			"Confirm Installation",
			"This will overwrite the existing installation. Continue?",
			func(confirmed bool) {
				if confirmed {
					performInstall(w, metadata, updateButton)
				}
			},
			w,
		)
		return
	}

	// Proceed with installation
	performInstall(w, metadata, updateButton)
}

// performInstall executes the installation
func performInstall(w fyne.Window, metadata *appimage.Metadata, updateButton func()) {
	err := installer.Install(metadata)
	if err != nil {
		dialog.ShowError(err, w)
		return
	}

	showSuccessDialog(w, "Application installed successfully!")
	updateButton()
}

// handleUninstall performs uninstallation with confirmation
func handleUninstall(w fyne.Window, metadata *appimage.Metadata, updateButton func()) {
	sanitizedName := installer.SanitizeAppName(metadata.Name)

	// Show confirmation dialog
	dialog.ShowConfirm(
		"Confirm Uninstall",
		fmt.Sprintf("Are you sure you want to uninstall %s?", metadata.Name),
		func(confirmed bool) {
			if !confirmed {
				return
			}

			err := installer.Uninstall(sanitizedName)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			showSuccessDialog(w, "Application uninstalled successfully!")
			updateButton()
		},
		w,
	)
}

// showSuccessDialog displays success message after install/uninstall
func showSuccessDialog(w fyne.Window, message string) {
	d := dialog.NewInformation("Success", message, w)
	d.SetOnClosed(func() {
		fyne.CurrentApp().Quit()
	})
	d.Show()
}

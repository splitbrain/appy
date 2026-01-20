package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"appy/appimage"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// showMainWindow displays the AppImage metadata in a GUI window
func showMainWindow(metadata *appimage.Metadata) {
	a := app.New()
	w := a.NewWindow("Appy - " + metadata.Name)

	content := buildMainWindowLayout(metadata, w)
	w.SetContent(content)
	w.Resize(fyne.NewSize(500, 450))
	w.ShowAndRun()
}

// buildMainWindowLayout creates the layout for displaying metadata
func buildMainWindowLayout(metadata *appimage.Metadata, w fyne.Window) fyne.CanvasObject {
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

	// Filename
	pathLabel := widget.NewLabel(filepath.Base(metadata.FilePath))
	pathLabel.Wrapping = fyne.TextWrapWord

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

	// Add overwrite warning if needed
	if warning := buildOverwriteWarning(metadata); warning != nil {
		infoContainer.Add(warning)
	}

	// Create action button
	buttonContainer := buildActionButton(metadata, w)

	mainContainer := container.NewBorder(
		iconContainer,
		buttonContainer,
		nil,
		nil,
		container.NewPadded(infoContainer),
	)

	return mainContainer
}

// buildOverwriteWarning creates a warning box if installation will overwrite an existing version
func buildOverwriteWarning(metadata *appimage.Metadata) fyne.CanvasObject {
	isInstalledPath := appimage.IsInstalledPath(metadata.Name, metadata.FilePath)
	if isInstalledPath || !appimage.IsInstalled(metadata.Name) {
		return nil
	}

	warningBox := container.NewVBox(
		widget.NewLabel(""),
		widget.NewSeparator(),
		container.NewHBox(
			widget.NewIcon(theme.WarningIcon()),
			widget.NewLabel("Installation will overwrite existing version"),
		),
	)

	sanitizedName := appimage.SanitizeAppName(metadata.Name)
	if desktopPath, err := appimage.GetDesktopFilePath(sanitizedName); err == nil {
		if originalFilename, err := appimage.GetOriginalFilename(desktopPath); err == nil && originalFilename != "" {
			originalLabel := widget.NewLabel(fmt.Sprintf("Currently installed: %s", originalFilename))
			originalLabel.TextStyle = fyne.TextStyle{Italic: true}
			warningBox.Add(originalLabel)
		}
	}

	return warningBox
}

// buildActionButton creates the Install/Uninstall button with its update logic
func buildActionButton(metadata *appimage.Metadata, w fyne.Window) fyne.CanvasObject {
	var actionButton *widget.Button
	var updateButton func()

	updateButton = func() {
		isInstalledPath := appimage.IsInstalledPath(metadata.Name, metadata.FilePath)
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

	return container.NewCenter(actionButton)
}

// showError displays an error dialog
func showError(err error) {
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

// handleInstall performs installation with confirmation if overwriting
func handleInstall(w fyne.Window, metadata *appimage.Metadata, updateButton func()) {
	if appimage.IsInstalled(metadata.Name) {
		dialog.ShowConfirm(
			"Confirm Installation",
			"This will overwrite the existing installation. Continue?",
			func(confirmed bool) {
				if confirmed {
					if err := appimage.Install(metadata); err != nil {
						dialog.ShowError(err, w)
						return
					}
					showSuccessDialog(w, "Application installed successfully!")
					updateButton()
				}
			},
			w,
		)
		return
	}

	if err := appimage.Install(metadata); err != nil {
		dialog.ShowError(err, w)
		return
	}
	showSuccessDialog(w, "Application installed successfully!")
	updateButton()
}

// handleUninstall performs uninstallation with confirmation
func handleUninstall(w fyne.Window, metadata *appimage.Metadata, updateButton func()) {
	sanitizedName := appimage.SanitizeAppName(metadata.Name)

	// Show confirmation dialog
	dialog.ShowConfirm(
		"Confirm Uninstall",
		fmt.Sprintf("Are you sure you want to uninstall %s?", metadata.Name),
		func(confirmed bool) {
			if !confirmed {
				return
			}

			err := appimage.Uninstall(sanitizedName)
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

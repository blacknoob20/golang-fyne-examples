package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

var selectedFilePath string // Global variable to store the selected file path

func onClickFileButton(myWindow fyne.Window, lblPathValue *widget.Label) {
	fileDialog := dialog.NewFileOpen(
		func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}

			if reader == nil {
				return
			}

			selectedFilePath = reader.URI().Path() // Save the path to the global variable
			lblPathValue.SetText(selectedFilePath) // Update the label with the new path
			fmt.Println("Selected file:", reader.URI().Path())
		}, myWindow)
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".p12"}))
	fileDialog.Show()
}

func render(myWindow fyne.Window) fyne.CanvasObject {
	lblFile := widget.NewLabel("Archivo P12")
	lblPathValue := widget.NewLabel(selectedFilePath)
	// TODO: Esto no funciona, hacer mas pequeña la fuente
	// lblPathValue.Theme().Size(fyne.ThemeSizeName('S'))

	openFileButton := widget.NewButton("Seleccionar Archivo", func() {
		onClickFileButton(myWindow, lblPathValue)
	})

	lblPass := widget.NewLabel("Contraseña")
	input := widget.NewEntry()
	input.SetPlaceHolder("Enter text...")

	btnFirmar := widget.NewButton("Firmar", func() {
		log.Println("Button tapped")
	})

	// * Crear un contenedor para el formulario
	formContainer := container.New(layout.NewFormLayout(), lblFile, openFileButton, lblPass, input)

	lblPath := widget.NewLabel("Ruta")
	// * Crear un contenedor para los resultados
	resultContainer := container.New(layout.NewFormLayout(), lblPath, lblPathValue)

	// * Crear un contenedor horizontal para centrar el botón
	btnFirmarContainer := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), btnFirmar, layout.NewSpacer())

	// * Crear un contenedor vertical que combine el formulario y el botón centrado
	return container.NewVBox(formContainer, btnFirmarContainer, resultContainer)
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Form Layout")

	myWindow.SetContent(render(myWindow))
	myWindow.Resize(fyne.NewSize(720, 480))
	myWindow.ShowAndRun()
}

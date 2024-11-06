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

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Form Layout")

	label1 := widget.NewLabel("Archivo P12")
	openFileButton := widget.NewButton("Seleccionar Archivo", func() {
		fileDialog := dialog.NewFileOpen(
			func(reader fyne.URIReadCloser, err error) {
				if err != nil {
					dialog.ShowError(err, myWindow)
					return
				}

				if reader == nil {
					return
				}

				fmt.Println("Selected file:", reader.URI().Path())
			}, myWindow)
		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".txt", ".md"}))
		fileDialog.Show()
	})

	label2 := widget.NewLabel("Contraseña")
	input := widget.NewEntry()
	input.SetPlaceHolder("Enter text...")

	btnSave := widget.NewButton("Firmar", func() {
		log.Println("Button tapped")
	})

	// Crear un contenedor para el formulario
	formContainer := container.New(layout.NewFormLayout(), label1, openFileButton, label2, input)

	// Crear un contenedor horizontal para centrar el botón
	buttonContainer := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), btnSave, layout.NewSpacer())

	// Crear un contenedor vertical que combine el formulario y el botón centrado
	finalContainer := container.NewVBox(formContainer, buttonContainer)

	myWindow.SetContent(finalContainer)
	myWindow.Resize(fyne.NewSize(720, 480))
	myWindow.ShowAndRun()
}

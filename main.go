package main

import (
	"bytes"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

var selectedFilePath string // Global variable to store the selected file path
var p12File []byte

type CertInfo struct {
	Issuer    string `json:"issuer"`
	ValidFrom string `json:"valid_from"`
	ValidTo   string `json:"valid_to"`
	CertPEM   string `json:"cert_pem"`
}

func extractCertDataWithOpenSSL(p12Data []byte, password string) (*CertInfo, error) {
	p12File := "temp.p12"
	pemFile := "extracted.pem"

	// Write the p12 data to a temporary file
	err := os.WriteFile(p12File, p12Data, 0644)
	if err != nil {
		return nil, err
	}
	defer os.Remove(p12File) // Clean up

	// Prepare the openssl command
	cmd := exec.Command("openssl", "pkcs12", "-in", p12File, "-out", pemFile, "-nodes", "-password", "pass:"+password, "-legacy")

	// Execute the command
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	defer os.Remove(pemFile) // Clean up

	// Read the extracted PEM file
	pemData, err := os.ReadFile(pemFile)
	if err != nil {
		return nil, err
	}

	// Parse the PEM data to extract issuer, valid from, and valid to
	issuer, validFrom, validTo, err := parsePEMData(pemData)
	if err != nil {
		return nil, err
	}

	certInfo := &CertInfo{
		Issuer:    issuer,
		ValidFrom: validFrom,
		ValidTo:   validTo,
		CertPEM:   string(pemData),
	}

	return certInfo, nil
}

func parsePEMData(pemData []byte) (issuer, validFrom, validTo string, err error) {
	cmd := exec.Command("openssl", "x509", "-noout", "-text")
	cmd.Stdin = bytes.NewReader(pemData)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", "", "", err
	}

	output := out.String()

	issuer = extractField(output, "Issuer:")
	validFrom = extractField(output, "Not Before:")
	validTo = extractField(output, "Not After :")

	if issuer == "" || validFrom == "" || validTo == "" {
		return "", "", "", errors.New("failed to extract all certificate fields")
	}

	return strings.TrimSpace(issuer), strings.TrimSpace(validFrom), strings.TrimSpace(validTo), nil
}

func extractField(output, field string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, field) {
			return strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		}
	}
	return ""
}

// func extractCertData(p12Data []byte, password string) (*x509.Certificate, error) {
// 	// Decode the P12 file
// 	privateKey, certificate, err := pkcs12.Decode(p12Data, password)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Type assertion to ensure the private key is of type *rsa.PrivateKey
// 	if _, ok := privateKey.(*rsa.PrivateKey); !ok {
// 		return nil, fmt.Errorf("expected RSA private key")
// 	}

// 	return certificate, nil
// }

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
			defer reader.Close()

			selectedFilePath = reader.URI().Path() // Save the path to the global variable
			lblPathValue.SetText(selectedFilePath) // Update the label with the new path

			// Read the entire file p12Data
			p12Data, err := io.ReadAll(reader)
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}

			p12File = p12Data
		}, myWindow)
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".p12"}))
	fileDialog.Show()
}

func onClickFirmarButton(myWindow fyne.Window, lblDetallesValue *widget.Label, txtPass *widget.Entry) {
	password := txtPass.Text
	cert, err := extractCertDataWithOpenSSL(p12File, password)
	if err != nil {
		dialog.ShowError(err, myWindow)
		return
	}

	// Detect MIME type using the first 512 bytes (common practice for MIME detection)
	var buffer bytes.Buffer
	buffer.Write([]byte(cert.CertPEM))

	mimeType := http.DetectContentType(buffer.Bytes())

	// Formatted string using fmt.Sprintf
	metaDataP12 := fmt.Sprintf("MIME type: %s", mimeType) // Extract the certificate data

	// Output certificate data
	// certData, _ := x509.MarshalPKIXPublicKey(cert.PublicKey)
	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte(cert.CertPEM)})
	fmt.Print(pemCert)

	certData := fmt.Sprintf("Issuer: %s\nValid From: %s\nValid To: %s\nCert:\n%s", cert.Issuer, cert.ValidFrom, cert.ValidTo, pemCert)

	formattedString := fmt.Sprintf("%s\n\n%s", metaDataP12, certData)
	lblDetallesValue.SetText(formattedString)

	dialog.ShowInformation("Aviso", "Firma exitosa", myWindow)
}

func render(myWindow fyne.Window) fyne.CanvasObject {
	lblFile := widget.NewLabel("Archivo P12")

	lblPass := widget.NewLabel("Contraseña")
	txtPass := widget.NewEntry()
	txtPass.SetPlaceHolder("Enter text...")

	lblPath := widget.NewLabel("Ruta")
	lblPathValue := widget.NewLabel("")

	lblDetalles := widget.NewLabel("Detalles")
	lblDetallesValue := widget.NewLabel("")
	// TODO: Esto no funciona, hacer mas pequeña la fuente
	// lblPathValue.Theme().Size(fyne.ThemeSizeName('S'))

	openFileButton := widget.NewButton("Seleccionar Archivo", func() {
		onClickFileButton(myWindow, lblPathValue)
	})

	btnFirmar := widget.NewButton("Firmar", func() {
		onClickFirmarButton(myWindow, lblDetallesValue, txtPass)
	})

	// * Crear un contenedor para el formulario
	formContainer := container.New(layout.NewFormLayout(), lblFile, openFileButton, lblPass, txtPass)

	// * Crear un contenedor para los resultados
	resultContainer := container.New(layout.NewFormLayout(), lblPath, lblPathValue, lblDetalles, lblDetallesValue)

	// * Crear un contenedor horizontal para centrar el botón
	btnFirmarContainer := container.NewHBox(layout.NewSpacer(), btnFirmar, layout.NewSpacer())

	// * Crear un contenedor vertical que combine el formulario y el botón centrado
	return container.NewVBox(formContainer, btnFirmarContainer, resultContainer)
}

func main() {
	myApp := app.NewWithID("my.fyne.myapp")
	myWindow := myApp.NewWindow("Firma Electrónica")

	myWindow.SetContent(render(myWindow))
	myWindow.Resize(fyne.NewSize(720, 480))
	myWindow.ShowAndRun()
}

package archivos

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"os"
	"path/filepath"
	"strings"

	_ "image/jpeg"
	_ "image/png"

	"github.com/HugoSmits86/nativewebp"
	"github.com/nfnt/resize"

	// "github.com/nickalie/go-webpbin"

	"github.com/vincent-petithory/dataurl"
)

func isImage(mimeType string) bool {
	switch mimeType {
	case "image/jpeg", "image/png", "image/bmp", "image/webp":
		return true
	default:
		return false
	}
}

func SubirImagen(img64, prefix, idbol string) (string, error) {
	mydataurl, err := dataurl.DecodeString(img64)
	if err != nil {
		return "", err
	}

	if !isImage(mydataurl.MediaType.ContentType()) {
		return "", errors.New("el archivo debe ser una imagen")
	}

	sizeInBytes := len(mydataurl.Data)
	sizeInKB := float64(sizeInBytes) / 1024.0
	if sizeInKB > 2048 {
		return "", errors.New("la imagen no debe exceder 2 MB")
	}

	dir := "res"
	fileName := prefix + "-" + idbol + ".webp"
	filePath := filepath.Join(dir, fileName)

	// Crear la carpeta si no existe
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	image, err := reducirImageWebp(img64)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(filePath, image, 0644)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func reducirImageWebp(img64 string) ([]byte, error) {
	index := strings.Index(img64, "base64,") + 7
	img64 = img64[index:]

	imgData, err := base64.StdEncoding.DecodeString(img64)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewReader(imgData))

	if err != nil {
		return nil, err
	}

	originalBounds := img.Bounds()
	originalWidth := originalBounds.Dx()

	maxAncho := 1024
	if originalWidth > maxAncho {
		img = resize.Resize(uint(maxAncho), 0, img, resize.Lanczos3)
	}

	var buf bytes.Buffer
	/* options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 90)
	if err != nil {
		return nil, err
	} */

	// if err = webp.Encode(&buf, img, &webp.Options{Lossless: true, Quality: 60, Method: 0}); err != nil {
	if err := nativewebp.Encode(&buf, img, &nativewebp.Options{}); err != nil {
		// if err := webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: 90}); err != nil {
		// if err := webp.Encode(&buf, img, options); err != nil {
		// if err := webpbin.Encode(&buf, img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func GetImagen(url string) (string, error) {
	data, err := os.ReadFile(url)
	if err != nil {
		t := err.Error()
		if strings.Contains(t, "no such") {
			data = []byte("iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAYAAACNbyblAAAAHElEQVQI12P4//8/w38GIAXDIBKE0DHxgljNBAAO")
		} else {
			return "", err
		}
	}

	base64Data := base64.StdEncoding.EncodeToString(data)

	// Construye el Data URI
	dataURI := "data:image/webp;base64," + base64Data
	return dataURI, nil
}

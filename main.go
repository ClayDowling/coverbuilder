package main

import (
	"embed"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"strings"

	svg "github.com/ajstarks/svgo"
	"github.com/muesli/reflow/wordwrap"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"golang.org/x/image/draw"

	_ "golang.org/x/image/tiff"
)

const DPI = 96

const FINAL_X = 1560
const FINAL_Y = 1200
const IMG_X = 780.0
const IMG_Y = 600.0

//go:embed dowling-studio-logo-white.png
var f embed.FS

func main() {
	var title string
	var img1 image.Image
	var img2 image.Image
	var img3 image.Image
	var img4 image.Image

	flag.StringVar(&title, "title", "Untitled", "Card set title")
	flag.Parse()
	if len(flag.Args()) != 4 {
		log.Fatal("4 images are required for a cover, only %d provided", len(flag.Args()))
	}

	img1 = openImage(flag.Arg(0))
	img2 = openImage(flag.Arg(1))
	img3 = openImage(flag.Arg(2))
	img4 = openImage(flag.Arg(3))

	logofile, _ := f.Open("dowling-studio-logo-white.png")
	logoimg, _, err := image.Decode(logofile)
	if err != nil {
		log.Fatalf("Decoding logo image: %v", err)
	}

	outimg := image.NewRGBA(image.Rect(0, 0, FINAL_X, FINAL_Y))

	scaled1 := scaleImage(img1)
	scaled2 := scaleImage(img2)
	scaled3 := scaleImage(img3)
	scaled4 := scaleImage(img4)

	draw.Copy(outimg, image.Point{}, scaled1, scaled1.Bounds(), draw.Over, nil)
	draw.Copy(outimg, image.Point{IMG_X + 1, 0}, scaled2, scaled2.Bounds(), draw.Over, nil)
	draw.Copy(outimg, image.Point{0, IMG_Y + 1}, scaled3, scaled3.Bounds(), draw.Over, nil)
	draw.Copy(outimg, image.Point{IMG_X + 1, IMG_Y + 1}, scaled4, scaled4.Bounds(), draw.Over, nil)
	draw.Copy(outimg, image.Point{56, 900}, logoimg, logoimg.Bounds(), draw.Over, nil)

	for x := 0; x < FINAL_X; x++ {
		outimg.SetRGBA(x, IMG_Y, color.RGBA{0xff, 0xff, 0xff, 0xff})
	}
	for y := 0; y < FINAL_X; y++ {
		outimg.SetRGBA(IMG_X, y, color.RGBA{0xff, 0xff, 0xff, 0xff})
	}

	csr := cases.Lower(language.AmericanEnglish)
	filebase := strings.Replace(csr.String(title), " ", "-", -1)
	pngfile := fmt.Sprintf("%s.png", filebase)
	svgfile := fmt.Sprintf("%s.svg", filebase)
	dst, err := os.Create(pngfile)
	if err != nil {
		log.Fatal(err)
	}
	defer dst.Close()
	err = png.Encode(dst, outimg)
	if err != nil {
		log.Fatal(err)
	}

	svgwriter, err := os.Create(svgfile)
	if err != nil {
		log.Fatalf("Creating SVG file: %v", err)
	}
	canvas := svg.New(svgwriter)
	canvas.Start(int(8.5*DPI), int(11*DPI), "style=\"font-family: Arial\"")

	backBlock(canvas, title, pngfile, 0.5*DPI, 0.5*DPI)
	backBlock(canvas, title, pngfile, 0.5*DPI, 5.5*DPI)

	canvas.End()

}

func backBlock(canvas *svg.SVG, title string, pngfile string, origin_x int, origin_y int) {
	cardLines := strings.Split(wordwrap.String(title, 10), "\n")
	canvas.Title(fmt.Sprintf("Cover for %s card pack", title))
	canvas.Image(int(0.5*DPI)+origin_x, int(0.25*DPI)+origin_y, 499 /*(5.2*DPI)*/, int(4*DPI), pngfile)
	canvas.Textlines(int(3.125*DPI)+origin_x, int(0.75*DPI)+origin_y, cardLines, 38, 40, "white", "middle")
	//canvas.Image(int(0.75*DPI)+origin_x, int(3.25*DPI)+origin_y, 216, 80, "dowling-studio-logo-white.png")
	descx := int(6.25*DPI) + origin_x
	descy := int(2.25*DPI) + origin_y
	svgattributes := fmt.Sprintf("style=\"fill:black; font-size: 14pt; text-anchor: middle;\" transform=\"rotate(-90 %d %d)\"", descx, descy)
	canvas.Text(descx, descy, "Eight 4x6 blank note cards with envelopes", svgattributes)
	canvas.Rect(0+origin_x, 0+origin_y, int(6.5*DPI), int(4.5*DPI), "fill=\"none\"", "stroke=\"black\"", "stroke-width=\"1\"")
}

func openImage(filename string) image.Image {
	reader, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	img, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	return img
}

func scaleImage(img image.Image) image.Image {

	myx := float64(img.Bounds().Max.X)
	myy := float64(img.Bounds().Max.Y)

	if myy > myx {
		img = rotateImage(img)
		myx = float64(img.Bounds().Max.X)
		myy = float64(img.Bounds().Max.Y)
	}

	scalex := float64(IMG_X) / myx
	scaley := float64(IMG_Y) / myy

	var scale float64

	if scalex > scaley {
		scale = scalex
	} else {
		scale = scaley
	}

	newx := myx * scale
	newy := myy * scale

	scaledimg := image.NewNRGBA(image.Rect(0, 0, int(newx), int(newy)))
	draw.ApproxBiLinear.Scale(scaledimg, scaledimg.Bounds(), img, img.Bounds(), draw.Over, nil)

	return scaledimg
}

func rotateImage(img image.Image) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	bounds.Max.X, bounds.Max.Y = bounds.Max.Y, bounds.Max.X

	dimg := image.NewRGBA(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			org := img.At(x, y)
			dimg.Set(y, width-x, org)
		}
	}

	return dimg
}

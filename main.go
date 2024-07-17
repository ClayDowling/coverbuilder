package main

import (
	"flag"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"

	"golang.org/x/image/draw"

	_ "golang.org/x/image/tiff"
)

const FINAL_X = 1560
const FINAL_Y = 1200
const IMG_X = 780.0
const IMG_Y = 600.0

func main() {
	var outfile string
	var img1 image.Image
	var img2 image.Image
	var img3 image.Image
	var img4 image.Image

	flag.StringVar(&outfile, "output", "cover.png", "Destination file")
	flag.Parse()
	if len(flag.Args()) != 4 {
		log.Fatal("4 images are required for a cover, only %d provided", len(flag.Args()))
	}

	img1 = openImage(flag.Arg(0))
	img2 = openImage(flag.Arg(1))
	img3 = openImage(flag.Arg(2))
	img4 = openImage(flag.Arg(3))

	outimg := image.NewRGBA(image.Rect(0, 0, FINAL_X, FINAL_Y))

	scaled1 := scaleImage(img1)
	scaled2 := scaleImage(img2)
	scaled3 := scaleImage(img3)
	scaled4 := scaleImage(img4)

	draw.Copy(outimg, image.Point{}, scaled1, scaled1.Bounds(), draw.Over, nil)
	draw.Copy(outimg, image.Point{IMG_X + 1, 0}, scaled2, scaled2.Bounds(), draw.Over, nil)
	draw.Copy(outimg, image.Point{0, IMG_Y + 1}, scaled3, scaled3.Bounds(), draw.Over, nil)
	draw.Copy(outimg, image.Point{IMG_X + 1, IMG_Y + 1}, scaled4, scaled4.Bounds(), draw.Over, nil)

	for x := 0; x < FINAL_X; x++ {
		outimg.SetRGBA(x, IMG_Y, color.RGBA{0xff, 0xff, 0xff, 0xff})
	}
	for y := 0; y < FINAL_X; y++ {
		outimg.SetRGBA(IMG_X, y, color.RGBA{0xff, 0xff, 0xff, 0xff})
	}

	dst, err := os.Create(outfile)
	if err != nil {
		log.Fatal(err)
	}
	defer dst.Close()
	err = png.Encode(dst, outimg)
	if err != nil {
		log.Fatal(err)
	}
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

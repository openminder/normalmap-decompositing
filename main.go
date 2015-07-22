package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

// Reading files requires checking most calls for errors.
// This helper will streamline our error checks below.
func check(e error) {
	if e != nil {
		fmt.Println(e)
	}
}

func FileNotFound(e error) {
	if e != nil {
		fmt.Println("Error: Source File not found!")
		fmt.Println(e)
	}
}

func DecodeError(e error) {
	if e != nil {
		fmt.Println("Error: File could not be opened! Please check file format. Only png is supported.")
	}
}

type Float2 struct {
	x float64
	y float64
}

/*
inline float2 EncodeFloatRG( float v )
{
        float2 kEncodeMul = float2(1.0, 255.0);
        float kEncodeBit = 1.0/255.0;
        float2 enc = kEncodeMul * v;
        enc = frac (enc);
        enc.x -= enc.y * kEncodeBit;
        return enc;
}
*/
func EncodeFloatRG(v float64) Float2 {
	encodeMul := Float2{1.0, 255.0}
	encodeBit := 1.0 / 255.0
	enc := Float2{encodeMul.x * v, encodeMul.y * v}
	enc.x = enc.x - math.Floor(enc.x)
	enc.y = enc.y - math.Floor(enc.y)
	enc.x -= enc.y * encodeBit
	return enc
}

/*
inline float DecodeFloatRG( float2 enc )
{
        float2 kDecodeDot = float2(1.0, 1/255.0);
        return dot( enc, kDecodeDot );
}

func (v Vector) Dot(ov Vector) float64 { return v.X*ov.X + v.Y*ov.Y + v.Z*ov.Z }
*/
func DecodeFloatRG(v Float2) float64 {
	decodeDot := Float2{1.0, 1 / 255.0}
	return v.x*decodeDot.x + v.y*decodeDot.y
}

func Normalize(v float64) float64 {
	return v / 65535
}

func DeNormalize(v float64) float64 {
	return v * 255
}

func encode(filename string, outfilename string) {
	// filename := "CarBody_Normal.png"
	infile, err := os.Open(filename)
	FileNotFound(err)
	defer infile.Close()

	// Decode will figure out what type of image is in the file on its own.
	// We just have to be sure all the image packages we want are imported.
	src, _, err := image.Decode(infile)
	DecodeError(err)

	// Create a new image ignoring the alpha channel for color calculation
	bounds := src.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	newImage := image.NewNRGBA(image.Rectangle{image.Point{0, 0}, image.Point{w, h}})

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			oldColor := src.At(x, y)
			r, g, _, _ := oldColor.RGBA()

			normR := float64(r) / 65535.0
			normG := float64(g) / 65535.0
			newr := EncodeFloatRG(normR)
			newg := EncodeFloatRG(normG)

			rrDe := DeNormalize(newr.x)
			rgDe := DeNormalize(newr.y)
			grDe := DeNormalize(newg.x)
			ggDe := DeNormalize(newg.y)

			if uint8(DeNormalize(normR))-uint8(rrDe) >= 10 {
				fmt.Printf("original: %v\n", normR)
				fmt.Printf("original as int: %v\n", uint8(DeNormalize(normR)))
				fmt.Printf("new: %v\n", newr.x)
				fmt.Printf("new as int: %v\n", uint8(DeNormalize(newr.x)))
			}

			// Set new color to current pixel
			newcolor := color.NRGBA{uint8(rrDe), uint8(rgDe), uint8(grDe), uint8(ggDe)}
			newImage.Set(x, y, newcolor)
		}
	}

	// Image is ready now and sits in RAM. Encode the image to the output file.
	outfile, err := os.Create(outfilename)
	check(err)
	defer outfile.Close()
	png.Encode(outfile, newImage)
}

var inFile string
var outFile string

func init() {
	flag.StringVar(&inFile, "source", "source.png", "Please enter the source image name")
	flag.StringVar(&outFile, "result", "result.png", "Please enter the result image name")
	flag.Parse()
}

func main() {
	encode(inFile, outFile)
}

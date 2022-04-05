package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

type Row = []int
type Template = []Row

const BG = 0b0001
const VISOR = 0b0010
const BODY = 0b0100
const ANY = 0b1000

var baseTemplates = [...]Template{
	{
		{BG, BODY, BODY, BODY},
		{BODY, BODY, VISOR, VISOR},
		{BODY, BODY, BODY, BODY},
		{BG, BODY, BODY, BODY},
		{ANY, BODY, BG, BODY},
		{ANY, BODY, BG, BODY},
	},
	{
		{BG, BODY, BODY, BODY},
		{BODY, BODY, VISOR, VISOR},
		{BODY, BODY, BODY, BODY},
		{BG, BODY, BODY, BODY},
		{ANY, BODY, BG, BODY},
	},
	{
		{BG, BODY, BODY, BODY},
		{BODY, BODY, VISOR, VISOR},
		{BG, BODY, BODY, BODY},
		{BG, BODY, BODY, BODY},
		{ANY, BODY, BG, BODY},
	},
	{
		{BG, BODY, BODY, BODY},
		{BODY, BODY, VISOR, VISOR},
		{BG, BODY, BODY, BODY},
		{ANY, BODY, BG, BODY},
	},
	{
		{BG, BODY, BODY, BODY},
		{BODY, BODY, VISOR, VISOR},
		{BODY, BODY, BODY, BODY},
		{BG, BODY, BG, BODY},
	},
	{
		{BG, BODY, BODY, BODY},
		{BG, BODY, VISOR, VISOR},
		{BODY, BODY, BODY, BODY},
		{BG, BODY, BODY, BODY},
		{ANY, BODY, BG, BODY},
	},
	{
		{BG, BODY, BODY, BODY},
		{BODY, BODY, VISOR, VISOR},
		{BODY, BODY, BODY, BODY},
		{BG | BODY, BODY, BODY, BODY},
		{BG, BODY, BG, BODY},
	},
	{
		{BODY, BODY, BODY},
		{BODY, VISOR, VISOR},
		{BODY, BODY, BODY},
		{BODY, BODY, BODY},
		{BODY, BG, BODY},
	},
	{
		{BODY, BODY, BODY},
		{BODY, VISOR, VISOR},
		{BODY, BODY, BODY},
		{BODY, BG, BODY},
	},
}

func delta(a uint8, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

func diff(c1 color.Color, c2 color.Color) uint8 {
	cc1 := color.NRGBAModel.Convert(c1).(color.NRGBA)
	cc2 := color.NRGBAModel.Convert(c2).(color.NRGBA)
	r := delta(cc1.R, cc2.R)
	g := delta(cc1.G, cc2.G)
	b := delta(cc1.B, cc2.B)
	a := delta(cc1.A, cc2.A)
	return r + g + b + a
}

func isSame(c1 color.Color, c2 color.Color, delta uint8) bool {
	return diff(c1, c2) < delta
}

func bodyPoint(template Template) (int, int) {
	for dy, row := range template {
		for dx, t := range row {
			if t == BODY {
				return dx, dy
			}
		}
	}
	panic("no body color in template")
}

func visorPoint(template Template) (int, int) {
	for dy, row := range template {
		for dx, t := range row {
			if t == VISOR {
				return dx, dy
			}
		}
	}
	panic("no visor color in template")
}

func reverse(numbers []int) []int {
	newNumbers := make([]int, len(numbers))
	for i, j := 0, len(numbers)-1; i <= j; i, j = i+1, j-1 {
		newNumbers[i], newNumbers[j] = numbers[j], numbers[i]
	}
	return newNumbers
}

func flipHorizontal(template Template) Template {
	var flipped Template
	for _, row := range template {
		flipped = append(flipped, reverse(row))
	}
	return flipped
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s [filename]\n", os.Args[0])
		return
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Printf("Failed to read file %s\n", os.Args[0])
		fmt.Println(err)
		return
	}
	img, err := png.Decode(f)
	if err != nil {
		fmt.Println("Failed to decode PNG")
		fmt.Println(err)
		return
	}
	bounds := img.Bounds()
	fmt.Printf("Bounds: %d %d\n", bounds.Dx(), bounds.Dy())

	outImg := image.NewNRGBA(image.Rectangle{
		Min: bounds.Min,
		Max: bounds.Max,
	})
	out, err := os.Create("output.png")
	var templates []Template
	for _, template := range baseTemplates {
		templates = append(templates, template)
		templates = append(templates, flipHorizontal(template))
	}
	count := 0
	for x := 0; x < bounds.Dx(); x++ {
		for y := 0; y < bounds.Dy(); y++ {
			var matched Template = nil
			for _, template := range templates {
				bodyX, bodyY := bodyPoint(template)
				visorX, visorY := visorPoint(template)
				bodyColor := img.At(x+bodyX, y+bodyY)
				visorColor := img.At(x+visorX, y+visorY)
				if isSame(bodyColor, visorColor, 16) {
					continue
				}
				if outImg.At(x, y).(color.NRGBA).A != 0 {
					continue
				}
				match := true
				for dy, row := range template {
					for dx, t := range row {
						pixel := img.At(x+dx, y+dy)
						if t&BODY != 0 && isSame(pixel, bodyColor, 8) {
							continue
						}
						if t&VISOR != 0 && isSame(pixel, visorColor, 8) {
							continue
						}
						if t&BG != 0 && !isSame(pixel, bodyColor, 6) {
							continue
						}
						if t == ANY {
							continue
						}

						match = false
						break
					}
					if !match {
						break
					}
				}
				if match {
					matched = template
					break
				}
			}
			originalColor := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			dimmedColor := color.RGBA{
				R: originalColor.R / 4,
				G: originalColor.G / 4,
				B: originalColor.B / 4,
				A: originalColor.A,
			}
			if outImg.At(x, y).(color.NRGBA).A == 0 {
				outImg.Set(x, y, dimmedColor)
			}
			if matched == nil {
				continue
			}
			count++
			for dy, row := range matched {
				for dx, t := range row {
					if t&(BODY|VISOR) != 0 {
						outImg.Set(x+dx, y+dy, img.At(x+dx, y+dy))
					}
				}
			}
		}
	}
	err = png.Encode(out, outImg)
	if err != nil {
		fmt.Println("Failed to encode png")
		fmt.Println(err)
		return
	}
	fmt.Println("Found", count)

}

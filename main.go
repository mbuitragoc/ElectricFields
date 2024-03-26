package main

import (
	"bufio"
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

var (
	MinValue = 0.0
	MaxValue = 20.0
)

type Point struct {
	X, Y   float64
	Charge float64
}

type TestPoint struct {
	X, Y float64
}

func CalculateElectricField(charge, pointX, pointY, testPointX, testPointY float64) (float64, float64) {
	k := 8.987551787e9
	dx := 0.0
	dy := 0.0
	if charge > 0 {
		dx = testPointX - pointX
		dy = testPointY - pointY
	} else {
		dx = pointX - testPointX
		dy = pointY - testPointY
	}
	r := math.Sqrt(dx*dx + dy*dy)
	if r == 0 {
		return 0, 0
	}
	direction := math.Atan2(dy, dx)

	electricField := math.Abs(k * charge / (r * r))

	ex := electricField * math.Cos(direction)
	ey := electricField * math.Sin(direction)
	return ex, ey
}

func GenerateTestPoints(interval float64, points []Point) []TestPoint {
	var testPoints []TestPoint
	existingPoints := make(map[float64]map[float64]bool)

	for _, pt := range points {
		if _, ok := existingPoints[pt.X]; !ok {
			existingPoints[pt.X] = make(map[float64]bool)
		}
		existingPoints[pt.X][pt.Y] = true
	}

	for x := MinValue; x <= MaxValue; x += interval {
		for y := MinValue; y <= MaxValue; y += interval {
			if _, exists := existingPoints[x][y]; !exists {
				testPoints = append(testPoints, TestPoint{X: x, Y: y})
			}
		}
	}

	return testPoints
}

func PlotPoints(p *plot.Plot, points []Point, testPoints []TestPoint) {
	for _, pt := range points {
		var c color.Color
		if pt.Charge > 0 {
			c = color.RGBA{R: 255, G: 0, B: 0, A: 255}
		} else {
			c = color.RGBA{R: 0, G: 0, B: 255, A: 255}
		}
		scatter, err := plotter.NewScatter(plotter.XYs{{X: pt.X, Y: pt.Y}})
		if err != nil {
			log.Fatalf("could not create scatter plot: %v", err)
		}
		scatter.GlyphStyle.Color = c
		scatter.GlyphStyle.Radius = vg.Points(5)
		p.Add(scatter)
	}

	for _, pt := range testPoints {
		scatter, err := plotter.NewScatter(plotter.XYs{{X: pt.X, Y: pt.Y}})
		if err != nil {
			log.Fatalf("could not create scatter plot: %v", err)
		}
		scatter.GlyphStyle.Color = color.RGBA{0, 0, 0, 255}
		scatter.GlyphStyle.Radius = vg.Points(2)
		p.Add(scatter)
	}
}

func PlotElectricFieldLines(p *plot.Plot, points []Point, testPoints []TestPoint) {
	for _, testPoint := range testPoints {
		sumEx := 0.0
		sumEy := 0.0
		for _, point := range points {
			ex, ey := CalculateElectricField(point.Charge, point.X, point.Y, testPoint.X, testPoint.Y)
			sumEx += ex
			sumEy += ey
		}

		sumNorm := math.Sqrt(sumEx*sumEx + sumEy*sumEy)

		var lineColor color.Color
		if sumNorm == 0 {
			lineColor = color.RGBA{R: 0, G: 0, B: 0, A: 0}
		} else {
			intensity := uint8(math.Min(255, 255*sumNorm/8.987551787e+07))
			lineColor = color.RGBA{R: 0, G: intensity, B: 255 - intensity, A: 255}
		}

		sumEx /= sumNorm
		sumEy /= sumNorm

		line, err := plotter.NewLine(plotter.XYs{{X: testPoint.X, Y: testPoint.Y}, {X: testPoint.X + sumEx, Y: testPoint.Y + sumEy}})
		if err != nil {
			log.Fatalf("could not create line plot: %v", err)
		}
		// line.LineStyle.Color = color.RGBA{R: 0, G: 0, B: 0, A: 255}
		line.LineStyle.Color = lineColor
		p.Add(line)
	}
}

func main() {
	fmt.Print("Enter the number of charges:")
	reader := bufio.NewReader(os.Stdin)
	numChargesStr, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("error reading input: %v", err)
	}
	numChargesStr = strings.TrimSpace(numChargesStr)
	numCharges, err := strconv.Atoi(numChargesStr)
	if err != nil {
		log.Fatalf("invalid number of charges: %v", err)
	}

	var charges []Point

	for i := 0; i < numCharges; i++ {
		fmt.Printf("Enter the position (x y) and charge value for charge %d, separated by spaces: ", i+1)
		inputStr, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("error reading input: %v", err)
		}
		inputStr = strings.TrimSpace(inputStr)
		parts := strings.Fields(inputStr)
		if len(parts) != 3 {
			log.Fatalf("invalid input format")
		}
		x, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			log.Fatalf("invalid position x: %v", err)
		}
		y, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			log.Fatalf("invalid position y: %v", err)
		}
		charge, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			log.Fatalf("invalid charge value: %v", err)
		}
		charges = append(charges, Point{X: x, Y: y, Charge: charge})
	}

	fmt.Print("Enter the maximum value for the plot: ")
	maxValueStr, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("error reading input: %v", err)
	}
	maxValueStr = strings.TrimSpace(maxValueStr)
	maxValue, err := strconv.ParseFloat(maxValueStr, 64)
	if err != nil {
		log.Fatalf("invalid maximum value: %v", err)
	}
	MaxValue = maxValue

	fmt.Print("Enter the filename for the final plot (e.g., my_plot.png): ")
	filename, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("error reading input: %v", err)
	}
	filename = strings.TrimSpace(filename)

	testPoints := GenerateTestPoints(1, charges)

	p := plot.New()
	p.Title.Text = "Electric Field Vectors"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"

	PlotPoints(p, charges, testPoints)

	PlotElectricFieldLines(p, charges, testPoints)

	p.X.Min = MinValue
	p.X.Max = MaxValue
	p.Y.Min = MinValue
	p.Y.Max = MaxValue

	if err := p.Save(8*vg.Inch, 8*vg.Inch, filename); err != nil {
		log.Fatalf("could not save plot: %v", err)
	}
}

package rexprint

import (
	"fmt"
	"math/rand"
)

func printAllColors() {
	colors := GetRandomColors()

	for i, c := range colors {
		fmt.Println(fmt.Sprintf("\x1b[%dm%s\x1b[0m", c, fmt.Sprintf("[%d]:%d", i, c)))
	}
}

func GetRandomColors() []int {
	colors := []int{
		31,
		32,
		33,
		34,
		35,
		36,
		//37, Barely visible
		//90, Barely visible
		//91, Very similar to red (31)
		92,
		93,
		94,
		95,
		96,
		97,
	}

	// Shuffle colors array
	//rand.Seed(time.Now().UnixNano())
	// Prefer a constant color distribution for easier debugging
	rand.Seed(3)
	for i := range colors {
		j := rand.Intn(i + 1)
		colors[i], colors[j] = colors[j], colors[i]
	}
	return colors
}

func PrintLnColor(id string, colors []int, i int, msg ...any) {
	colorIndex := i % len(colors)
	colored := fmt.Sprintf("\x1b[%dm%s\x1b[0m", colors[colorIndex], fmt.Sprintf("[%s]", id))
	msg = append([]any{colored}, msg...)
	fmt.Println(msg...)
}

func getColored(color int, text string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color, text)
}

func Dim(text string) string {
	return getColored(37, text)
}

func ErrColor(text string) string {
	return getColored(41, text)
}

func OutColor(text string) string {
	return getColored(42, text)
}

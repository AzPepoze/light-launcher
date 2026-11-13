package contract

import (
	"regexp"
	"strconv"
)

// defaultCover is the poster ratio, used when a source names no cover and no tile.
const defaultCover = "3:4"

// coverMaxSide is the largest width or height a cover ratio can use.
const coverMaxSide = 32

var coverPattern = regexp.MustCompile(`^([1-9][0-9]?):([1-9][0-9]?)$`)

// NormalizeCover returns cover when it is set, else the ratio of the older tile name.
func NormalizeCover(cover, tile string) string {
	if cover != "" {
		return cover
	}
	return coverOfTile(tile)
}

// coverOfTile maps the older tile names to a cover ratio. Any other tile gets the poster ratio.
func coverOfTile(tile string) string {
	switch tile {
	case "wide":
		return "16:9"
	case "square":
		return "1:1"
	default:
		return defaultCover
	}
}

// ValidCover reports whether cover is "W:H" with sides of 1 to 32 and a width to height ratio from 0.25 to 4.
func ValidCover(cover string) bool {
	width, height, ok := coverSides(cover)
	return ok && width <= coverMaxSide && height <= coverMaxSide && width <= 4*height && height <= 4*width
}

// coverSides reads the two sides of a "W:H" string.
func coverSides(cover string) (width, height int, ok bool) {
	match := coverPattern.FindStringSubmatch(cover)
	if match == nil {
		return 0, 0, false
	}
	width, errWidth := strconv.Atoi(match[1])
	height, errHeight := strconv.Atoi(match[2])
	return width, height, errWidth == nil && errHeight == nil
}

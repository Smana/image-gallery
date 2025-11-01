package settings

import (
	"fmt"
	"hash/fnv"
)

// TagColorPalette is a curated set of accessible, distinguishable colors
// chosen from TailwindCSS color palette with WCAG AA contrast compliance
var TagColorPalette = []string{
	"#3B82F6", // blue-500
	"#10B981", // green-500
	"#F59E0B", // amber-500
	"#EF4444", // red-500
	"#8B5CF6", // violet-500
	"#EC4899", // pink-500
	"#06B6D4", // cyan-500
	"#F97316", // orange-500
	"#14B8A6", // teal-500
	"#A855F7", // purple-500
	"#6366F1", // indigo-500
	"#10B981", // emerald-500
}

// TagColorPaletteClasses maps hex colors to TailwindCSS background classes
var TagColorPaletteClasses = map[string]string{
	"#3B82F6": "bg-blue-500 text-white",
	"#10B981": "bg-green-500 text-white",
	"#F59E0B": "bg-amber-500 text-white",
	"#EF4444": "bg-red-500 text-white",
	"#8B5CF6": "bg-violet-500 text-white",
	"#EC4899": "bg-pink-500 text-white",
	"#06B6D4": "bg-cyan-500 text-white",
	"#F97316": "bg-orange-500 text-white",
	"#14B8A6": "bg-teal-500 text-white",
	"#A855F7": "bg-purple-500 text-white",
	"#6366F1": "bg-indigo-500 text-white",
}

// GetTagColor returns a consistent color for a given tag name using FNV-1a hashing
// The same tag name will always produce the same color across all instances
func GetTagColor(tagName string) string {
	if tagName == "" {
		return TagColorPalette[0] // Default to blue for empty tags
	}

	// Use FNV-1a hash for fast, consistent hashing
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(tagName))
	hashValue := hasher.Sum32()

	// Map hash to palette index
	paletteIndex := int(hashValue) % len(TagColorPalette)
	return TagColorPalette[paletteIndex]
}

// GetTagColorClass returns TailwindCSS classes for a tag based on its color
func GetTagColorClass(tagName string) string {
	color := GetTagColor(tagName)
	if class, exists := TagColorPaletteClasses[color]; exists {
		return class
	}
	// Fallback to blue if color not in class map
	return "bg-blue-500 text-white"
}

// GetTagStyle returns inline style with background color for a tag
// Useful for HTML rendering where Tailwind classes aren't available
func GetTagStyle(tagName string) string {
	color := GetTagColor(tagName)
	return fmt.Sprintf("background-color: %s; color: white;", color)
}

// GetLightTagColorClass returns a lighter version of the tag color for non-selected states
// Uses Tailwind's 100-weight colors for backgrounds with darker text
var LightTagColorClasses = map[string]string{
	"#3B82F6": "bg-blue-100 text-blue-800",
	"#10B981": "bg-green-100 text-green-800",
	"#F59E0B": "bg-amber-100 text-amber-800",
	"#EF4444": "bg-red-100 text-red-800",
	"#8B5CF6": "bg-violet-100 text-violet-800",
	"#EC4899": "bg-pink-100 text-pink-800",
	"#06B6D4": "bg-cyan-100 text-cyan-800",
	"#F97316": "bg-orange-100 text-orange-800",
	"#14B8A6": "bg-teal-100 text-teal-800",
	"#A855F7": "bg-purple-100 text-purple-800",
	"#6366F1": "bg-indigo-100 text-indigo-800",
}

// GetLightTagColorClass returns light background classes for tag display
func GetLightTagColorClass(tagName string) string {
	color := GetTagColor(tagName)
	if class, exists := LightTagColorClasses[color]; exists {
		return class
	}
	return "bg-gray-100 text-gray-800"
}

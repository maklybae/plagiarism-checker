package textstat

import (
	"bytes"
	"unicode"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func CountStats(data []byte) (lines, words, chars int) {
	lines = bytes.Count(data, []byte{'\n'})
	chars = len([]rune(string(data)))
	inWord := false

	for _, r := range string(data) {
		if unicode.IsSpace(r) {
			inWord = false
		} else if !inWord {
			words++
			inWord = true
		}
	}

	return lines, words, chars
}

// Similarity возвращает метрику схожести двух текстов в диапазоне [0,1], где 1 — полное совпадение.
func Similarity(a, b string) float64 {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(a, b, false)
	dist := float64(dmp.DiffLevenshtein(diffs))

	lenA := float64(len([]rune(a)))
	lenB := float64(len([]rune(b)))
	maxLen := max(lenB, lenA)

	if maxLen == 0 {
		return 1 // оба текста пустые
	}

	return 1 - dist/maxLen
}

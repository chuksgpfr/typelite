package tokenizer

import "strings"

func Tokenize(word string) []string {
	stopWord := " "
	tokens := strings.Split(word, stopWord)

	return tokens
}

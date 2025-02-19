package shared

import (
	"log"
	"regexp"
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

// Function to count tokens using tiktoken
func NumTokensFromText(text string) int {
	encoder, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		log.Fatalf("Failed to get encoding: %v", err)
	}
	return len(encoder.Encode(text, nil, nil))
}

type RecursiveCharacterTextSplitter struct {
	separators       []string
	chunkSize        int
	chunkOverlap     int
	keepSeparator    bool
	isSeparatorRegex bool
}

func NewRecursiveCharacterTextSplitter(chunkSize int, chunkOverlap int, separators []string) *RecursiveCharacterTextSplitter {
	return &RecursiveCharacterTextSplitter{
		chunkSize:     chunkSize,
		chunkOverlap:  chunkOverlap,
		separators:    separators,
		keepSeparator: true,
	}
}

func (r *RecursiveCharacterTextSplitter) SplitText(text string) []string {
	return r.splitTextRecursive(text, r.separators)
}

func (r *RecursiveCharacterTextSplitter) splitTextRecursive(text string, separators []string) []string {
	var finalChunks []string
	separator := separators[len(separators)-1]
	var newSeparators []string

	for i, s := range separators {
		if s == "" {
			separator = s
			break
		}
		if matched, _ := regexp.MatchString(regexp.QuoteMeta(s), text); matched {
			separator = s
			newSeparators = separators[i+1:]
			break
		}
	}

	splits := strings.Split(text, separator)
	goodSplits := []string{}

	for _, s := range splits {
		if len(s) < r.chunkSize {
			goodSplits = append(goodSplits, s)
		} else {
			if len(goodSplits) > 0 {
				finalChunks = append(finalChunks, strings.Join(goodSplits, separator))
				goodSplits = []string{}
			}
			if len(newSeparators) == 0 {
				finalChunks = append(finalChunks, s)
			} else {
				finalChunks = append(finalChunks, r.splitTextRecursive(s, newSeparators)...)
			}
		}
	}

	if len(goodSplits) > 0 {
		finalChunks = append(finalChunks, strings.Join(goodSplits, separator))
	}

	return finalChunks
}

func SplitText(text string, maxTokens int) []string {
	maxTokens = maxTokens - (10 * maxTokens / 100) // Reduce by 10% for overlap

	separators := []string{"\n\n", "\n", " ", ""}
	splitter := NewRecursiveCharacterTextSplitter(maxTokens, maxTokens/10, separators)
	return splitter.SplitText(text)
}

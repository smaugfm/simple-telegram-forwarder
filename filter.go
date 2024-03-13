package main

import (
	"fmt"
	tdlib "github.com/zelenin/go-tdlib/client"
	"regexp"
)

type RegexFilter struct {
	regex *regexp.Regexp
}

func (r *RegexFilter) Describe() string {
	return fmt.Sprintf("<RegexFilter %s>", r.regex)
}

func (r *RegexFilter) Passes(msg *tdlib.Message) bool {
	text := getTextFromMessage(msg)
	return r.regex.MatchString(text)
}

type EmptyFilter struct {
}

func (e EmptyFilter) Passes(_ *tdlib.Message) bool {
	return true
}

func (e EmptyFilter) Describe() string {
	return "always passing empty filter"
}

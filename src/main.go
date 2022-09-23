// Package main provides ...
package main

import (
	"fmt"
	"sync"
	"time"
	"os/exec"
)

type Selection struct {
	text string
	index int64
}

type TranslatedText struct {
	text string
	priority int64
	index int64
}

type PQ []TranslatedText

func Command(cmd string) ([]byte, error) {
	c := exec.Command("bash", "-c", cmd)
	output, err := c.CombinedOutput()
	return output, err
}

func handleSelected(sel []byte) string {
	text := string(sel)
	return text
}

// 获取选择的内容并返回
func getSel(currentIndex *int64) Selection {
	sel, err := Command("xsel")
	if err != nil {
		fmt.Println(err.Error())
		return Selection{"", -1}
	}

	selectedText := handleSelected(sel)

	selection := Selection{ selectedText, *currentIndex }
	*currentIndex += 1
	return selection
}

func compare(a Selection, b Selection) int {
	if a.text < b.text {
		return -1
	} else if a.text > b.text {
		return 1
	} else {
		return 0
	}
}

func translate(sel Selection, nextIndex *int64) TranslatedText {
	translation := TranslatedText{sel.text, 0, *nextIndex}
	*nextIndex += 1
	return translation
}

func push(pq *PQ, translatedText TranslatedText) {
	*pq = append(*pq, translatedText)
}

func pop(pq *PQ) TranslatedText {
	var selectedText TranslatedText
	var selectedIndex int = 0
	selectedText = (*pq)[selectedIndex]
	for i := 1; i < len(*pq); i++ {
		if selectedText.priority > (*pq)[i].priority {
			selectedText = (*pq)[i]
			selectedIndex = i
		}
	}

	*pq = append((*pq)[:selectedIndex], (*pq)[selectedIndex+1:]...)

	return selectedText
}

/**
 * 打印翻译后的文本
 */
func printText(translatedText TranslatedText) {
	fmt.Println(fmt.Sprint(translatedText.index) + ">>> " + translatedText.text)
}

func main() {
	var wg sync.WaitGroup
	var sid int64 = 0
	var tid int64 = 0
	var dt float64 = 0.3 // 秒
	var preSel Selection = Selection{ "", 0 }
	var q PQ

	for {
		var sel Selection = getSel(&sid)
		var diff int = compare(sel, preSel)
		preSel = sel
		if diff != 0 {
			wg.Add(1)
			go func() {
				var translatedText TranslatedText = translate(sel, &tid)
				push(&q, translatedText)
				var top TranslatedText = pop(&q)
				printText(top)
				wg.Done()
			}()
		}
		time.Sleep(time.Duration( dt * float64(time.Second) ))
	}

	wg.Wait()
}


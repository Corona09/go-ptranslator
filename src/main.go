// Package main provides ...
package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
	"sync"
	"time"
	"io"
)

type Selection struct {
	text string
	index int64
}

type TranslatedText struct {
	srcText string
	destText string
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
	translation := TranslatedText{sel.text, sel.text, 0, *nextIndex}
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

func httpGet(url string) string {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}
	return result.String()
}

func google_translate_longstring(srcLang string, targetLang string, text string) string {
	url := fmt.Sprintf(
		"https://translate.googleapis.com/translate_a/single?client=gtx&sl=%s&tl=%s&dt=t&q=%s",
		srcLang,
		targetLang,
		text)
	result := httpGet(url)
	return result
}

/**
 * 打印翻译后的文本
 */
func printText(translatedText TranslatedText) {
	fmt.Println("原文: " + translatedText.srcText)
	translatedText.destText = google_translate_longstring("en", "zh-CN", translatedText.srcText)
	fmt.Println(fmt.Sprint(translatedText.index) + " >>> " + translatedText.destText)
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
			var translatedText TranslatedText = translate(sel, &tid)
			push(&q, translatedText)
			var top TranslatedText = pop(&q)
			printText(top)
		}
		time.Sleep(time.Duration( dt * float64(time.Second) ))
	}

	wg.Wait()
}


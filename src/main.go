package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/tidwall/gjson"
)

type Selection struct {
	text string
	index int64
}

type TranslatedText struct {
	srcText string
	destText string
	explanationCN []string
	explanationWeb []string
	webPhrase map[string][]string
	pronounce []string
	priority int64
	index int64
}

/**
 * 翻译文本
 * @param sel 待翻译的文本
 * @param srcLang 源语言
 * @param destLang 目标语言
 * @param nextIndex 下一个翻译后的文本的 id
 * @return 翻译之后的文本
 */
func translate(sel Selection, srcLang string, destLang string, nextIndex *int64) TranslatedText {
	tr := TranslatedText{
		srcText: sel.text,
		destText: sel.text, 
		explanationCN: nil,
		explanationWeb: nil,
		webPhrase: nil,
		priority: 0,
		index: *nextIndex,
	}
	*nextIndex += 1

	n := strings.Count(tr.srcText, " ")
	if n == 0 && len(tr.srcText) < 30 {
		// tr.destText = google_translate_shortword(srcLang, destLang, tr.srcText)
		google_translate_shortword(tr.srcText, &tr)
	} else {
		tr.destText = google_translate_longstring(srcLang, destLang, tr.srcText)
	}

	return tr
}

/**
 * 使用谷歌翻译短的单词
 * @param srcLang 源语言 (en)
 * @param targetLang 目标语言 (zh-CN)
 * @param text 待翻译文本
 * @return 翻译后的文本
 */
func google_translate_shortword(text string, tr *TranslatedText) {
	req := fmt.Sprintf("http://dict.youdao.com/w/%s", text)
	// resp := HttpGet(u)
	resp, err := http.Get(req)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	tr.srcText = text

	tr.explanationCN = getExplanationCN(*doc)
	tr.explanationWeb = getExplanationWeb(*doc)
	tr.webPhrase = getWebPhrase(*doc)
	tr.pronounce = getPronounce(*doc)
}

/**
 * 使用谷歌翻译长文本
 * @param srcLang 源语言 (en)
 * @param targetLang 目标语言 (zh-CN)
 * @param text 待翻译文本
 * @return 翻译后的文本
 */
func google_translate_longstring(srcLang string, targetLang string, text string) string {
	u := fmt.Sprintf(
		"https://translate.googleapis.com/translate_a/single?client=gtx&sl=%s&tl=%s&dt=t&q=%s",
		srcLang,
		targetLang,
		text)
	// [[["翻译","translate",null,null,10]],null,"en",null,null,null,null,[]]
	resp := HttpGet(u)
	result := gjson.Parse(resp).Get("0").Get("0").Get("0").String()
	return result
}

// 打印翻译后的文本
// @param tr 翻译后的文本
func printText(tr TranslatedText) {
	cyan := color.New(color.FgCyan).Add(color.Bold)
	white := color.New(color.FgWhite).Add(color.Bold)
	cyan.Printf("* * * "); white.Printf("<%02d> %s", tr.index, time.Now().String()[:19]); cyan.Printf(" * * *\n")
	
	bold := color.New(color.Bold)
	greenBold := color.New(color.FgHiGreen).Add(color.Bold)

	prompt := color.New(color.FgHiYellow).Add(color.Bold)
	prompt.Printf(" [原文]"); greenBold.Printf(" >>> "); bold.Println(tr.srcText + " (" + fmt.Sprint(len(tr.srcText)) + ")")
	if strings.Count(tr.srcText, " ") >= 1 {
		prompt.Printf(" [翻译]"); greenBold.Printf(" >>> "); white.Println(tr.destText)
		fmt.Println()
	} else {
		printYoudaoTrans(tr)
	}
}

func welcome() {
	bdu := color.New(color.Bold)
	bdu.Printf("**************************************\n")
	bdu.Printf("*     Welcome to use GTranslator!    *\n")
	bdu.Printf("*   Copyright(C)Corona 2022, GPL v3  *\n")
	bdu.Printf("**************************************\n")
	fmt.Println()
}

func main() {
	var sid int64 = 0
	var tid int64 = 1
	var dt float64 = 0.3 // 秒
	var preSel Selection = Selection{ "", 0 }
	var q PQ

	const MAX_TEXT_LENGTH int = 255
	
	var srcLang string = "en"
	var destLang string = "zh-CN"

	ClearSel()

	welcome()

	for {
		var sel Selection = GetSel(&sid)
		var diff int = Compare(sel, preSel)
		
		if len(sel.text) > MAX_TEXT_LENGTH {
			// 文本过长, 输出错误提示
			fmt.Printf("Its too long\n\n")
			ClearSel()
			continue
		}

		preSel = sel
		if diff != 0 {
			var translatedText TranslatedText = translate(sel, srcLang, destLang, &tid)
			push(&q, translatedText)
			var top TranslatedText = pop(&q)
			printText(top)
		}
		time.Sleep(time.Duration( dt * float64(time.Second) ))
	}
}


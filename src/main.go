package main

import (
	"fmt"
	"strings"
	"time"
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
	tr := TranslatedText{sel.text, sel.text, 0, *nextIndex}
	*nextIndex += 1

	n := strings.Count(tr.srcText, " ")
	if n == 0 && len(tr.srcText) < 30 {
		tr.destText = google_translate_shortword(srcLang, destLang, tr.srcText)
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
func google_translate_shortword(srcLang string, targetLang string, text string) string {
	u := fmt.Sprintf(
		"https://translate.googleapis.com/translate_a/single?client=gtx&sl=%s&tl=%s&dj=1&dt=t&dt=bd&dt=qc&dt=rm&dt=ex&dt=at&dt=ss&dt=rw&dt=ld&q=%s&button&tk=233819.233819",
		srcLang,
		targetLang,
		text,
	)
	// {"sentences":[{"trans":"这","orig":"The","backend":10},{"translit":"Zhè"}],"src":"en","alternative_translations":[{"src_phrase":"The","alternative":[{"word_postproc":"这","score":1000,"has_preceding_space":true,"attach_to_next_token":false,"backends":[10]},{"word_postproc":"该","score":0,"has_preceding_space":true,"attach_to_next_token":false,"backends":[3],"backend_infos":[{"backend":3}]},{"word_postproc":"那个","score":0,"has_preceding_space":true,"attach_to_next_token":false,"backends":[8]}],"srcunicodeoffsets":[{"begin":0,"end":3}],"raw_src_segment":"The","start_pos":0,"end_pos":0}],"confidence":1.0,"spell":{},"ld_result":{"srclangs":["en"],"srclangs_confidences":[1.0],"extended_srclangs":["en"]}}
	resp := HttpGet(u)
	sentences := gjson.Parse(resp).Get("sentences")
	trans := sentences.Get("0").Get("trans").String()
	return trans
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

/**
 * 打印翻译后的文本
 */
func printText(translatedText TranslatedText) {
	color.Cyan("* * * <%02d> %s * * *", translatedText.index, time.Now().String()[:19])
	bold := color.New(color.Bold)
	greenBold := color.New(color.FgGreen).Add(color.Bold)
	blueBold := color.New(color.FgBlue).Add(color.Bold)
	bold.Printf("- Original Text"); greenBold.Printf(" >>> "); bold.Println(translatedText.srcText )
	bold.Printf("- Translation  ");  blueBold.Printf(" >>> "); bold.Println(translatedText.destText)
	fmt.Println()
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


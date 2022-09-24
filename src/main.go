// Package main provides ...
package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

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

type PQ []TranslatedText

/**
 * 执行 bash 命令
 * @param cmd 要执行的命令
 * @return (命令的执行输出, 报错)
 */
func Command(cmd string) ([]byte, error) {
	c := exec.Command("bash", "-c", cmd)
	output, err := c.CombinedOutput()
	return output, err
}

/**
 * 处理选择的文本
 * @param sel 被选择的文本
 * @return 处理后的文本
 */
func handleSelected(sel []byte) string {
	text := string(sel)
	text = strings.Trim(text, " ")
	text = strings.Trim(text, "\t")
	text = strings.Replace(text, "\n", " ", -1)
	text = strings.Trim(text, "\n")
	return text
}

/**
 * 获取选择的内容并返回
 * @param currentIndex 下一个被选择文本的 id
 * @param 被选择的文本
 */
func getSel(nextIndex *int64) Selection {
	sel, err := Command("xsel")
	if err != nil {
		fmt.Println(err.Error())
		return Selection{"", -1}
	}

	selectedText := handleSelected(sel)

	selection := Selection{ selectedText, *nextIndex }
	*nextIndex += 1
	return selection
}

/**
 * 比较两个选择
 * @return 相同返回 0, 不同返回非 0
 */
func compare(a Selection, b Selection) int {
	if a.text < b.text {
		return -1
	} else if a.text > b.text {
		return 1
	} else {
		return 0
	}
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
 * 将翻译后的文本压入队列中
 */
func push(pq *PQ, translatedText TranslatedText) {
	*pq = append(*pq, translatedText)
}

/**
 * 取出队列中 index 最小的文本并返回
 */
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
 * 发送 http get 请求
 * @param req 请求路径
 * @return 返回的而结果
 */
func httpGet(req string) string {
	client := &http.Client{Timeout: 5 * time.Second}
	// 需要对 req 路径进行转义
	u, _ := url.Parse(req)
	q := u.Query()
	u.RawQuery = q.Encode()
	resp, err := client.Get(u.String())
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
	resp := httpGet(u)
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
	resp := httpGet(u)
	result := gjson.Parse(resp).Get("0").Get("0").Get("0").String()
	return result
}

/**
 * 打印翻译后的文本
 */
func printText(translatedText TranslatedText) {
	fmt.Println("原文: " + translatedText.srcText)
	// translatedText.destText = google_translate_longstring("en", "zh-CN", translatedText.srcText)
	fmt.Println(fmt.Sprint(translatedText.index) + " >>> " + translatedText.destText)
}

func main() {
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
			var translatedText TranslatedText = translate(sel, "en", "zh-CN", &tid)
			push(&q, translatedText)
			var top TranslatedText = pop(&q)
			printText(top)
		}
		time.Sleep(time.Duration( dt * float64(time.Second) ))
	}
}


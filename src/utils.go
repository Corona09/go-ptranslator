// Package main provides ...
package main

import (
	"fmt"
	"strings"
	"os/exec"
	"net/http"
	"net/url"
	"time"
	"bytes"
	"io"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
)

// 清除选择
func ClearSel() {
	Command("xsel -c")
}

// 获取选择的内容并返回
// @param currentIndex 下一个被选择文本的 id
// @param 被选择的文本
func GetSel(nextIndex *int64) Selection {
	sel, err := Command("xsel")
	if err != nil {
		fmt.Println(err.Error())
		return Selection{"", -1}
	}

	selectedText := HandleSelected(sel)

	selection := Selection{ selectedText, *nextIndex }
	*nextIndex += 1
	return selection
}

// 处理选择的文本
// @param sel 被选择的文本
// @return 处理后的文本
func HandleSelected(sel []byte) string {
	text := string(sel)
	text = strings.Trim(text, " ")
	text = strings.Trim(text, "\t")
	text = strings.Trim(text, "\n")
	texts := strings.Split(text, "\n")
	for i:= 0; i < len(texts); i++ {
		texts[i] = strings.Trim(texts[i], " ")
		texts[i] = strings.Trim(texts[i], "\t")
		texts[i] = strings.Trim(texts[i], "\n")
	}
	text = strings.Join(texts, " ")
	text = strings.Replace(text, "- ", "-", -1)
	return text
}

// 比较两个选择
// @return 相同返回 0, 不同返回非 0
func Compare(a Selection, b Selection) int {
	if a.text < b.text {
		return -1
	} else if a.text > b.text {
		return 1
	} else {
		return 0
	}
}

// 执行 bash 命令
// @param cmd 要执行的命令
// @return (命令的执行输出, 报错)
func Command(cmd string) ([]byte, error) {
	c := exec.Command("bash", "-c", cmd)
	output, err := c.CombinedOutput()
	return output, err
}

type PQ []TranslatedText

// 将翻译后的文本压入队列中
func push(pq *PQ, translatedText TranslatedText) {
	*pq = append(*pq, translatedText)
}

// 取出队列中 index 最小的文本并返回
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

// 发送 http get 请求
// @param req 请求路径
// @return 返回的而结果
func HttpGet(req string) string {
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

// 将字符串中连续的多个空白替换为一个
func removeMultipleSpaces(s string) string {
	r := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(r.ReplaceAllString(s, " "))
}

// 获取关键词
func getKeyword(doc goquery.Document) string {
	kw := doc.Find("#results-contents #phrsListTab h2 .keyword").Text()
	return kw
}

// 获取发音
func getPronounce(doc goquery.Document) []string {
	s := doc.Find("#results-contents #phrsListTab h2 div.baav").Text()
	re_en := regexp.MustCompile("英\\s*\\[.*?\\]")
	re_us := regexp.MustCompile("美\\s*\\[.*?\\]")
	var result []string = make([]string, 0)

	f_en := re_en.FindAllString(s, 1)
	f_us := re_us.FindAllString(s, 1)
	if f_en != nil {
		result = append(result, removeMultipleSpaces(f_en[0]))
	}
	if f_us != nil {
		result = append(result, removeMultipleSpaces(f_us[0]))
	}

	return result
}

// 获取中文释义
func getExplanationCN(doc goquery.Document) []string {
	var result []string = make([]string, 0)

	doc.Find("#phrsListTab .trans-container ul li").Each(func(i int, s *goquery.Selection) {
		result = append(result, removeMultipleSpaces(s.Text()))
	})
	return result
}

// 获取网络释义
func getExplanationWeb(doc goquery.Document) []string {
	var result []string = make([]string, 0)
	doc.Find("#tWebTrans div.wt-container .title").Each(func(i int, s *goquery.Selection) {
		result = append(result, removeMultipleSpaces(s.Text()))
	})
	return result
}

// 获取网络短语
func getWebPhrase(doc goquery.Document) map[string][]string {
	var result map[string][]string = make(map[string][]string)

	doc.Find("#webPhrase p.wordGroup").Each(func(i int, s *goquery.Selection) {
		title := s.Find(".contentTitle a.search-js").Text()
		f := strings.TrimSpace(s.Text())
		r := regexp.MustCompile(`\s+`)
		f = r.ReplaceAllString(f, " ")
		f = strings.ReplaceAll(f, title, "")
		result[title] = strings.Split(f, ";")
	})
	return result
}

func printYoudaoTrans(tr TranslatedText) {
	prompt := color.New(color.FgHiYellow).Add(color.Bold)
	yellow := color.New(color.FgHiYellow).Add(color.Bold)
	white := color.New(color.FgWhite)
	green := color.New(color.FgHiGreen).Add(color.Bold)
	cyan := color.New(color.FgHiCyan).Add(color.Bold).Add(color.Italic)
	red := color.New(color.FgHiRed).Add(color.Bold).Add(color.Italic)

	// 打印音标
	prompt.Printf(" [翻译]"); green.Printf(" >>> ");
	white.Printf("%s ", tr.srcText)
	
	var foundTranslation bool = false
	if tr.pronounce != nil && len(tr.pronounce) > 0 {
		foundTranslation = true
		for _, p := range tr.pronounce {
			yellow.Printf("%s ", p)
		}
	}

	fmt.Println()

	// 打印中文释义
	if tr.explanationCN != nil && len(tr.explanationCN) > 0 {
		foundTranslation = true
		for _, exp := range tr.explanationCN {
			exps := strings.Split(exp, "；")
			indient := 0
			for k, subexp := range exps {
				r := regexp.MustCompile("[(a-z)*]\\.(.*)")
				if k == 0 {
					if r.MatchString(subexp) {
						indient = strings.Index(subexp, ".") + 1
						x := strings.SplitN(subexp, ".", 2)
						cyan.Printf(" " + x[0])
						white.Println(" * " + strings.TrimSpace(x[1]))
					} else {
						cyan.Printf(" Phares ")
						white.Println(subexp)
					}
				} else {
					white.Printf(" ")
					for i := 0; i < indient; i++ {
						white.Printf(" ")
					}
					white.Println("* " + subexp)
				}
			}
			fmt.Println()
		}
	}

	// 打印 web 释义
	if tr.explanationWeb != nil && len(tr.explanationWeb) > 0 {
		foundTranslation = true
		cyan.Printf(" Web ")
		for i, exp := range tr.explanationWeb {
			if i == 0 {
				white.Println("* " + exp)
			} else {
				white.Println("     * " + exp)
			}
		}
		cyan.Println()
	}
	// 打印 web 短语
	if tr.webPhrase != nil && len(tr.webPhrase) > 1 {
		foundTranslation = true
		cyan.Println(" Web Phares")
		for key := range tr.webPhrase {
			cyan.Println("  " + key)
			value := tr.webPhrase[key]
			for _, v := range value {
				white.Println("   - " + v)
			}
		}
		white.Println()
	}

	if !foundTranslation {
		prompt.Printf(" [错误] ")
		red.Println("Translation not found")
		red.Println()
	}
}

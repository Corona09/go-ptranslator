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
)

/**
 * 获取选择的内容并返回
 * @param currentIndex 下一个被选择文本的 id
 * @param 被选择的文本
 */
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

/**
 * 处理选择的文本
 * @param sel 被选择的文本
 * @return 处理后的文本
 */
func HandleSelected(sel []byte) string {
	text := string(sel)
	text = strings.Trim(text, " ")
	text = strings.Trim(text, "\t")
	text = strings.Replace(text, "\n", " ", -1)
	text = strings.Trim(text, "\n")
	return text
}

/**
 * 比较两个选择
 * @return 相同返回 0, 不同返回非 0
 */
func Compare(a Selection, b Selection) int {
	if a.text < b.text {
		return -1
	} else if a.text > b.text {
		return 1
	} else {
		return 0
	}
}

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

type PQ []TranslatedText

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

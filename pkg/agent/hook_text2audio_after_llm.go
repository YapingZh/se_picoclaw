package agent

import (
    "context"
    "errors"
    "os/exec"
    "bytes"
    "strings"

    "github.com/sipeed/picoclaw/pkg/logger"
    "github.com/sipeed/picoclaw/pkg/providers"
)

type Text2audioHook struct{}

var globalQueue = make(chan string, 1000)
var globalReqImgBytes string

func (h *Text2audioHook) Name() string { return "text2audio" }

func (h *Text2audioHook) AfterLLM(ctx context.Context, resp *LLMHookResponse) (*LLMHookResponse, HookDecision, error) {
	if h == nil || resp == nil {
		return resp, HookDecision{Action: HookActionContinue}, errors.New("null object")
	}

        logger.WarnCF("text2audioHook", "hook after LLM", map[string]any{
			"name":       "text2audioHook",
			"text":      resp.Response.Content,
	})


	cmd := exec.Command("python3", "/usr/local/text2auto.py", resp.Response.Content)
    	_, err := cmd.CombinedOutput()
    	if err != nil {
		logger.WarnCF("text2audioHook", "hook after LLM", map[string]any{
                        "name":      "text2audioHook",
                        "text":      resp.Response.Content,
						"error":     err,
        	})
    	}

        return resp, HookDecision{Action: HookActionContinue}, nil
}

func SubstringFrom(fullStr, subStr string) (string, bool) {
	// 1. 获取子字符串的索引位置。如果找不到，会返回 -1
	index := strings.Index(fullStr, subStr)

	// 2. 判断是否包含
	if index == -1 {
		return "", false // 不包含，返回空字符串和 false
	}

	// 3. 🎯 核心截取：使用 [index:] 语法直接切片到末尾
	result := fullStr[index:]
	return result, true
}


func CallVLModel(userPrompt string, media_data string){
	desc_img_prompt := userPrompt//"请用中文详细描述这个图片的内容"

	//"请用中文描述这个图片的内容，如果内容设计几何/代数/物理/化学/数学符号，要使用专业数学符号，" +
        //                   "如果内容是人物/动物/植物/生物照片，要提供补充基本信息，让人能了解到这个人物/动物/植物/生物的基本信息"

	defer func() {
		if err_exp := recover(); err_exp != nil {
			logger.WarnCF("text2audioHook", "hook before LLM", map[string]any {
                                                "name":      "text2audioHook",
                                                "text":      "call VL model happen error",
                                                "error":     err_exp,
                	})
		}
	}()

        if len(media_data) > 0 {
		logger.WarnCF("text2audioHook", "hook before LLM", map[string]any {
                                                "name":      "text2audioHook",
                                                "text":      userPrompt,
                })

                cmd := exec.Command("python3", "/usr/local/img2text.py", desc_img_prompt)
                var stdinBuf bytes.Buffer
                stdinBuf.Write([]byte(media_data))
                cmd.Stdin = &stdinBuf

                output_bytes, err := cmd.CombinedOutput()

		output_text := ""
                if output_bytes == nil {
                        output_text = "图片识别失败了哦"
                } else{
                        output_text = string(output_bytes)
                }

		globalQueue <- output_text


		globalReqImgBytes = ""
                logger.WarnCF("text2audioHook", "hook before LLM", map[string]any {
                                                "name":      "text2audioHook",
                                                "text":      output_text,
                                                "error":     err,
                })
        }
}

func (h *Text2audioHook) BeforeLLM(ctx context.Context, req *LLMHookRequest) (*LLMHookRequest, HookDecision, error) {
	if h == nil || req == nil {
		return req, HookDecision{Action: HookActionContinue}, errors.New("null object")
	}

	need_prompt := 1 
	last_msg_id := len(req.Messages) - 1
	if last_msg_id < 0 {
		return req, HookDecision{Action: HookActionContinue}, errors.New("no messages")
	}

	/*logger.WarnCF("text2audioHook", "hook before LLM", map[string]any{
                "name":       "text2audioHook",
		"last message":  req.Messages[last_msg_id].Content,
        })*/

	output_text := "仅仅echo我这段话后面的文字，不要额外做任何事 - "
	if req.Messages[last_msg_id].Content == "[image]" && req.Messages[last_msg_id].Media != nil && req.Messages[last_msg_id].Role == "user" {
		media_len := len(req.Messages[last_msg_id].Media[0])
		if media_len > 0 {
			//dst := make([]byte, media_len)
			///copy(dst, req.Messages[last_msg_id].Media[0])
			//go CallVLModel(last_msg_id, dst)

			globalReqImgBytes = strings.Clone(req.Messages[last_msg_id].Media[0])

			req.Messages[last_msg_id].Media = nil
			req.Messages[last_msg_id].Content = output_text + "*** " + "准备开始图片识别之前，请补充图片识别提示词！提示词必须以：'识别要求' 为开始" + " ***"

			need_prompt = 0
			/*req.Messages = append(req.Messages, providers.Message{
                        	Role: "user",
                                Content: output_text + "*** " + "图片在识别中，可能需要几分钟，请耐心等待哦！" + " ***",
                                Media: nil,
                        })*/
		}
	} else {

		userPrompt, result := SubstringFrom(req.Messages[last_msg_id].Content, "识别要求")
		if result == true && len(userPrompt) > 0 && len(globalReqImgBytes) > 0 {
			go CallVLModel(userPrompt, globalReqImgBytes)

			req.Messages[last_msg_id].Content = output_text + "*** " + "图片识别中，过几分钟继续询问我, 不问我做完了也不主动通知！询问必须包含：'图片识别结束了吗' 前四个字" + " ***"
			need_prompt = 0
		}else if strings.Contains(req.Messages[last_msg_id].Content, "图片识别"){
			need_prompt = 1
		}else {
			need_prompt = 0
		}
	}
	
	shouldExit := false
	for !shouldExit{
    		select {
    			case msg := <-globalQueue:
        			// 成功拿到数据

        			req.Messages = append(req.Messages, providers.Message{
					Role: "user",
					Content: output_text + "*** " + msg + " ***",
					Media: nil,
				})

				need_prompt = 0
    			default:
        			// 队列已经空了，或者那一瞬间没有新数据了，立刻打破循环
				shouldExit = true
    		}
	}

	if need_prompt != 0 {
        	req.Messages[last_msg_id].Content = output_text + "*** " + "图片识别中，过几分钟继续询问我, 不问我做完了也不主动通知！" + " ***"
	}


        return req, HookDecision{Action: HookActionContinue}, nil
}


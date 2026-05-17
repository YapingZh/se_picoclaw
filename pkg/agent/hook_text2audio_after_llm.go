package agent

import (
    "context"
    "errors"
    "os/exec"
    "bytes"
    "strings"
    "time"

    "github.com/sipeed/picoclaw/pkg/logger"
    //"github.com/sipeed/picoclaw/pkg/providers"
    "github.com/sipeed/picoclaw/pkg/bus"
    "github.com/sipeed/picoclaw/pkg/session"
)

type Text2audioHook struct{}

//var globalQueue = make(chan string, 1000)
var globalReqImgBytes string
var globalVoiceId string
var globalVLUserPromt string
var msgBus *bus.MessageBus

func (h *Text2audioHook) SetBus(bus *bus.MessageBus) { 
	msgBus = bus 
	globalVLUserPromt = ""
	globalReqImgBytes = ""
	globalVoiceId = ""
}
func (h *Text2audioHook) Name() string { return "text2audio" }

func CallTTSModel(Response string) {
	cmd_kill := exec.Command("pkill", "-f", "text2auto.py")
        _, err_kill := cmd_kill.CombinedOutput()
        if err_kill != nil {
                if exitErr, ok := err_kill.(*exec.ExitError); ok {
                        if exitErr.ExitCode() == 1 {
                                err_kill = nil
                        }else {
                                time.Sleep(3 * time.Second)
                        }
                } else{
                        time.Sleep(3 * time.Second)
                }
        } else {
                time.Sleep(3 * time.Second)
        }

        logger.WarnCF("text2audioHook", "hook after LLM", map[string]any{
                        "name":       "text2audioHook",
                        "text":       Response,
        })

	voiceid := "female_cx"
        if len(globalVoiceId) > 0 {
                voiceid = globalVoiceId
        }

        cmd := exec.Command("python3", "/usr/local/text2auto.py", Response, voiceid)
        /*logger.WarnCF("text2audioHook", "hook after LLM", map[string]any{
                        "name":      "text2audioHook",
                        "text":      "Start-->" + Response,
	})*/

        _, err := cmd.CombinedOutput()
        if err != nil {
                logger.WarnCF("text2audioHook", "hook after LLM", map[string]any{
                        "name":      "text2audioHook",
                        "error":     err,
                })
        }

        /*logger.WarnCF("text2audioHook", "hook after LLM", map[string]any{
                        "name":      "text2audioHook",
                        "text":      "End-->" + Response,
        })*/

}

func (h *Text2audioHook) AfterLLM(ctx context.Context, resp *LLMHookResponse) (*LLMHookResponse, HookDecision, error) {
	if h == nil || resp == nil {
		return resp, HookDecision{Action: HookActionContinue}, errors.New("null object")
	}

	CallTTSModel(resp.Response.Content)

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

func SendVLModelResponse(channel, chatID, content, replyToMessageID, agentId, sessionKey string, sessionScope *session.SessionScope) error {
	pubCtx, pubCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer pubCancel()

	outboundCtx := bus.NewOutboundContext(channel, chatID, replyToMessageID)
	outboundAgentID, outboundSessionKey, outboundScope := outboundTurnMetadata(
		agentId,
		sessionKey,
		sessionScope,
	)

	return msgBus.PublishOutbound(pubCtx, bus.OutboundMessage{
		Context:          outboundCtx,
		AgentID:          outboundAgentID,
		SessionKey:       outboundSessionKey,
		Scope:            outboundScope,
		Content:          content,
		ReplyToMessageID: replyToMessageID,
	})
}

func CallVLModel(channel, chatID, replyToMessageID, agentId, sessionKey string, sessionScope *session.SessionScope){
	desc_img_prompt := globalVLUserPromt//"请用中文详细描述这个图片的内容"

	defer func() {
		globalVLUserPromt = "" //globalReqImgBytes = ""

		if err_exp := recover(); err_exp != nil {
			logger.WarnCF("text2audioHook", "hook before LLM", map[string]any {
                                                "name":      "text2audioHook",
                                                "text":      "call VL model happen error",
                                                "error":     err_exp,
                	})
		}
	}()

        if len(globalReqImgBytes) > 0 {
		logger.WarnCF("text2audioHook", "hook before LLM", map[string]any {
                                                "name":      "text2audioHook",
                                                "text":      desc_img_prompt,
                })

                cmd := exec.Command("python3", "/usr/local/img2text.py", desc_img_prompt)
                var stdinBuf bytes.Buffer
                stdinBuf.Write([]byte(globalReqImgBytes))
                cmd.Stdin = &stdinBuf

                output_bytes, err := cmd.CombinedOutput()

		output_text := ""
                if output_bytes == nil {
                        output_text = "图片识别失败了哦"
                } else{
                        output_text = string(output_bytes)
                }

		//globalQueue <- output_text
		globalVLUserPromt = "" //globalReqImgBytes = ""

                logger.WarnCF("text2audioHook", "hook before LLM", map[string]any {
                                                "name":      "text2audioHook",
                                                "text":      output_text,
                                                "error":     err,
                })

		CallTTSModel(output_text)
		SendVLModelResponse(channel, chatID, output_text, replyToMessageID, agentId, sessionKey, sessionScope)
        }else {
		globalVLUserPromt = "" //globalReqImgBytes = ""
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
			globalReqImgBytes = strings.Clone(req.Messages[last_msg_id].Media[0])

			req.Messages[last_msg_id].Media = nil
			req.Messages[last_msg_id].Content = output_text + "*** " + "准备开始图片识别之前，请补充图片识别提示词！提示词必须以：'识别要求' 为开始" + " ***"

			need_prompt = 0
		}
	} else {

		userPrompt, result := SubstringFrom(req.Messages[last_msg_id].Content, "识别要求")
		if result == true && len(userPrompt) > 0 && strings.Contains(globalReqImgBytes, "data:image/") {
			channel := strings.Clone(req.Context.Inbound.Channel)
			chatID := strings.Clone(req.Context.Inbound.ChatID)
			replyToMessageID := strings.Clone(req.Context.Inbound.ReplyToMessageID)
			agentId := strings.Clone(req.Context.Scope.AgentID)
			sessionKey := strings.Clone(req.Meta.SessionKey)

			sessionScope := session.CloneScope(req.Context.Scope)

			globalVLUserPromt = userPrompt
			go CallVLModel(channel, chatID, replyToMessageID, agentId, sessionKey, sessionScope)

			req.Messages[last_msg_id].Content = output_text + "*** " +"开始识图！" + " ***"
			need_prompt = 0
		}else if strings.Contains(req.Messages[last_msg_id].Content, "图片识别"){
			need_prompt = 1
		}else {
			userPrompt, result := SubstringFrom(req.Messages[last_msg_id].Content, "音色切换")
			if result == true && len(userPrompt) > 0 {

				if strings.Contains(userPrompt, "爸爸") {
					globalVoiceId = "baba"
				}else if strings.Contains(userPrompt, "男性普通话") {
					globalVoiceId = "male_cx"
				}else if strings.Contains(userPrompt, "女性普通话") {
					globalVoiceId = "female_cx"
				}else if strings.Contains(userPrompt, "男性英语调") {
					globalVoiceId = "male_eng"
				}else if strings.Contains(userPrompt, "女性英语调") {
					globalVoiceId = "female_eng"
				}else if strings.Contains(userPrompt, "蜡笔小新") {
					globalVoiceId = "lbxiaoxin"
				}else {
					globalVoiceId = "female_cx"
					userPrompt = "女性普通话"
				}

                                req.Messages[last_msg_id].Content = output_text + "*** " + "音色已修改成" + userPrompt + "。支持的音色有 爸爸/男性普通话/女性普通话/男性英语调/女性英语调/蜡笔小新" + " ***"
				need_prompt = 0
                        } else{
				if strings.Contains(req.Messages[last_msg_id].Content, "识别要求") {
					req.Messages[last_msg_id].Content = output_text + "*** " + "请重新发送图片给我！" + " ***"
				}

				need_prompt = 0
			}
		}
	}
	
	/*
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
	*/

	if need_prompt != 0 {
		if strings.Contains(globalReqImgBytes, "data:image/") && len(globalVLUserPromt) > 0 {
        		req.Messages[last_msg_id].Content = output_text + "*** " + "图片识别中！" + " ***"
		} else {
			req.Messages[last_msg_id].Content = output_text + "*** " + "请重新发送图片给我！" + " ***"
		}
	}
	
	logger.WarnCF("text2audioHook", "hook before LLM", map[string]any{
                "name":       "text2audioHook",
                "Last req message":  req.Messages[last_msg_id].Content,
        })

        return req, HookDecision{Action: HookActionContinue}, nil
}


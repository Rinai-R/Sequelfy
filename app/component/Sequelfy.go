package component

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Rinai-R/Sequelfy/app/consts"
	"github.com/Rinai-R/Sequelfy/app/model"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/sirupsen/logrus"
)

type Sequelfy struct {
	App compose.Runnable[*model.Request, *schema.Message]
}

func NewSequelfyApp() *Sequelfy {
	ctx := context.Background()
	chatModel, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		BaseURL: consts.AgentBaseURL,
		Region:  consts.AgentRegion,
		APIKey:  consts.AgentAPIKey,
		Model:   consts.AgentModel,
	})
	if err != nil {
		klog.Fatal("init eino failed: ", err)
	}

	template := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage("你是一个可爱的小说续写专用女仆，你的任务是帮助用户续写小说，使用可爱的语气和风格进行交流"),
		schema.SystemMessage("虽然你需要扮演可爱女仆的角色，但是你的最终目的是帮助用户续写小说，注意，不可以与用户调情"),
		schema.SystemMessage("我会提供你与用户的历史消息，所以你应该允许用户分片传输它的小说"),
		schema.SystemMessage("同时，你还需要将所有的小说信息包括用户提供的小说内容都返回给用户"),
		schema.MessagesPlaceholder("history_message", true),
		schema.SystemMessage("你和用户的历史消息为: {history_message}"),
		schema.SystemMessage("当前时间为: {current_time}"),
		schema.UserMessage("用户当前发出的消息为: {unread_data}"),
	)

	lamda := compose.InvokableLambda(func(ctx context.Context, req *model.Request) (map[string]any, error) {
		output := map[string]any{
			"current_time":    time.Now().Format("2006-01-02 15:04:05"),
			"unread_data":     req.UnreadData,
			"history_message": req.History,
		}
		return output, nil
	})
	AskChain := compose.NewChain[*model.Request, *schema.Message]()
	AskChain.
		AppendLambda(lamda).
		AppendChatTemplate(template).
		AppendChatModel(chatModel)
	NewSequelfyApp, err := AskChain.Compile(ctx)
	if err != nil {
		klog.Fatal("init eino failed: ", err)
	}
	return &Sequelfy{NewSequelfyApp}
}

func (s *Sequelfy) Run(ctx context.Context, history []*schema.Message, sigChan chan os.Signal) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Println("user: ")
			var input string
			fmt.Scan(&input)

			if input == "" {
				logrus.Info("Empty input. Skipping...")
				continue
			}
			resp, err := s.run(ctx, &model.Request{
				UnreadData: input,
				History:    history,
			})
			if err != nil {
				logrus.Error(err)
				continue
			}
			fmt.Println("智能续写的可爱女仆: ", resp.Text)
			history = append(history, &schema.Message{
				Role:    schema.User,
				Content: input,
			})
			history = append(history, &schema.Message{
				Role:    schema.Assistant,
				Content: resp.Text,
			})
			go func() {
				s.Save(input, resp.Text)
			}()
		}
	}
}

func (s *Sequelfy) run(ctx context.Context, req *model.Request) (*model.Response, error) {
	output, err := s.App.Invoke(ctx, req)
	if err != nil {
		return nil, err
	}
	return &model.Response{
		Text: output.Content,
	}, nil
}

func (s *Sequelfy) Clear() error {
	msg, err := s.Load()
	if err != nil {
		return err
	}
	f, err := os.Create(consts.HistoryFilePath + msg[0].Content[:min(15, len(msg[0].Content))] + ".json")
	if err != nil {
		return err
	}
	defer f.Close()
	bytes, err := os.ReadFile(consts.CurFilePath)
	if err != nil {
		return err
	}
	_, err = f.Write(bytes)
	if err != nil {
		return err
	}
	return os.Remove(consts.CurFilePath)
}

func (s *Sequelfy) Save(UserText, BotText string) error {
	var msg []*schema.Message
	msg = append(msg, &schema.Message{
		Role:    schema.User,
		Content: UserText,
	})
	msg = append(msg, &schema.Message{
		Role:    schema.Assistant,
		Content: BotText,
	})
	bytes, err := sonic.Marshal(msg)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(consts.CurFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

func (s *Sequelfy) Load() ([]*schema.Message, error) {
	var msg = []*schema.Message{}
	f, err := os.OpenFile(consts.CurFilePath, os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	f.Close()
	bytes, err := os.ReadFile(consts.CurFilePath)
	if err != nil {
		return nil, err
	}
	if len(bytes) == 0 {
		return msg, nil
	}
	err = sonic.Unmarshal(bytes, &msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *Sequelfy) Change(HistoryFileName string) error {
	f, err := os.Open(consts.CurFilePath)
	if err == nil {
		f.Close()
		s.Clear()
	}
	bytes, err := os.ReadFile(consts.HistoryFilePath + HistoryFileName + ".json")
	if err != nil {
		return err
	}
	f, err = os.OpenFile(consts.CurFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

package config

import (
	"strconv"
)

type Strings struct {
	Lang     string
	CurrConf string
	Usage    usagestr
	Config   configstr
	Reply    replystr
	Error    errorstr
}

type errorstr struct {
	Title               string
	UnknownTitle        string
	UnknownDesc         string
	NoCmd               string
	SubCmd              string
	Syntax              string
	SyntaxDesc          string
	MustBoolean         string
	MustValue           string
	MustTimeoutDuration string
	LongResponse        string
	TimeOut             string
}

type usagestr struct {
	Title  string
	Config configusagestr
	Cmd    cmdusagestr
}

type configstr struct {
	Title    string
	Announce string
	Item     itemstr
}

type itemstr struct {
	Prefix   string
	Lang     string
	Maxtoken string
	Timeout  string
}

type replystr struct {
	ExecTime string
	Second   string
	Cost     string
}

type cmdusagestr struct {
	ChatTitle string
	ChatUsage string
	PingTitle string
	PingUsage string
	HelpTitle string
	HelpUsage string
	ConfTitle string
	ConfUsage string
}

type configusagestr struct {
	Desc     string
	Prefix   string
	Lang     string
	MaxToken string
	TimeOut  string
}

var (
	Lang map[string]Strings
)

func loadLang() {
	Lang = map[string]Strings{}
	Lang["japanese"] = Strings{
		Lang:     "japanese",
		CurrConf: "現在の設定",
		Usage: usagestr{
			Title: "使い方: ",
			Config: configusagestr{
				Desc:     "各種設定を行います。\n設定項目と内容は以下をご確認ください。",
				Prefix:   "コマンドの接頭詞を指定します。\n現在の設定は`" + CurrentConfig.Guild.Prefix + "`です。",
				Lang:     "言語を指定します。現在の設定は`" + CurrentConfig.Guild.Lang + "`です。",
				MaxToken: "使用する最大トークン数を指定します。現在の設定は`" + strconv.Itoa(CurrentConfig.Guild.MaxToken) + "`です。",
			},
			Cmd: cmdusagestr{
				ChatTitle: "`" + CurrentConfig.Guild.Prefix + "chat`",
				ChatUsage: "`" + CurrentConfig.Guild.Prefix + "chat " + "<message>`\nChatGPTに文章を送信します。\n🤔をリアクションした場合は処理を通すのに成功していますので、処理が完了するまでお待ちください。\n処理が完了すると返信します。\n`-l <int>`でログを読み込むことが出来ます。",
				PingTitle: "`" + CurrentConfig.Guild.Prefix + "ping`",
				PingUsage: "`" + CurrentConfig.Guild.Prefix + "ping`\nBotが起動状態か確認できます。\n返信とともに🏓をリアクションした場合、Botが利用できる状態です。",
				HelpTitle: "`" + CurrentConfig.Guild.Prefix + "help`",
				HelpUsage: "`" + CurrentConfig.Guild.Prefix + "help`\nBotの使い方を確認できます。\nこのメッセージを返信します。",
				ConfTitle: "`" + CurrentConfig.Guild.Prefix + "config`",
				ConfUsage: "`" + CurrentConfig.Guild.Prefix + "config <SetName> <SetValue>`\nBotの設定を確認できます。\n何も引数を設定しなかった場合、現在の設定を表示します。\n引数を設定すると、その設定を変更できます。",
			},
		},
		Config: configstr{
			Title:    "設定の更新",
			Announce: "\"に更新しました。",
			Item: itemstr{
				Prefix:   "Prefixを\"",
				Lang:     "botの使用言語を\"",
				Maxtoken: "botが使用するトークンの最大値を\"",
				Timeout:  "タイムアウトの時間を\"",
			},
		},
		Reply: replystr{
			ExecTime: "実行時間: ",
			Second:   "秒",
			Cost:     "この月に使用された料金: ¥ ",
		},
		Error: errorstr{
			UnknownTitle:        "予期せぬエラーが発生しました。",
			UnknownDesc:         "この問題は管理者に報告されます。",
			NoCmd:               "そのようなコマンドはありません。",
			SubCmd:              "引数が不正です。",
			Syntax:              "構文エラー",
			SyntaxDesc:          "パラメータの解析に失敗しました。\nコマンドの構文が正しいかお確かめください。",
			MustBoolean:         "その引数は`true`または`false`である必要があります。",
			MustValue:           "その引数は`1`から`4095`の範囲の整数である必要があります。",
			MustTimeoutDuration: "その引数は1以上の自然数である必要があります。",
			LongResponse:        "AIの生成した文章が長すぎました。指示を変更してもう一度お試しください。",
			TimeOut:             "要求がタイムアウトしました。もう一度お試しください。",
		},
	}
	Lang["english"] = Strings{
		Lang:     "english",
		CurrConf: "Current config",
		Usage: usagestr{
			Title: "Usage: ",
			Config: configusagestr{
				Desc:     "Do configuration.\nItem list is below.",
				Prefix:   "Specify command prefix.\nCurrent config is `" + CurrentConfig.Guild.Prefix + "`.",
				Lang:     "Specify language.\nCurrent config is `" + CurrentConfig.Guild.Lang + "`.",
				MaxToken: "Specify MaxTokens.\nCurrent config is `" + strconv.Itoa(CurrentConfig.Guild.MaxToken) + "`.",
			},
			Cmd: cmdusagestr{
				ChatTitle: "`" + CurrentConfig.Guild.Prefix + "chat`",
				ChatUsage: "`" + CurrentConfig.Guild.Prefix + "chat " + "<message>`\nSend a message to ChatGPT.\nIf Bot reacted 🤔, your message has been passing the process, so please wait for the process to complete.\nWhen the process is complete, Bot send reply to an embed.\nAlso, you can load logs by `-r <int>`.",
				PingTitle: "`" + CurrentConfig.Guild.Prefix + "ping`",
				PingUsage: "`" + CurrentConfig.Guild.Prefix + "ping`\nYou can check if the Bot is in startup status. \nIf Bot has reacted 🏓 and sent reply to an embed to your ping message, Bot is in startup status.",
				HelpTitle: "`" + CurrentConfig.Guild.Prefix + "help`",
				HelpUsage: "`" + CurrentConfig.Guild.Prefix + "help`\nYou can check how to use the Bot. \nSend reply to this message.",
				ConfTitle: "`" + CurrentConfig.Guild.Prefix + "config`",
				ConfUsage: "`" + CurrentConfig.Guild.Prefix + "config <SetName> <SetValue>`\nYou can check the configuration of Bot. \nIf you don't give any arguments, the current settings are displayed. \nIf you set any of the arguments, you can change its settings.",
			},
		},
		Config: configstr{
			Title:    "Config Update",
			Announce: " \".",
			Item: itemstr{
				Prefix:   "Prefix for \"",
				Lang:     "Language used by bot for \"",
				Maxtoken: "Max Tokens used by bot for \"",
				Timeout:  "The time until timeout for \"",
			},
		},
		Reply: replystr{
			ExecTime: "Execution time: ",
			Second:   "s",
			Cost:     "Fees used in this month: $ ",
		},
		Error: errorstr{
			UnknownTitle:        "Unexpected error is occurred.",
			UnknownDesc:         "This issue will be reported",
			NoCmd:               "No such command.",
			SubCmd:              "Invalid argument.",
			Syntax:              "Syntax error",
			SyntaxDesc:          "Failed to parse parameter.\nPlease check your command syntax.",
			MustBoolean:         "That argument must be `true` or `false`.",
			MustValue:           "That argument must be `1` to `4095` and integer value.",
			MustTimeoutDuration: "That argument must be a natural number greater than or equal to 1.",
			LongResponse:        "The AI-generated text is too long. Please modify your instructions and try again.",
			TimeOut:             "The request has timed out. Please try again.",
		},
	}
}

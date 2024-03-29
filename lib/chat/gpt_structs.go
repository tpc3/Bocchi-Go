package chat

type OpenaiRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Top_p       float64   `json:"top_p"`
	Temperature float64   `json:"temperature"`
}

type OpenaiRequestImg struct {
	Model       string  `json:"model"`
	Messages    []Img   `json:"messages"`
	Top_p       float64 `json:"top_p"`
	Temperature float64 `json:"temperature"`
	MaxToken    int     `json:"max_tokens"`
}

type OpenaiResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int      `json:"created"`
	Choices []Choice `json:"choices"`
	Usages  Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Messages     Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Content interface{}

type Img struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ImageContent struct {
	Type     string   `json:"type"`
	ImageURL ImageURL `json:"image_url"`
}

type ImageURL struct {
	Url    string `json:"url"`
	Detail string `json:"detail"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ErrorResponse struct {
	Error ErrorDetails `json:"error"`
}

type ErrorDetails struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param"`
	Code    string `json:"code"`
}

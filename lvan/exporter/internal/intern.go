package internal

// 命令请求结构
type CommandRequest struct {
	Cmd     string            `json:"cmd" yaml:"cmd"`
	Version string            `json:"version" yaml:"version"`
	Args    []string          `json:"args" yaml:"args"`
	Env     map[string]string `json:"-" yaml:"env,omitempty"`
}

type CmdResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Id   string `json:"id"`
}
type ResultResponse struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Id   string         `json:"id"`
	Job  CommandRequest `json:"job"`
}

type CommandMeta struct {
	Encoding  Charset  `yaml:"encoding"`
	Shell     []string `yaml:"shell"`
	Resources []string `yaml:"resources"`
}

type Charset string

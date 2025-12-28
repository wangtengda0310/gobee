package customCmd

var Commands = map[string]func([]string){}

var RegisterCommand = func(cmd string, f func([]string)) {
	Commands[cmd] = f
}

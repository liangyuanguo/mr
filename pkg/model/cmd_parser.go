package model

import (
	"strconv"
	"strings"
)

type CmdParser struct {
	FullArgs []string
	Args     []string
	Options  map[string]string
}

func Parse(cmd string) *CmdParser {
	c := &CmdParser{
		Options: make(map[string]string),
	}
	c.parseCmd(cmd)
	return c
}

func (c *CmdParser) parseCmd(cmd string) {
	cmd = strings.TrimLeft(cmd, " ") + " "
	worldBegin := -1
	hasArg := false

	for i := 0; i < len(cmd); i++ {
		if cmd[i] != ' ' {
			if worldBegin == -1 {
				worldBegin = i
			}
			continue
		} else if worldBegin == -1 {
			continue
		}

		word := cmd[worldBegin:i]
		worldBegin = -1

		if word == "--" || word == "---" {
			remaining := strings.TrimSpace(cmd[i+1:])
			if remaining != "" {
				c.Args = append(c.Args, remaining)
				c.FullArgs = append(c.FullArgs, remaining)
			}
			break
		} else if word == "-" && !hasArg {
			hasArg = true
			continue
		}

		c.FullArgs = append(c.FullArgs, word)
		if hasArg {
			c.Args = append(c.Args, word)
			hasArg = false
		} else if strings.HasPrefix(word, "-") {
			eqIndex := strings.Index(word, "=")
			if eqIndex == -1 {
				c.Options[word[1:]] = ""
			} else {
				c.Options[word[1:eqIndex]] = word[eqIndex+1:]
			}
		} else {
			c.Args = append(c.Args, word)
		}
	}
}

func (c *CmdParser) ToCmd() string {
	var cmdString strings.Builder
	if len(c.Args) > 0 {
		cmdString.WriteString(c.Args[0])
		cmdString.WriteString(" ")
	}

	for k, v := range c.Options {
		cmdString.WriteString("-")
		cmdString.WriteString(k)
		if v != "" {
			cmdString.WriteString("=")
			cmdString.WriteString(v)
		}
		cmdString.WriteString(" ")
	}

	if len(c.Args) > 1 {
		midArgs := c.Args[1 : len(c.Args)-1]
		for _, arg := range midArgs {
			if strings.HasPrefix(arg, "-") {
				cmdString.WriteString("- ")
			}
			cmdString.WriteString(arg)
			cmdString.WriteString(" ")
		}

		lastArg := c.Args[len(c.Args)-1]
		if strings.Contains(lastArg, " ") {
			cmdString.WriteString("-- ")
		}
		cmdString.WriteString(lastArg)
	}

	return strings.TrimSpace(cmdString.String())
}

func (c *CmdParser) Copy() *CmdParser {
	c2 := &CmdParser{
		Options: make(map[string]string),
	}
	c2.Args = make([]string, len(c.Args))
	copy(c2.Args, c.Args)
	for k, v := range c.Options {
		c2.Options[k] = v
	}
	return c2
}

func (c *CmdParser) GetArgInt(pos int, dft int) int {
	if pos >= len(c.Args) {
		return dft
	}
	val, err := strconv.Atoi(c.Args[pos])
	if err != nil {
		return dft
	}
	return val
}

func (c *CmdParser) GetArgStr(pos int, dft string) string {
	if pos >= len(c.Args) {
		return dft
	}
	val := c.Args[pos]
	return val
}

func (c *CmdParser) GetArgFloat(pos int, dft float64) float64 {
	if pos >= len(c.Args) {
		return dft
	}
	val, err := strconv.ParseFloat(c.Args[pos], 64)
	if err != nil {
		return dft
	}
	return val
}

func (c *CmdParser) GetOptInt(opt string, dft int) int {
	val, ok := c.Options[opt]
	if !ok {
		return dft
	}
	result, err := strconv.Atoi(val)
	if err != nil {
		return dft
	}
	return result
}

func (c *CmdParser) GetOptFloat(opt string, dft float64) float64 {
	val, ok := c.Options[opt]
	if !ok {
		return dft
	}
	result, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return dft
	}
	return result
}

func (c *CmdParser) GetArgBool(pos int, dft bool) bool {
	if pos >= len(c.Args) {
		return dft
	}
	val := strings.ToLower(c.Args[pos])
	return checkBoolString(val) || dft
}

func (c *CmdParser) GetOptBool(opt string, dft bool) bool {
	val, ok := c.Options[opt]
	if !ok {
		return dft
	}
	return checkBoolString(strings.ToLower(val))
}

func (c *CmdParser) GetOptStr(opt string, dft string) string {
	val, ok := c.Options[opt]
	if !ok {
		return dft
	}
	return val
}

func checkBoolString(s string) bool {
	return s == "true" || s == "1" || s == "t" || s == "y" || s == "yes" || s == ""
}

func (c *CmdParser) GetOptStrArr(opt string) []string {
	val, ok := c.Options[opt]
	if !ok {
		return []string{}
	}
	return strings.Split(val, ",")
}

func (c *CmdParser) GetOptIntArr(opt string) []int {
	var result []int
	val, ok := c.Options[opt]
	if !ok {
		return result
	}
	items := strings.Split(val, ",")
	for _, item := range items {
		i, err := strconv.Atoi(item)
		if err != nil {
			i = 0
		}
		result = append(result, i)
	}
	return result
}

func (c *CmdParser) GetOptFloatArr(opt string) []float64 {
	var result []float64
	val, ok := c.Options[opt]
	if !ok {
		return result
	}
	items := strings.Split(val, ",")
	for _, item := range items {
		f, err := strconv.ParseFloat(item, 64)
		if err != nil {
			f = 0.0
		}
		result = append(result, f)
	}
	return result
}

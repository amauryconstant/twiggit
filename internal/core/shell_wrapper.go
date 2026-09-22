package core

import (
	"fmt"
	"strings"
	"time"
)

func ShellWrapper(shellType ShellType) (string, error) {
	template, err := wrapperTemplate(shellType)
	if err != nil {
		return "", err
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	content := strings.ReplaceAll(template, "{{SHELL_TYPE}}", string(shellType))
	content = strings.ReplaceAll(content, "{{TIMESTAMP}}", now)

	return content, nil
}

func wrapperTemplate(shellType ShellType) (string, error) {
	config := shellTemplateConfigFor(shellType)

	body := `### BEGIN TWIGGIT WRAPPER
# Twiggit ` + string(shellType) + ` wrapper - Generated on {{TIMESTAMP}}
` + config.funcDef + `
` + config.caseBegin + `
    cd)
        # Handle cd command with directory change
        target_dir=$(command twiggit ` + config.argsVar + `)
        if [ $? -eq 0 ] && [ -n "$target_dir" ]; then
            builtin cd "$target_dir"
        fi
        ` + config.elif + `
    create)
        # Handle create command with -C flag
	` + config.ifSyntax + ` " ` + config.argsVar + ` " == *" -C "* ` + config.andOperator + ` " ` + config.argsVar + ` " == *" --cd "* ` + config.thenSyntax + `
			target_dir=$(command twiggit ` + config.argsVar + `)
			if [ $? -eq 0 ] && [ -n "$target_dir" ]; then
				builtin cd "$target_dir"
			fi
		` + config.elseSyntax + `
			command twiggit ` + config.argsVar + `
		` + config.fiSyntax + `
		` + config.elif + `
	delete)
		# Handle delete command with -C flag
		` + config.ifSyntax + ` " ` + config.argsVar + ` " == *" -C "* ` + config.andOperator + ` " ` + config.argsVar + ` " == *" --cd "* ` + config.thenSyntax + `
            target_dir=$(command twiggit ` + config.argsVar + `)
            if [ $? -eq 0 ] && [ -n "$target_dir" ]; then
                builtin cd "$target_dir"
            fi
        ` + config.elseSyntax + `
            command twiggit ` + config.argsVar + `
        ` + config.fiSyntax + `
        ` + config.elif + `
    *)
        # Pass through all other commands
        command twiggit ` + config.argsVar + `
        ` + config.elif + `
` + config.caseEnd + `
` + config.funcEnd + `
### END TWIGGIT WRAPPER`

	switch shellType {
	case ShellBash:
		return body + `
### BEGIN TWIGGIT COMPLETION
source <(command twiggit _carapace bash)
### END TWIGGIT COMPLETION`, nil
	case ShellZsh:
		return body + `
### BEGIN TWIGGIT COMPLETION
source <(command twiggit _carapace zsh)
### END TWIGGIT COMPLETION`, nil
	case ShellFish:
		return body + `
### BEGIN TWIGGIT COMPLETION
command twiggit _carapace fish | source
### END TWIGGIT COMPLETION`, nil
	default:
		return "", fmt.Errorf("unsupported shell type %q", shellType)
	}
}

type shellTemplateConfig struct {
	caseBegin     string
	caseEnd       string
	elif          string
	caseSeparator string
	argsVar       string
	ifSyntax      string
	thenSyntax    string
	elseSyntax    string
	fiSyntax      string
	andOperator   string
	funcDef       string
	funcEnd       string
}

func shellTemplateConfigFor(shellType ShellType) shellTemplateConfig {
	switch shellType {
	case ShellBash, ShellZsh:
		return shellTemplateConfig{
			caseBegin:     `case "$1" in`,
			caseEnd:       "esac",
			elif:          ";;",
			caseSeparator: "    ",
			argsVar:       `"$@"`,
			ifSyntax:      "if [[",
			thenSyntax:    "]]; then",
			elseSyntax:    "else",
			fiSyntax:      "fi",
			andOperator:   "]] || [[",
			funcDef:       "twiggit() {",
			funcEnd:       "}",
		}
	case ShellFish:
		return shellTemplateConfig{
			caseBegin:     `switch "$argv[1]"`,
			caseEnd:       "end",
			elif:          "",
			caseSeparator: "    ",
			argsVar:       "$argv",
			ifSyntax:      "if",
			thenSyntax:    "",
			elseSyntax:    "else",
			fiSyntax:      "end",
			andOperator:   "or",
			funcDef:       "function twiggit",
			funcEnd:       "end",
		}
	default:
		return shellTemplateConfig{
			caseBegin:     `case "$1" in`,
			caseEnd:       "esac",
			elif:          ";;",
			caseSeparator: "    ",
			argsVar:       `"$@"`,
			ifSyntax:      "if [[",
			thenSyntax:    "]]; then",
			elseSyntax:    "else",
			fiSyntax:      "fi",
			andOperator:   "]] || [[",
			funcDef:       "twiggit() {",
			funcEnd:       "}",
		}
	}
}

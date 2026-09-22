package output

import (
	"strings"

	"twiggit/internal/core"
)

// Shell wrapper templates. Each non-empty template substitutes
// ${SHELL_NAME} with the user's shell so init-scripts can refer
// to themselves symbolically. Empty template falls through to
// the per-shell default below.
const (
	placeholderShellName = "${SHELL_NAME}"

	// bashInit wraps twiggit init bash in a function so the
	// user can `eval "$(twiggit init bash)"` cleanly.
	bashInit = `# twiggit shell integration for bash
# Generated for ${SHELL_NAME}; eval "$(twiggit init bash)" to install.
_twiggit_cd() {
    local target
    target=$(twiggit cd "$@")
    if [[ -n "$target" ]]; then
        cd "$target" || return
    fi
}
alias tg=_twiggit_cd
`

	// zshInit mirrors bashInit but with zsh-correct quoting.
	zshInit = `# twiggit shell integration for zsh
# Generated for ${SHELL_NAME}; eval "$(twiggit init zsh)" to install.
_twiggit_cd() {
    local target
    target=$(twiggit cd "$@")
    if [[ -n "$target" ]]; then
        cd "$target" || return
    fi
}
alias tg=_twiggit_cd
`

	// fishInit uses fish-native function syntax.
	fishInit = `# twiggit shell integration for fish
# Generated for ${SHELL_NAME}; twiggit init fish | source to install.
function _twiggit_cd
    set -l target (twiggit cd $argv)
    if test -n "$target"
        cd $target
    end
end
alias tg=_twiggit_cd
`
)

// ComposeWrapper renders the shell wrapper script. When template
// is non-empty, ${SHELL_NAME} and ${SHELL_TYPE} are substituted
// and the result is returned verbatim. When template is empty,
// the per-shell default template (bashInit / zshInit / fishInit)
// is used. Unknown shell types fall back to bash so init scripts
// always produce something runnable.
func ComposeWrapper(template string, shellType core.ShellType) string {
	if template == "" {
		template = defaultTemplate(shellType)
	}
	return substitute(template, shellType)
}

func defaultTemplate(shellType core.ShellType) string {
	switch shellType {
	case core.ShellZsh:
		return zshInit
	case core.ShellFish:
		return fishInit
	case core.ShellBash:
		return bashInit
	default:
		return bashInit
	}
}

func substitute(template string, shellType core.ShellType) string {
	out := strings.ReplaceAll(template, placeholderShellName, string(shellType))
	out = strings.ReplaceAll(out, "${SHELL_TYPE}", string(shellType))
	return out
}

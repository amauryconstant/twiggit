# twiggit.plugin.zsh - Zsh plugin for twiggit
# Works with Oh My Zsh, antidote, zinit, or standalone sourcing

if (( ! $+commands[twiggit] )); then
  return
fi

eval "$(twiggit init zsh)"

_TwiggitLazyCompletion() {
  unfunction _TwiggitLazyCompletion
  source <(twiggit _carapace zsh)
  _comps[twiggit]=_twiggit
}

typeset -g -A _comps
_comps[twiggit]=_TwiggitLazyCompletion

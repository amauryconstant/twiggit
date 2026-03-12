# twiggit.bash - Bash integration for twiggit
# Source this file from your .bashrc

if ! command -v twiggit &>/dev/null; then
  return
fi

eval "$(twiggit init bash)"

source <(twiggit _carapace bash)

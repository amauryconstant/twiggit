# twiggit.fish - Fish integration for twiggit
# Source this file or copy to ~/.config/fish/conf.d/twiggit.fish

if not type -q twiggit
    return
end

twiggit init fish | source

twiggit _carapace fish | source

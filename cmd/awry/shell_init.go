package main

func shellInitScript(shell string) string {
	if shell == "fish" {
		return `function awry
  switch "$argv[1]"
    case '' use
      set -l _awry_output (env AWRY_SHELL=fish command awry $argv)
      or return $status
      if test -n "$_awry_output"
        eval $_awry_output
      end
    case '*'
      command awry $argv
  end
end
`
	}

	return `awry() {
  case "$1" in
    ""|use)
      local _awry_output
      _awry_output="$(command awry "$@")" || return $?
      if [ -n "$_awry_output" ]; then
        eval "$_awry_output"
      fi
      ;;
    *)
      command awry "$@"
      ;;
  esac
}
`
}

package main

func shellInitScript(shell string) string {
	_ = shell
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

#!/usr/bin/env bash

ROOT_DIR=`dirname "$(perl -e 'use Cwd "abs_path"; print abs_path(shift)' $0)"`
CLI_ENTRYPOINT=`basename $0`

COLOR_BLACK="\033[30m"
COLOR_RED="\033[31m"
COLOR_GREEN="\033[32m"
COLOR_YELLOW="\033[33m"
COLOR_BLUE="\033[34m"
COLOR_MAGENTA="\033[35m"
COLOR_CYAN="\033[36m"
COLOR_LIGHT_GRAY="\033[37m"
COLOR_DARK_GRAY="\033[38m"
COLOR_NORMAL="\033[39m"

function bcli_trim_whitespace() {
    # Function courtesy of http://stackoverflow.com/a/3352015
    local var="$*"
    var="${var#"${var%%[![:space:]]*}"}"   # remove leading whitespace characters
    var="${var%"${var##*[![:space:]]}"}"   # remove trailing whitespace characters
    echo -n "$var"
}

function bcli_show_header() {
    echo -e "$(bcli_trim_whitespace "$(cat "$1/.name")")"
    echo -e "${COLOR_CYAN}Version  ${COLOR_NORMAL}$(bcli_trim_whitespace "$(cat "$1/.version")")"
    echo -e "${COLOR_CYAN}Author   ${COLOR_NORMAL}$(bcli_trim_whitespace "$(cat "$1/.author")")"
}

function blci_help() {
  # If we don't have any additional help arguments, then show the app's
  # header as well.
  if [ $# == 1 ]; then
      bcli_show_header "$ROOT_DIR/.go"
  fi

  # Locate the correct level to display the helpfile for, either a directory
  # with no further arguments, or a command file.
  HELP_FILE="$ROOT_DIR/.go/"
  HELP_ARG_START=2
  while [[ -d "$HELP_FILE" && $HELP_ARG_START -le $# ]]; do
      HELP_FILE="$HELP_FILE/${!HELP_ARG_START}"
      HELP_ARG_START=$(($HELP_ARG_START+1))
  done

  # If we've got a directory's helpfile to show, then print out the list of
  # commands in that directory along with its help content.
  if [[ -d "$HELP_FILE" ]]; then
      echo -e "${COLOR_GREEN}$CLI_ENTRYPOINT ${COLOR_CYAN}${@:2:$(($HELP_ARG_START-1))} ${COLOR_NORMAL}"

      # If there's a help file available for this directory, then show it.
      if [[ -f "$HELP_FILE/.help" ]]; then
          cat "$HELP_FILE/.help"
          echo ""
      fi

      echo ""
      echo -e "${COLOR_MAGENTA}Commands${COLOR_NORMAL}"
      echo ""

      for file in $HELP_FILE/*; do
          cmd=`basename "$file"`

          # Don't show hidden files as available commands
          if [[ "$cmd" != .* && "$cmd" != *.usage && "$cmd" != *.help ]]; then
              echo -en "${COLOR_GREEN}$CLI_ENTRYPOINT ${COLOR_CYAN}${@:2:$(($HELP_ARG_START-1))} $cmd ${COLOR_NORMAL}"

              if [[ -f "$file.usage" ]]; then
                  bcli_trim_whitespace "$(cat "$file.usage")"
                  echo ""
              elif [[ -d "$file" ]]; then
                  echo -e "${COLOR_MAGENTA}...${COLOR_NORMAL}"
              else
                  echo ""
              fi
          fi
      done

      exit 0
  fi

  echo -en "${COLOR_GREEN}$CLI_ENTRYPOINT ${COLOR_CYAN}${@:2:$(($HELP_ARG_START-1))} ${COLOR_NORMAL}"
  if [[ -f "$HELP_FILE.usage" ]]; then
      bcli_trim_whitespace "$(cat "$HELP_FILE.usage")"
      echo ""
  else
      echo ""
  fi


  if [[ -f "$HELP_FILE.help" ]]; then
      cat "$HELP_FILE.help"
      echo ""
  fi
}

# Locate the correct command to execute by looking through the .go directory
# for folders and files which match the arguments provided on the command line.
CMD_FILE="$ROOT_DIR/.go/"
CMD_ARG_START=1
while [[ -d "$CMD_FILE" && $CMD_ARG_START -le $# ]]; do

    # If the user provides help as the last argument on a directory, then
    # show them the help for that directory rather than continuing
    if [[ "${!CMD_ARG_START}" == "help" ]]; then
        # Strip off the "help" portion of the command
        ARGS=("$@")
        unset "ARGS[$((CMD_ARG_START-1))]"
        ARGS=("${ARGS[@]}")

        blci_help $0 ${ARGS[@]}
        exit 3
    fi

    CMD_FILE="$CMD_FILE/${!CMD_ARG_START}"
    CMD_ARG_START=$(($CMD_ARG_START+1))
done

# Place the arguments for the command in their own list
# to make future work with them easier.
CMD_ARGS=("${@:CMD_ARG_START}")

# If we hit a directory by the time we run out of arguments, then our user
# hasn't completed their command, so we'll show them the help for that directory
# to help them along.
if [ -d "$CMD_FILE" ]; then
    blci_help $0 $@
    exit 3
fi

# If we didn't couldn't find the exact command the user entered then warn them
# about it, and try to be helpful by displaying help for that directory.
if [[ ! -f "$CMD_FILE" ]]; then
    blci_help $0 ${@:1:$(($CMD_ARG_START-1))}
    >&2 echo -e "\033[31mWe could not find the command \033[36m$CLI_ENTRYPOINT ${@:1:$CMD_ARG_START}\033[39m"
    >&2 echo -e "To help out, we've shown you the help docs for \033[36m$CLI_ENTRYPOINT ${@:1:$(($CMD_ARG_START-1))}\033[39m"
    exit 3
fi

# If --help is passed as one of the arguments to the command then show
# the command's help information.
arg_i=0 # We need the index to be able to strip list indices
for arg in "${CMD_ARGS[@]}"; do
    if [[ "${arg}" == "--help" ]]; then
        # Strip off the `--help` portion of the command
        unset "CMD_ARGS[$arg_i]"
        CMD_ARGS=("${CMD_ARGS[@]}")

        # Pass the result to the help script for interrogation
        blci_help $0 ${@:1:$((CMD_ARG_START - 1))} ${CMD_ARGS[@]}
        exit 3
    fi
    arg_i=$((arg_i+1))
done

# Run the command and capture its exit code for introspection
"$CMD_FILE" ${CMD_ARGS[@]}
EXIT_CODE=$?

# If the command exited with an exit code of 3 (our "show help" code)
# then show the help documentation for the command.
if [[ $EXIT_CODE == 3 ]]; then
    blci_help $0 $@
fi

# Exit with the same code as the command
exit $EXIT_CODE

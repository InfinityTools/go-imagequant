#!/bin/sh

# A build script that automatically picks the right library from the subfolders in "libs".
# Use this script if you are unable or don't want to use the system library.

show_help() {
  echo "Usage $0 [options]"
  echo ""
  echo "Options:"
  echo "  --libdir path   Override library path"
  echo "  --help          This help"
  exit 0
}

# show_message(msg: string, level: int): Print "msg" and exit with "level". Default level is 0.
show_message() {
  if test $# != 0; then
    echo $1
    shift
  fi
  if test $# != 0; then
    exit $1
  else
    exit 0
  fi
}

if test ! $(which go); then
  show_message "Error: Go compiler not found." 1
fi

# Setting up important variables
gosrcpath=$(go env GOPATH)
packageRoot="github.com/InfinityTools/go-imagequant"
ldprefix="$gosrcpath/src/$packageRoot"
uselibdir=0

# Evaluating command line arguments...
while test $# != 0
do
  case $1 in
  --libdir)
    shift
    if test $# = 0; then
      show_message "Missing argument: --libdir" 1
    fi
    uselibdir=1
    libdir="$1"
    ;;
  --help)
    show_help
    ;;
  esac
  shift
done


# Setting package-specific linker options
ldargs="-limagequant -lm"

if test $uselibdir = 0; then
  libos=$(go env GOOS)
  libarch=$(go env GOARCH)
  echo "Detected: os=$libos, arch=$libarch"
  libdir="$ldprefix/libs/$libos/$libarch"
else
    echo "Using libdir: $libdir"
fi

if test ! -d "$libdir"; then
  show_message "Error: Path does not exist: $libdir" 1
fi

echo "Building library..."
CGO_LDFLAGS="-L$libdir $ldargs" go build && go install && show_message "Finished." 0 || show_message "Cancelled." 1

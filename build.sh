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

# terminate(msg: string, level: int): Print "msg" and exit with "level". Default level is 0.
terminate() {
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

# Checking Go compiler
if test ! $(which go); then
  terminate "Error: Go compiler not found." 1
fi

# Package-specific settings
pkgRoot="github.com/InfinityTools"
uselibdir=0

# Evaluating command line arguments...
while test $# != 0
do
  case $1 in
  --libdir)
    shift
    if test $# = 0; then
      terminate "Missing argument: --libdir" 1
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

  pkgImagequant=$pkgRoot/go-imagequant
  ldprefix=$(go list -f {{.Dir}} $pkgImagequant)
  test $? = 0 || terminate "Package not found: $pkgImagequant" 1
  libdir="$ldprefix/libs/$libos/$libarch"
else
    echo "Using libdir: $libdir"
fi

if test ! -d "$libdir"; then
  terminate "Error: Path does not exist: $libdir" 1
fi

echo "Building library..."
CGO_LDFLAGS="-L$libdir $ldargs" go build && go install && terminate "Finished." 0 || terminate "Cancelled." 1

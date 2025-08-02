#!/bin/bash
export DYLD_LIBRARY_PATH=/opt/homebrew/lib
export CPATH=/opt/homebrew/Cellar/tomlplusplus/3.4.0/include
exec go test -v ./...

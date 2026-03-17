#!/bin/bash

set -e

command -v gtar &>/dev/null && { shopt -s expand_aliases; alias tar=gtar; }


ln_soft(){
    local target="$1"
    local link="$2"

# Detect Windows
    if [[ -n "$WINDIR" ]] || [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "msys2" ]]; then
        touch "${target}"
        powershell -Command "New-Item -ItemType SymbolicLink -Path '${link}' -Target '${target}' -Force"
    else
        ln -s "${target}" "${link}"
    fi
}
ln_hard(){
    local source="$1"
    local link="$2"

#Window detection
    if [[ -n "$WINDIR" ]] || [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "msys2" ]]; then
        powershell -Command "New-Item -ItemType HardLink -Path '${link}' -Value '${source}' -Force"
    else
        ln "${source}" "${link}"
    fi
}

TEST_DIR=$(mktemp -d /tmp/test-tardiff-XXXXXX)

create_orig () {
    DIR=$1

    mkdir -p $DIR
    pushd $DIR &> /dev/null

    mkdir data
    mkdir data/dir1
    mkdir data/dir2
    echo foo > data/dir1/foo.txt
    echo bar > data/dir1/bar.txt
    echo movedata > data/dir1/move.txt
    
    # Skip broken symlink on Windows: tar refuses to archive symlinks with non-existent targets
    if [[ -z "$WINDIR" ]] && [[ "$OSTYPE" != "msys" ]] && [[ "$OSTYPE" != "msys2" ]]; then
        ln_soft not-exist data/broken
    fi
    ln_soft foo.txt data/dir1/symlink
    ln_hard data/dir1/foo.txt data/dir1/hardlink


    echo "PART1" > data/sparse
    dd of=data/sparse if=/dev/null bs=1024k seek=1 count=1 &> /dev/null
    echo "PART2" >> data/sparse

    popd &> /dev/null
}

modify_orig () {
    DIR=$1
    SRC=$2

    mkdir -p $DIR
    # Extract old data
    tar xf $SRC -C $DIR
    pushd $DIR &> /dev/null

    # Modify it
    echo newdata > data/newfile
    mv data/dir1/move.txt data/dir2/move.txt

    echo bar >> data/dir1/bar.txt
    mv data/dir1/bar.txt data/dir1/bar.TXT # Rename we should pick up
    ln_hard data/dir1/foo.txt data/dir1/hardlink2

    popd &> /dev/null
}

compress_tar () {
    FILE=$1
    gzip -k $FILE
    bzip2 -k $FILE
}

create_tar () {
    FILE=$1
    DIR=$2
    tar cf $FILE --sparse -C $DIR data
    compress_tar $FILE
}

create_orig $TEST_DIR/orig
create_tar $TEST_DIR/orig.tar $TEST_DIR/orig

modify_orig $TEST_DIR/modified $TEST_DIR/orig.tar
create_tar $TEST_DIR/modified.tar $TEST_DIR/modified

echo Generating tardiff
./tar-diff $TEST_DIR/orig.tar.gz $TEST_DIR/modified.tar.bz2 $TEST_DIR/changes.tardiff

echo Applying tardiff
./tar-patch $TEST_DIR/changes.tardiff $TEST_DIR/orig $TEST_DIR/reconstructed.tar

echo Verifying reconstruction
cmp $TEST_DIR/reconstructed.tar $TEST_DIR/modified.tar

echo OK

cleanup () {
    rm -rf $TEST_DIR
}
trap cleanup EXIT

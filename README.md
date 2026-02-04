# tar-diff

`tar-diff` is a golang library and set of commandline tools to diff and patch tar files.

`pkg/tar-diff` and the `tar-diff` tool take two (optionally compressed) tar files and generate a single file representing the delta between them (a tardiff file).

`pkg/tar-patch` takes a tardiff file and the uncompressed contents (such as an extracted directory) of the first tar file and reconstructs (binary identically) the second tar file (uncompressed).

## Example
```
$ tar-diff old.tar.gz new.tar.gz delta.tardiff
$ tar xf old.tar.gz -C extracted/
$ tar-patch delta.tardiff extracted/ reconstructed.tar
$ zcat new.tar.gz | shasum
$ shasum reconstructed.tar
```


## Build requirements

- golang >= 1.25 (see [`go.mod`](go.mod))
- `make`
- `tar`
- `diffutils`, `bzip2`, `gzip` (for tests)

## Runtime dependencies

None. The built binaries are self-contained.


The main use case for `tar-diff` is for more efficient distribution of [OCI images](https://github.com/opencontainers/image-spec).
These images are typically transferred as compressed tar files, but the content is referred to and validated by the checksum of
the uncompressed content. This makes it possible to use an extracted earlier version of an image in combination with a tardiff file
to reconstruct and validate the current version of the image.

Delta compression is based on [bsdiff](http://www.daemonology.net/bsdiff/) and [zstd compression](https://facebook.github.io/zstd/).

The `tar-diff` file format is described in [file-format.md](file-format.md).

## License

`tar-diff` is licensed under the Apache License, Version 2.0. See
[LICENSE](LICENSE) for the full license text.
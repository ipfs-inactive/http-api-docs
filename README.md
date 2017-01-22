# ipfs-http-api-docs


[![](https://img.shields.io/badge/made%20by-Protocol%20Labs-blue.svg?style=flat-square)](http://ipn.io)
[![](https://img.shields.io/badge/project-IPFS-blue.svg?style=flat-square)](http://ipfs.io/)
[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)
[![Build Status](https://travis-ci.org/ipfs/ipfs-http-api-docs.svg?branch=master)](https://travis-ci.org/ipfs/ipfs-http-api-docs)

> A generator for go-ipfs API endpoints documentation.


## Table of Contents

- [Install](#install)
- [Usage](#usage)
- [Captain](#captain)
- [Contribute](#contribute)
- [License](#license)

## Install

In order to build this project, you need to download `go-ipfs` and install its gx dependencies:

```
> go get github.com/ipfs/go-ipfs
> cd $GOPATH/src/github.com/ipfs/go-ipfs
> make deps
```

Then you can:

```
> go install github.com/ipfs/ipfs-http-api-docs/ipfs-http-api-docs-md
```

Or alternatively checkout this repository and run:

```
> make install
```

## Usage

After installing you can run:

```
> ipfs-http-api-docs-md
```

This should spit out a Markdown document. This is exactly the api.md documentation at https://github.com/ipfs/website/blob/master/content/pages/docs/api.md , so you can redirect the output to just overwrite that file.

## Captain

This project is captained by @hsanjuan.

## Contribute

PRs accepted.

Small note: If editing the README, please conform to the [standard-readme](https://github.com/RichardLitt/standard-readme) specification.

## License

MIT © Hector Sanjuan

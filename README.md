# boshtags

boshtags is a [ctags][]-compatible tag generator for [Concourse][] pipelines.

## Installation

Install or update boshtags using the `go get` command:
```bash
go get -u github.com/luan/boshtags
```

## Usage

	boshtags [options] file(s)

	-L="": source file names are read from the specified file. If file is "-", input is read from standard in.
	-R=false: recurse into directories in the file list.
	-f="": write output to specified file. If file is "-", output is written to standard out.
	-silent=false: do not produce any output on error.
	-sort=true: sort tags.
	-tag-relative=false: file paths should be relative to the directory containing the tag file.
	-v=false: print version.

## Vim [Tagbar][] configuration

Put the following configuration in your vimrc:

```vim
let g:tagbar_type_concourse = {
    \ 'ctagstype' : 'concourse',
    \ 'kinds'     : [
        \ 'p:primitives',
        \ 't:resource_types',
        \ 'g:groups',
        \ 'r:resources',
        \ 'i:inputs',
        \ 'k:tasks',
        \ 'o:outputs',
        \ 'j:jobs',
    \ ],
    \ 'sro' : '.',
    \ 'kind2scope' : {
        \ 'p' : 'ptype',
        \ 'j' : 'stype'
    \ },
    \ 'scope2kind' : {
        \ 'ptype' : 'p',
        \ 'stype' : 'j'
    \ },
    \ 'ctagsbin'  : expand(bin_path),
    \ 'ctagsargs' : '-sort -silent'
\ }
```

### Vim+Tagbar Screenshot
![vim Tagbar boshtags](https://raw.githubusercontent.com/luan/boshtags/master/screenshots/screenshot-01.png)


[ctags]: http://ctags.sourceforge.net
[concourse]: http://concourse.ci
[tagbar]: http://majutsushi.github.com/tagbar/
[screenshot]: https://raw.githubusercontent.com/luan/boshtags/master/screenshots/screenshot-01.png

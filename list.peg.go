package main

const listGrammar = `prgm <- list+
list <- _?^ open atom_left+ close
atom_left <- atom _?^
atom <- number / name / list
name <- ~'[a-zA-Z/\-\*\+=><]+'
number <- ~'-?\d+\.?\d*'
open <- '('
close <- ')'
_ <- ~'\s+'`

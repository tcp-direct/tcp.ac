package main

import "github.com/prologic/bitcask"
import "os"

////////////// global declarations
// datastores
var imgDB *bitcask.Bitcask
var hashDB *bitcask.Bitcask
var keyDB *bitcask.Bitcask
var urlDB *bitcask.Bitcask
var txtDB *bitcask.Bitcask

// config directives
var debugBool bool
var baseUrl string
var webPort string
var webIP string
var dbDir string
var logDir string
var uidSize int
var keySize int

// utilitarian globals
var err error
var fn string
var s string
var i int
var f *os.File

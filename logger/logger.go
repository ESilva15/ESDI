package logger

import (
	"log"
	"os"
	"sync"
)

var l *log.Logger
var once sync.Once

func createLogger() {
  l = log.New(os.Stdout, "esdi", log.LstdFlags | log.Lshortfile)
}

func GetInstance() *log.Logger {
	once.Do(func() {
    createLogger() 
	})

  return l
}

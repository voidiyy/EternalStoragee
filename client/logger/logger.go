package logger

import (
	"log"
	"os"
)

type EtrnlLogger struct {
	Info func(msg string)
	Err  func(err error, msg string) error
	Msg  func(msg, remote string)
}

func NewEtrnlLogger() *EtrnlLogger {
	return &EtrnlLogger{
		Info: func(msg string) {
			l := log.New(os.Stdout, "[INFO]:", log.Ldate|log.Ltime|log.Lshortfile)
			l.Println(msg)
		},
		Err: func(err error, msg string) error {
			l := log.New(os.Stdout, "[ERROR]:", log.Ldate|log.Ltime|log.Lshortfile)
			l.Println(msg, err.Error())
			return err
		},
		Msg: func(msg, remote string) {
			l := log.New(os.Stdout, "[MSG]: ", log.Ldate)
			l.Println(msg + remote)
		},
	}
}

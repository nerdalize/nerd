package transferarchiver

import "io"

//Reporter describes how an archiver reports
type Reporter interface {
	StartArchivingProgress(label string, total int64) func(int64)
	StopArchivingProgress()
	StartUnarchivingProgress(label string, total int64, rr io.Reader) io.Reader
	StopUnarchivingProgress()
}

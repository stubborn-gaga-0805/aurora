package helpers

import (
	"github.com/cheggaaa/pb/v3"
	"os"
	"time"
)

func NewProgressBar(count int, msg string) (bar *pb.ProgressBar) {
	var progressTemp = `{{string . "prefix" | blue}} {{ bar . "[" "=" (cycle . "↖" "↗" "↘" "↙" ">" ">" ">") "-" "]"}} {{percent . | blue}} {{speed . | blue }}   {{string . "duration" | green}} {{etime . | green}} {{string . "end"}}`

	bar = pb.New(count)
	bar.SetTemplate(pb.ProgressBarTemplate(progressTemp)).
		Set("prefix", msg).
		Set("end", "\n").
		Set("duration", "duration:").
		SetRefreshRate(time.Second * 10).
		SetWidth(160).
		SetWriter(os.Stdout).
		Start()

	return bar
}

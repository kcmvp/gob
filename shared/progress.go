package shared

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/schollz/progressbar/v3"
)

func NewProgress() *progressbar.ProgressBar {
	spinner := []string{"←", "↑", "→", "↓"}
	return progressbar.NewOptions(-1,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetWidth(10),
		progressbar.OptionSpinnerCustom(lo.Map(spinner, func(item string, idx int) string {
			return fmt.Sprintf("[yellow]%s", item)
		})),
		progressbar.OptionOnCompletion(func() {
			fmt.Printf("\r%-50s\n", " ")
		}),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionEnableColorCodes(true),
	)
}

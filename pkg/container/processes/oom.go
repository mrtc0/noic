package processes

import (
	"fmt"
	"os"
)

func ApplyOOMScoreAdj(value int) error {
	oomScoreAdjPath := "/proc/self/oom_score_adj"

	f, err := os.OpenFile(oomScoreAdjPath, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%d", value))
	return err
}

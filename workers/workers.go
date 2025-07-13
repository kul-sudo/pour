package workers

import (
	"fmt"
	"os/exec"
	"os"
)

const SEGMENT_DURATION = 3
const SEGMENTS_DIR = "chunks"

func Segmentation() {
	err := os.Mkdir(SEGMENTS_DIR, 0755)
	if err != nil {
		fmt.Printf("failed to create dir for segments\n")
		// return
	}

	cmd := exec.Command("ffmpeg", "-i", "rtmp://localhost:1935/live/playpath", "-c", "copy", "-f", "segment", "-segment_time", fmt.Sprintf("%d", SEGMENT_DURATION), "-reset_timestamps", "1", "-segment_format_options", "movflags=frag_keyframe+empty_moov", SEGMENTS_DIR + "/%d.mp4")
	err = cmd.Run()
	if err != nil {
		fmt.Printf("failed to run segmentation process\n")
		os.Exit(0)
	}
}

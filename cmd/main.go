// this was a bash script that I converted to Go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run tronglerize.go <video_file> <output_folder>")
		return
	}

	videoFile := os.Args[1]
	outFolder := os.Args[2]
	cacheFolder := "/tmp/tronglerize"

	// Create cache folder
	fmt.Println("Creating cache folder")
	cacheDir := filepath.Join(cacheFolder, videoFile)
	os.MkdirAll(cacheDir, 0755)

	// Clear cache folder
	fmt.Println("Clearing cache folder")
	removeContents(cacheDir)

	// Get frame rate
	fmt.Println("Getting frame rate")
	frameRate := getFrameRate(videoFile)

	// Convert video to images
	fmt.Println("Converting video to images")
	convertToImages(videoFile, cacheDir, frameRate)

	// Process images
	fmt.Println("Processing images")
	processImages(cacheDir, outFolder)

	// Play video
	fmt.Println("Playing video")
	Play(outFolder)
}

func getFrameRate(videoFile string) string {
	cmd := exec.Command("ffmpeg", "-i", videoFile)
	output, _ := cmd.CombinedOutput()

	re := regexp.MustCompile(`, (\d+(?:\.\d+)?) fps`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		return matches[1]
	}
	return "24" // default frame rate
}

func convertToImages(videoFile, cacheDir, frameRate string) {
	outputPattern := filepath.Join(cacheDir, "frame%d.jpg")
	cmd := exec.Command("ffmpeg", "-i", videoFile, "-vf", "fps="+frameRate, outputPattern)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println(cmd.String())
	cmd.Run()
}

func processImages(cacheDir, outFolder string) {
	fmt.Println(filepath.Join(cacheDir, "*.jpg"))
	files, _ := filepath.Glob(filepath.Join(cacheDir, "*.jpg"))
	for _, file := range files {
		fmt.Print("Processing ", file, " ... ")
		baseName := filepath.Base(file)
		nameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
		outFile := filepath.Join(outFolder, nameWithoutExt+".bin")

		cmd := exec.Command("./triangle/triangle", "-pts", "5000", "-in", file, "-out", outFile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		fmt.Println(cmd.String())
		cmd.Run()
	}
}

func removeContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

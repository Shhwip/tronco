// this was a bash script that I converted to Go
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/shhwip/triangle/v2"
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
	err := processImages(cacheDir, outFolder)
	if err != nil {
		fmt.Println("Error processing images:", err)
		return
	}

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

func processImages(cacheDir, outFolder string) error {
	fmt.Println(filepath.Join(cacheDir, "*.jpg"))
	files, _ := filepath.Glob(filepath.Join(cacheDir, "*.jpg"))
	for _, file := range files {
		fmt.Print("Processing ", file, " ... ")
		baseName := filepath.Base(file)
		nameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
		outFile := filepath.Join(outFolder, nameWithoutExt+".bin")

		proc := &triangle.Processor{
			BlurRadius:      2,
			SobelThreshold:  10,
			PointsThreshold: 10,
			PointRate:       0.075,
			BlurFactor:      1,
			EdgeFactor:      6,
			MaxPoints:       5000,
			Wireframe:       0,
			Noise:           0,
			StrokeWidth:     1,
			IsStrokeSolid:   false,
			Grayscale:       false,
			ShowInBrowser:   false,
			BgColor:         "",
		}

		tri := &triangle.Image{
			Processor: *proc,
		}

		input, err := os.Open(file)
		if err != nil {
			return err
		}
		defer input.Close()

		output, err := os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer output.Close()

		img, err := tri.DecodeImage(input)
		if err != nil {
			return err
		}
		nodes, colors, err := tri.Save(img, *proc)
		if err != nil {
			return err
		}
		length := len(nodes)
		if length%6 != 0 {
			fmt.Printf("Invalid number of nodes: %d\n", length)
			return errors.New("invalid number of nodes")
		}

		outBytes := make([]byte, 0, length*2+len(colors)*3)
		outBytes = append(outBytes, byte(length>>8), byte(length&0xff))
		for _, node := range nodes {
			outBytes = append(outBytes, byte(node>>8))
			outBytes = append(outBytes, byte(node&0xff))
		}
		outBytes = append(outBytes, colors...)
		if len(outBytes) != length*2+2+length/2 {
			fmt.Printf("Invalid output length: %d\n", len(outBytes))
			return errors.New("invalid output length")
		}
		// Save the triangulated image as a binary file.
		_, err = output.Write(outBytes)
		if err != nil {
			return err
		}
	}
	return nil
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

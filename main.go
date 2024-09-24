package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/shhwip/triangle/v2"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run tronglerize.go <video_file> <output_folder>")
		return
	}

	videoFile := os.Args[1]
	outFolder := os.Args[2]

	f, err := os.Open(outFolder)
	if err != nil {
		fmt.Println("Error opening output folder:", err)
		return
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err != io.EOF {
		fmt.Println("Output folder is not empty")
		return
	}

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
	err = processImages(cacheDir, outFolder)
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
	files, _ := filepath.Glob(filepath.Join(cacheDir, "*.jpg"))

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 20) // current limit of 20 goroutines, its the best for my system
	errChan := make(chan error, len(files))

	for _, file := range files {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(file string) {
			defer wg.Done()
			defer func() { <-semaphore }()
			if err := processImage(file, outFolder); err != nil {
				errChan <- fmt.Errorf("error processing %s: %v", file, err)
			}
		}(file)
	}
	wg.Wait()
	close(errChan)

	var errs []string
	for err := range errChan {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors occurred during processing: %s", strings.Join(errs, "; "))
	}

	return nil
}

func processImage(file, outFolder string) error {
	fmt.Print("Processing ", file, " ... \n")
	baseName := filepath.Base(file)
	nameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	outFile := filepath.Join(outFolder, nameWithoutExt+".bin")

	// change these if you want to change how the triangles are generated
	//TODO: make these flags
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
		return fmt.Errorf("invalid number of nodes: %d", length)
	}

	outBytes := make([]byte, 0, length*2+len(colors)*3)
	outBytes = append(outBytes, byte(length>>8), byte(length&0xff))
	for _, node := range nodes {
		outBytes = append(outBytes, byte(node>>8))
		outBytes = append(outBytes, byte(node&0xff))
	}
	outBytes = append(outBytes, colors...)

	if len(outBytes) != length*2+2+length/2 {
		return fmt.Errorf("invalid output length: %d", len(outBytes))
	}

	_, err = output.Write(outBytes)
	if err != nil {
		return err
	}
	fmt.Printf("frame: %s has %d triangles\n", outFile, length/6)

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

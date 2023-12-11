package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

var (
	compressType = []string{"ebook", "screen", "printer", "prepress", "default"}
)

var (
	Version     string
	Commit      string
	BuildSource = "unknown"
)

func verifyOutput(def string, out string) string {
	prompt := &survey.Input{
		Message: fmt.Sprintf("输入文件名（默认：%s）：", def),
		Default: out,
	}
	var output string
	if err := survey.AskOne(prompt, &output); err != nil {
		log.Fatal(err)
	}

	if output == "" {
		output = out
	} else if !strings.HasSuffix(output, ".pdf") {
		output += ".pdf"
	}

	_, err := os.Stat(output)
	if err == nil {
		confirm := false
		confirmPrompt := &survey.Confirm{
			Message: fmt.Sprintf("输出文件已存在，想要覆盖吗？"),
		}
		if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
			log.Fatal(err)
		}
		if !confirm {
			return ""
		}
	}

	return output
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: cpdf [-c | --compress] [-m | --merge] [-v | --version]")
		return
	}

	compress := false
	merge := false
	version := false

	for _, arg := range os.Args[1:] {
		switch arg {
		case "-c", "--compress":
			compress = true
		case "-m", "--merge":
			merge = true
		case "-v", "--version":
			version = true
		}
	}

	if version {
		fmt.Printf("cpdf v%s@%s, %s, %s/%s\n", Version, Commit, BuildSource, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	fmt.Println("欢迎使用 pdf 工具")

	pdfFiles, err := filepath.Glob("*.pdf")
	if err != nil {
		log.Fatal(err)
	}

	if compress {
		compressPDF(pdfFiles)
	} else if merge {
		mergePDF(pdfFiles)
	} else {
		for {
			choice := getUserChoice()
			switch choice {
			case 1:
				mergePDF(pdfFiles)
			case 2:
				compressPDF(pdfFiles)
			case 3:
				fmt.Println("退出")
				return
			default:
				fmt.Println("无效选项")
			}
		}
	}
}

func getUserChoice() int {
	prompt := &survey.Select{
		Message: "你想要：",
		Options: []string{"合并 pdf", "压缩 pdf", "退出"},
	}
	var choice string
	if err := survey.AskOne(prompt, &choice); err != nil {
		log.Fatal(err)
	}

	switch choice {
	case "合并 pdf":
		return 1
	case "压缩 pdf":
		return 2
	case "退出":
		return 3
	default:
		return 0
	}
}

func mergePDF(pdfFiles []string) {
	choices := make([]string, len(pdfFiles))
	for i, file := range pdfFiles {
		choices[i] = file
	}

	prompt := &survey.MultiSelect{
		Message: "选择想要合并的 pdf",
		Options: choices,
	}
	var mergingFiles []string
	if err := survey.AskOne(prompt, &mergingFiles); err != nil {
		log.Fatal(err)
	}

	if len(mergingFiles) == 0 {
		fmt.Println("没有文件输入！")
		return
	}

	output := verifyOutput("merged.pdf", "merged.pdf")
	if output == "" {
		return
	}

	args := append([]string{"-q", "-dNOPAUSE", "-sDEVICE=pdfwrite", fmt.Sprintf("-sOutputFile=%s", output)}, mergingFiles...)
	cmd := exec.Command("gs", args...)
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error executing command: %v", err)
	}

	fmt.Println("合并中...")
	fmt.Printf("合并成功！输出文件：%s\n", output)
}

func compressPDF(pdfFiles []string) {
	choices := make([]string, len(pdfFiles))
	for i, file := range pdfFiles {
		choices[i] = file
	}

	prompt := &survey.Select{
		Message: "选择想要压缩的 PDF",
		Options: choices,
	}
	var choice string
	if err := survey.AskOne(prompt, &choice); err != nil {
		log.Fatal(err)
	}

	compressingFile := choice
	output := verifyOutput("原文件名.compressed.pdf", fmt.Sprintf("%s.compressed.pdf", strings.TrimSuffix(compressingFile, ".pdf")))
	if output == "" {
		return
	}

	cType := getUserChoiceForCompressType()
	initialSize, err := getFileSize(compressingFile)
	if err != nil {
		log.Fatal(err)
	}

	command := fmt.Sprintf("gs -q -sDEVICE=pdfwrite -dCompatibilityLevel=1.6 -dNumRenderingThreads=4 -dPDFSETTINGS=/%s -dNOPAUSE -dQUIET -dBATCH -sOutputFile=\"%s\" \"%s\" -c quit", cType, output, compressingFile)
	cmd := exec.Command("sh", "-c", command)
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Error executing command: %v", err)
	}

	fmt.Println("压缩中...")

	finalSize, err := getFileSize(output)
	if err != nil {
		log.Fatal(err)
	}

	compressionPercentage := (1 - float64(finalSize)/float64(initialSize)) * 100
	compressionResult := fmt.Sprintf("%.2f%%", compressionPercentage)
	if compressionPercentage > 0 {
		compressionResult = fmt.Sprintf("\033[31m-%.2f%%\033[0m", compressionPercentage)
	} else if compressionPercentage < 0 {
		compressionResult = fmt.Sprintf("\033[32m+%.2f%%\033[0m", -compressionPercentage)
	}
	fmt.Printf("%s : %s -> %s, %s\n", compressingFile, formatSize(initialSize), formatSize(finalSize), compressionResult)
	fmt.Println("压缩成功！")
}

func getUserChoiceForCompressType() string {
	prompt := &survey.Select{
		Message: "选择压缩方式：",
		Options: compressType,
	}
	var choice string
	if err := survey.AskOne(prompt, &choice); err != nil {
		log.Fatal(err)
	}
	return choice
}

func getFileSize(filename string) (int64, error) {
	file, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return file.Size(), nil
}

func formatSize(size int64) string {
	const (
		b  = 1
		kb = 1024 * b
		mb = 1024 * kb
	)
	switch {
	case size >= mb:
		return fmt.Sprintf("%.2f MB", float64(size)/mb)
	case size >= kb:
		return fmt.Sprintf("%.2f KB", float64(size)/kb)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

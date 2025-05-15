package main

import (
	"bufio"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
)

var verboseMode bool = false
var outputFormat string = "csv" // Varsayılan format

func isVerbose() bool {
	if verboseMode {
		return true
	}
	for _, arg := range os.Args {
		if arg == "--verbose" {
			return true
		}
	}
	if os.Getenv("DELVE") != "" || os.Getenv("GODEBUG") != "" {
		return true
	}
	return false
}

func printError(format string, v ...interface{}) {
	if isVerbose() {
		color.New(color.FgRed).Fprintf(os.Stderr, "HATA: "+format+"\n", v...)
	} else {
		os.Exit(1)
	}
}

func isIP(input string) bool {
	parts := strings.Split(input, ".")
	if len(parts) != 4 {
		return false
	}
	for _, p := range parts {
		if len(p) == 0 {
			return false
		}
	}
	return true
}

func writeOutput(filename string, results [][]string, format string) error {
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	switch format {
	case "csv":
		w := csv.NewWriter(outFile)
		w.Comma = ','
		for _, row := range results {
			if err := w.Write(row); err != nil {
				return err
			}
		}
		w.Flush()
	case "tsv":
		w := csv.NewWriter(outFile)
		w.Comma = '\t'
		for _, row := range results {
			if err := w.Write(row); err != nil {
				return err
			}
		}
		w.Flush()
	case "json":
		enc := json.NewEncoder(outFile)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	case "text":
		for _, row := range results {
			_, err := outFile.WriteString(strings.Join(row, " ") + "\n")
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("Bilinmeyen format: %s", format)
	}
	return nil
}

func printRow(row []string, format string) {
	switch format {
	case "csv":
		color.New(color.FgGreen).Println(strings.Join(row, ","))
	case "tsv":
		color.New(color.FgGreen).Println(strings.Join(row, "\t"))
	case "json":
		b, _ := json.Marshal(row)
		color.New(color.FgGreen).Println(string(b))
	case "text":
		color.New(color.FgGreen).Println(strings.Join(row, " "))
	}
}

func fetchAndSaveByIP(ip string, client *http.Client) {
	page := 1
	var results [][]string
	var headers []string
	firstPage := true

	for {
		url := fmt.Sprintf("https://rapiddns.io/s/%s?page=%d", ip, page)
		if isVerbose() {
			fmt.Printf("Sayfa %d: %s\n", page, url)
		}
		resp, err := client.Get(url)
		if err != nil {
			printError("Sayfa %d için HTTP isteği başarısız: %v", page, err)
			return
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			resp.Body.Close()
			printError("Sayfa %d için HTML ayrıştırma başarısız: %v", page, err)
			return
		}

		if firstPage {
			doc.Find("table.table.table-striped.table-bordered thead tr th").Each(func(i int, s *goquery.Selection) {
				headers = append(headers, strings.TrimSpace(s.Text()))
			})
			if len(headers) > 0 {
				results = append(results, headers)
			}
			firstPage = false
		}

		rows := doc.Find("table.table.table-striped.table-bordered tbody tr")
		if rows.Length() == 0 {
			if isVerbose() {
				fmt.Printf("Sayfa %d için No rows found\n", page)
			}
			resp.Body.Close()
			break
		}

		rows.Each(func(i int, s *goquery.Selection) {
			cols := s.Find("td")
			row := []string{}
			cols.Each(func(j int, td *goquery.Selection) {
				row = append(row, strings.TrimSpace(td.Text()))
			})
			if len(row) > 0 {
				results = append(results, row)
				printRow(row, outputFormat)
			}
		})
		resp.Body.Close()
		if doc.Find("ul.pagination li.page-item.active + li.page-item a").Length() == 0 {
			break
		}
		page++
	}

	outFileName := fmt.Sprintf("%s-rapiddns-ip.out", ip)
	err := writeOutput(outFileName, results, outputFormat)
	if err != nil {
		printError("Çıktı dosyası yazılamadı: %v", err)
		return
	}
	fmt.Printf("Tüm veriler %s dosyasına kaydedildi.\n", outFileName)
}

func fetchAndSaveByDomain(input string, client *http.Client) {
	// Bu fonksiyon için de istenirse benzer format desteği eklenebilir.
	url := fmt.Sprintf("https://rapiddns.io/subdomain/%s?full=1#result", input)
	resp, err := client.Get(url)
	if err != nil {
		printError("HTTP isteği başarısız: %v", err)
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		printError("HTML ayrıştırma başarısız: %v", err)
		return
	}

	tempFile, err := os.Create("domains-temp.txt")
	if err != nil {
		printError("Geçici dosya oluşturulamadı: %v", err)
		return
	}
	defer tempFile.Close()
	doc.Find("table.table.table-striped.table-bordered tbody tr").Each(func(i int, s *goquery.Selection) {
		td := s.Find("td").First()
		subdomain := strings.TrimSpace(td.Text())
		if subdomain != "" {
			_, _ = tempFile.WriteString(subdomain + "\n")
		}
	})
	tempFile.Close()

	linesSeen := make(map[string]struct{})
	outFileName := fmt.Sprintf("%s-rapiddns.out", input)
	outFile, err := os.Create(outFileName)
	if err != nil {
		printError("Çıktı dosyası oluşturulamadı: %v", err)
		return
	}
	defer outFile.Close()

	tempRead, err := os.Open("domains-temp.txt")
	if err != nil {
		printError("Geçici dosya okuma hatası: %v", err)
		return
	}
	defer tempRead.Close()

	scanner := bufio.NewScanner(tempRead)
	for scanner.Scan() {
		line := scanner.Text()
		if _, exists := linesSeen[line]; !exists && strings.TrimSpace(line) != "" {
			_, _ = outFile.WriteString(line + "\n")
			linesSeen[line] = struct{}{}
			color.New(color.FgGreen).Println(line)
		}
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		printError("Satır okuma hatası: %v", err)
	}

	os.Remove("domains-temp.txt")
}

func printHelp() {
	help := `\nRapidDNS Go Query Tool\n\nUsage:\n  rapiddnsquery <ip-or-domain> [--format=csv|tsv|json|text] [--verbose]\n  echo "1.2.3.4" | rapiddnsquery [options]\n\nOptions:\n  --format    Output format: csv (default), tsv, json, text\n  --verbose   Print detailed progress and errors to screen\n  -h, --help  Show this help message\n\nIf <ip-or-domain> is not supplied, the program reads the first line from stdin.\n\nExamples:\n  rapiddnsquery 8.8.8.8 --format=json --verbose\n  cat iplist.txt | rapiddnsquery --format=csv\n`
	fmt.Println(help)
}

func main() {
	formatFlag := flag.String("format", "csv", "Output format: csv, tsv, json, text")
	verboseFlag := flag.Bool("verbose", false, "Verbose output")
	helpFlag := flag.Bool("help", false, "Show help")
	flag.BoolVar(helpFlag, "h", false, "Show help (shorthand)")
	flag.Parse()
	outputFormat = strings.ToLower(*formatFlag)
	verboseMode = *verboseFlag

	if *helpFlag {
		printHelp()
		return
	}

	args := flag.Args()
	var input string
	if len(args) >= 1 && strings.TrimSpace(args[0]) != "" {
		input = args[0]
	} else {
		// stdin'den oku
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				input = line
				break
			}
		}
		if input == "" {
			fmt.Println("No input provided.\n")
			printHelp()
			os.Exit(1)
		}
	}

	if input == "" {
		fmt.Println("No input provided.\n")
		printHelp()
		os.Exit(1)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	if isIP(input) {
		fetchAndSaveByIP(input, client)
	} else {
		fetchAndSaveByDomain(input, client)
	}
}


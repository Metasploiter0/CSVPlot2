/*
This is the first draft of an internet financial stock price
provider made by Metasploiter0 (.metasploiter).
It works with the built-in http/template package
in addition to the go-echarts packages, logged
serverside by sirupsen/logrus.
Notes for improvements:
	-Add a configuration file in json, xml or else
	-Make the process of creating a graph possible from an api service
	-Improve the performance and recreatability of graphs as represented data
  -Create a version, that uses only temp files as data storage or use directories

*/

package main

import (
	"encoding/csv"
	"html/template"
	"io"
	"log"
	"net/http"
  "net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	//"time"
	//"fmt"

	//"github.com/tatsushid/go-fastping"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/sirupsen/logrus"
)

var (
	httpPort string = ":8080"
	apiKey   string = "OVNETQMES1QUWEUY"
	alive    bool   = true

	//callSrc string = Files[0].name
	wg sync.WaitGroup

	Files = []FileInfo{
		{name: "default", csvFile: "<nil>", htmlFile: "index.html", title: "<nil>"},
		{name: "aapl", csvFile: "datafiles/daily_AAPL.csv", htmlFile: "aapl.html", title: "Apple's daily stock price (AAPL)"},
		{name: "intl", csvFile: "datafiles/daily_INTL.csv", htmlFile: "intl.html", title: "Intel's daily stock price (INTL)"},
		{name: "ibm", csvFile: "datafiles/daily_IBM.csv", htmlFile: "ibm.html", title: "IBM's daily stock price (IBM)"},
    	{name: "dax", csvFile: "datafiles/daily_DAX.csv", htmlFile: "dax.html", title: "Daily stock price to the german stock index (DAX)"},
    	{name: "usd", csvFile: "datafiles/daily_USD.csv", htmlFile: "usd.html", title: "US Dollar's daily stock price"},
		{name: "overview", csvFile: "<nil>", htmlFile: "overview.html", title: "Overview of all available stock prices"},
		{name: "maintenance", csvFile: "<nil>", htmlFile: "maintenance.html", title: "<nil>"},
	}

	//callSrc string = Files[0].name
)

type FileInfo struct {
	name     string
	csvFile  string
	htmlFile string
	title    string
}

type dataStruct struct {
	date string
	data [4]float32
}

func getCSVfromAPI(instrument string, quiet ...bool) {
	var stdVerbose bool = true

	if len(quiet) != 0 {
		stdVerbose = quiet[0]
	}

	var srcUrl string = "https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol=" + instrument + "&apikey=" + apiKey + "&datatype=csv"

	var _ , err = url.Parse(srcUrl)
	if err != nil {
		if stdVerbose {
			logrus.Errorf("Error occured while parsing data file: %v", err)
		}
		return
	}

  
	var fileName string = "daily_"+strings.ToUpper(instrument)+".csv"
	file, err := os.Create("datafiles/"+fileName)
	if err != nil {
		log.Fatal(err)
	}
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	// Put content on file
	resp, err := client.Get(srcUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	size, err := io.Copy(file, resp.Body)

	defer file.Close()

	logrus.Infof("Downloaded a file %s with size %d", fileName, size)

}

func readCSV(filename string, quiet ...bool) ([][]string, error) {
	var stdVerbose bool = true

	if len(quiet) != 0 {
		stdVerbose = quiet[0]
	}
	file, err := os.Open(filename)
	if err != nil {
		if stdVerbose {
			logrus.Errorf("Error occured, while opening file %s: %v", filename, err)
		}
		return nil, err
	}
	defer file.Close()
	var reader *csv.Reader = csv.NewReader(file)
	if err != nil {
		if stdVerbose {
			logrus.Errorf("Error occurred, while loading file: %v%v", filename, err)
		}
		return nil, err
	}
	contents, err := reader.ReadAll()
	if err != nil {
		if stdVerbose {
			logrus.Errorf("Error occurred while reading all contents of file: %v", err)
		}
		return nil, err
	}
	return contents, nil
}

func returnFmt(array []string, index int, len0 int, quiet ...bool) ([]dataStruct, error) {
	var dataTemplate []dataStruct = make([]dataStruct, len0/6)
	var stdVerbose bool = true

	if len(quiet) != 0 {
		stdVerbose = quiet[0]
	}

	for i := index; i < len0; i += 6 {
		conv1, err := strconv.ParseFloat(array[i+4], 32)
		if err != nil {
			if stdVerbose {
				logrus.Errorf("Error occurred while converting data into float: %v", err)
			}
			return nil, err
		}
		conv2, err := strconv.ParseFloat(array[i+3], 32)
		if err != nil {
			if stdVerbose {
				logrus.Errorf("Error occurred while converting data into float: %v", err)
			}
			return nil, err
		}
		conv3, err := strconv.ParseFloat(array[i+2], 32)
		if err != nil {
			if stdVerbose {
				logrus.Errorf("Error occurred while converting data into float: %v", err)
			}
			return nil, err
		}
		conv4, err := strconv.ParseFloat(array[i+1], 32)
		if err != nil {
			if stdVerbose {
				logrus.Errorf("Error occurred while converting data into float: %v", err)
			}
			return nil, err
		}
		dataTemplate[i/6] = dataStruct{
			date: array[i+5],
			data: [4]float32{float32(conv1), float32(conv2), float32(conv3), float32(conv4)},
		}
	}

	return dataTemplate, nil
}

func createData(contents [][]string, quiet ...bool) ([]dataStruct, error) {
	var sepData []string = []string{}
	var stdVerbose bool = true

	if len(quiet) != 0 {
		stdVerbose = quiet[0]
	}

	for i0 := 0; i0 < len(contents); i0++ {
		for i1 := 0; i1 < 6; i1++ {
			sepData = append(sepData, contents[i0][i1])
		}
	}

	if stdVerbose {
		for i := 0; i < len(sepData); i++ {
			logrus.Printf(sepData[i])
		}
	}

	var sepDataR []string
	for i := len(sepData) - 1; i >= 0; i-- {
		sepDataR = append(sepDataR, sepData[i])
	}

	dataStruct, err := returnFmt(sepDataR, 0, 600)
	if err != nil {
		if stdVerbose {
			logrus.Errorf("Error occurred while converting data to format %s: %s", sepDataR, err)
		}
		return nil, err
	}
	logrus.Print(dataStruct)
	return dataStruct, nil
}

func plotKline(data []dataStruct, fileName string, graphTitle string, quiet ...bool) (components.Charter, error) {
	var kline, stdVerbose = charts.NewKLine(), true

	if len(quiet) != 0 {
		stdVerbose = quiet[0]
	}

	x := make([]string, 0)
	y := make([]opts.KlineData, 0)
	for i := 0; i < len(data); i++ {
		x = append(x, data[i].date)
		y = append(y, opts.KlineData{Value: data[i].data})
	}

	kline.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: graphTitle, //Add subtitle later !!!
		}),
		charts.WithXAxisOpts(opts.XAxis{
			SplitNumber: 20,
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Scale: true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Start:      50,
			End:        100,
			XAxisIndex: []int{0},
		}),
	)

	kline.SetXAxis(x).AddSeries("daily rate", y)

	var page *components.Page = components.NewPage()
	page.AddCharts(kline)
	file, err := os.Create(fileName)
	if err != nil {
		if stdVerbose {
			logrus.Errorf("Error occurred while creating page: %v", err)
		}
		return nil, err
	}

	page.Render(io.MultiWriter(file))
	return kline, nil
}

func defaultPage(w http.ResponseWriter, r *http.Request) {
	var fileName string = Files[0].htmlFile

	temp, err := template.ParseFiles(fileName)
	if err != nil {
		logrus.Errorf("Error parsing file %s: %v", fileName, err)
		return
	}
	err = temp.ExecuteTemplate(w, fileName, nil)
	if err != nil {
		logrus.Errorf("Error executing file %s: %v", fileName, err)
		return
	}
}

func aaplPage(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles(Files[1].htmlFile)
	if err != nil {
		logrus.Errorf("Error parsing file %s: %v", Files[1].htmlFile, err)
		return
	}
	err = temp.ExecuteTemplate(w, Files[1].htmlFile, nil)
	if err != nil {
		logrus.Errorf("Error executing file %s: %v", Files[1].htmlFile, err)
		return
	}
}

func intlPage(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles(Files[2].htmlFile)
	if err != nil {
		logrus.Errorf("Error parsing file %s: %v", Files[2].htmlFile, err)
		return
	}
	err = temp.ExecuteTemplate(w, Files[2].htmlFile, nil)
	if err != nil {
		logrus.Errorf("Error executing file %s: %v", Files[2].htmlFile, err)
		return
	}
}

func ibmPage(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles(Files[3].htmlFile)
	if err != nil {
		logrus.Errorf("Error parsing file %s: %v", Files[3].htmlFile, err)
		return
	}
	err = temp.ExecuteTemplate(w, Files[3].htmlFile, nil)
	if err != nil {
		logrus.Errorf("Error executing file %s: %v", Files[3].htmlFile, err)
		return
	}
}

func daxPage(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles(Files[4].htmlFile)
	if err != nil {
		logrus.Errorf("Error parsing file %s: %v", Files[4].htmlFile, err)
		return
	}
	err = temp.ExecuteTemplate(w, Files[4].htmlFile, nil)
	if err != nil {
		logrus.Errorf("Error executing file %s: %v", Files[4].htmlFile, err)
		return
	}
}

func usdPage(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles(Files[5].htmlFile)
	if err != nil {
		logrus.Errorf("Error parsing file %s: %v", Files[5].htmlFile, err)
		return
	}
	err = temp.ExecuteTemplate(w, Files[5].htmlFile, nil)
	if err != nil {
		logrus.Errorf("Error executing file %s: %v", Files[5].htmlFile, err)
		return
	}
}


func overviewPage(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles(Files[6].htmlFile)
	if err != nil {
		logrus.Errorf("Error parsing file %s: %v", Files[6].htmlFile, err)
		return
	}
	err = temp.ExecuteTemplate(w, Files[6].htmlFile, nil)
	if err != nil {
		logrus.Errorf("Error executing file %s: %v", Files[6].htmlFile, err)
		return
	}
}

func stopRepl(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles(Files[7].htmlFile)
	if err != nil {
		logrus.Errorf("Error parsing file %s: %v", Files[7].htmlFile, err)
		return
	}
	err = temp.ExecuteTemplate(w, Files[7].htmlFile, nil)
	if err != nil {
		logrus.Errorf("Error executing file %s: %v", Files[7].htmlFile, err)
		return
	}
	logrus.Infof("Repl will be shut down soon")
	alive = false
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path { //FIX THIS SWITCH-STATEMENT WITH FOR-LOOP AND IF-STATEMENTS LATER !!!
	case "/" + Files[1].name:
		aaplPage(w, r)
	case "/" + Files[2].name:
		intlPage(w, r)
	case "/" + Files[3].name:
		ibmPage(w, r)
	case "/" + Files[4].name:
		daxPage(w, r)
	case "/" + Files[5].name:
		usdPage(w, r)
	case "/" + Files[6].name:
		overviewPage(w, r)
	case "/" + Files[7].name:
		stopRepl(w, r)
	default:
		defaultPage(w, r) //Also, build a function, that works for all sites only with page as input
	}
}

func createKline(csvFile string, page string, title string) components.Charter {
	var contents, _ = readCSV(csvFile)
	var data, _ = createData(contents)
	kline, _ := plotKline(data, page, title)
	return kline
}

func createOverviewPage(pages []components.Charter, title string, quiet ...bool) error {
	var stdVerbose = true

	if len(quiet) != 0 {
		stdVerbose = quiet[0]
	}

	wg.Add(1)

	go func() {
		defer wg.Done()
		var page *components.Page = components.NewPage()
		for _, kline := range pages {
			page.AddCharts(kline)
		}

		file, err := os.Create(Files[6].htmlFile)
		if err != nil {
			if stdVerbose {
				logrus.Errorf("Error occurred while creating page: %v", err)
			}
			return
		}

		page.Render(io.MultiWriter(file))
	}()

	wg.Wait()
	return nil
}

func cleanTemp(path string, quiet ...bool) error {
  var stdVerbose = true

	if len(quiet) != 0 {
		stdVerbose = quiet[0]
	}
  var err error = os.RemoveAll(path)
  if err != nil {
    if stdVerbose {
      logrus.Errorf("Error occured while cleaning the remaining temp: %v", err)
    }
    return err
  }
  return nil
}

func main() {

  var err error = os.Mkdir("datafiles", 0750)
	if err != nil {
		logrus.Errorf("Error occured while creating directory for data files: %v", err)
	}
	wg.Add(5)

	go func() {
		defer wg.Done()
		getCSVfromAPI(Files[1].name)
	}()
	go func() {
		defer wg.Done()
		getCSVfromAPI(Files[2].name)
	}()
	go func() {
		defer wg.Done()
		getCSVfromAPI(Files[3].name)
	}()
	go func() {
		defer wg.Done()
		getCSVfromAPI(Files[4].name)
	}()
	go func() {
		defer wg.Done()
		getCSVfromAPI(Files[5].name)
	}()

	wg.Wait()

	var klines []components.Charter

	wg.Add(5)
  	go func() {
		defer wg.Done()
		klines = append(klines, createKline(Files[1].csvFile, Files[1].htmlFile, Files[1].title))
	}()
	go func() {
		defer wg.Done()
		klines = append(klines, createKline(Files[2].csvFile, Files[2].htmlFile, Files[2].title))
	}()
	go func() {
		defer wg.Done()
		klines = append(klines, createKline(Files[3].csvFile, Files[3].htmlFile, Files[3].title))
	}()
	go func() {
		defer wg.Done()
		klines = append(klines, createKline(Files[4].csvFile, Files[4].htmlFile, Files[4].title))
	}()
	go func() {
		defer wg.Done()
		klines = append(klines, createKline(Files[5].csvFile, Files[5].htmlFile, Files[5].title))
	}() //Add for loop for shortness
	wg.Wait()

	createOverviewPage(klines, Files[6].title)
	http.HandleFunc("/", requestHandler)
	err = http.ListenAndServe(httpPort, nil)
	if err != nil {
		logrus.Errorf("Error occurred while serving the page: %v", err)
		return
	}
  
  wg.Add(6)
  for i := 1; i < 6; i += 1 {
    go func() {
      defer wg.Done()
      cleanTemp(Files[i].htmlFile)
    }() 
  } 
  wg.Wait()
}
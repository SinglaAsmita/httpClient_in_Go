package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

func sendReq(parsedURL string, parsedPath string) (string, int, int64) {
	// Set the time out value
	timeout, _ := time.ParseDuration("5s")

	// Start the timer
	start := time.Now()
	dialer := net.Dialer{
		Timeout: timeout,
	}

	// Create a new connection
	tlsConn, err := tls.DialWithDialer(&dialer, "tcp", fmt.Sprintf("%s:https", parsedURL), nil)
	if err != nil {
		fmt.Printf("Could not connect with %s%s\nPlease verify if valid HTTPS endpoint", parsedURL, parsedPath)
		helpMessage()
		os.Exit(0)
	}
	defer tlsConn.Close()
	tlsConn.Write([]byte("GET " + parsedPath + " HTTP/1.0\r\nHost: " + parsedURL + "\r\n\r\n"))
	response, _ := ioutil.ReadAll(tlsConn)

	// Stop the timer
	end := time.Now()
	timeTaken := end.Sub(start)
	responseString := string(response)

	// Get the Http Response code
	code, err := strconv.Atoi(responseString[9:12])
	if err != nil {
		os.Exit(0)
	}
	tlsConn.Close()
	return responseString, code, timeTaken.Milliseconds()
}

func handleWorker(link string, profile bool, count int) {
	// Check if URL valid
	if isValidURL(link) {
		// Replace Http/Https from URL string
		httpPattern := regexp.MustCompile("^(http://|https://)")
		cleanedLink := httpPattern.ReplaceAllString(strings.ToLower(link), "")
		parsedURL := cleanedLink
		parsedPath := "/"

		// Get the index for URL path start
		slicedURL := strings.Index(cleanedLink, "/")
		if slicedURL != -1 {
			parsedPath = cleanedLink[slicedURL:]
			parsedURL = cleanedLink[:slicedURL]
		}

		// Declare Profiling arrays
		timeTakenArr := make([]int, count)
		HTTPResponseCodes := make([]int, count)
		HTTPResponses := make([]string, count)
		HTTPResponseLength := make([]int, count)
		totalTime := 0

		// Send HTTP request based on count requested
		for i := 0; i < count; i++ {
			HTTPRespBody, HTTPCode, timeTaken := sendReq(parsedURL, parsedPath)
			HTTPResponses[i] = HTTPRespBody
			HTTPResponseLength[i] = len(HTTPRespBody)
			HTTPResponseCodes[i] = HTTPCode
			timeTakenArr[i] = int(timeTaken)
			totalTime += int(timeTaken)
		}

		if !profile {
			// if not running in profile mode just print the HTTP response
			fmt.Println(HTTPResponses[0])
		} else {
			// Do Profiling
			sort.Ints(timeTakenArr)
			sort.Ints(HTTPResponseLength)
			errorCodes := getErrorCodes(HTTPResponseCodes)
			fmt.Printf("\n-----------------------------------------------\n")
			fmt.Println("[*] HTTP Profiler written by Asmita Singla")
			fmt.Printf("[*] Currently profiling %s\n", link)
			fmt.Printf("[*] Total no of HTTP request sent %d\n", count)
			fmt.Printf("[*] Fastest Response Time: %d milliseconds\n", timeTakenArr[0])
			fmt.Printf("[*] Slowest Response Time: %d milliseconds\n", timeTakenArr[len(timeTakenArr)-1])
			fmt.Printf("[*] Mean Response Time %d milliseconds\n", totalTime/len(timeTakenArr))
			fmt.Printf("[*] Median Request Time %d milliseconds\n", timeTakenArr[int(math.Floor(float64(len(timeTakenArr))/2.0))])
			fmt.Printf("[*] Successful Requests %d%%\n", ((count-len(errorCodes))/count)*100)
			fmt.Printf("[*] Error Codes Received %v\n", errorCodes)
			fmt.Printf("[*] Smallest Response Size %d\n", HTTPResponseLength[0])
			fmt.Printf("[*] Largest Response Size %d\n", HTTPResponseLength[len(HTTPResponseLength)-1])
			fmt.Println("------------------------------------------------")
		}
	} else {
		invalidMessage()
		helpMessage()
	}
}

func getErrorCodes(HTTPResponseCodes []int) []int {
	// Get all the error code received while profiling
	// For simplicity I am considering anything other than 200 as error response
	errorCodes := make([]int, 0)
	for _, HTTPResponseCode := range HTTPResponseCodes {
		if HTTPResponseCode != 200 {
			errorCodes = append(errorCodes, HTTPResponseCode)
		}
	}
	return errorCodes
}

func isValidURL(urlString string) bool {
	// regex to match valid url's, reference (https://stackoverflow.com/a/27379352)
	isValidURL, _ := regexp.MatchString(
		`(http(s)?:\/\/.)?(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}(\.[a-z]{2,6}|:[0-9]{3,4})\b([-a-zA-Z0-9@:%_\+.~#?&\/\/=]*)`,
		strings.ToLower(urlString))
	return isValidURL
}

func helpMessage() {
	fmt.Printf("\n----------------------HTTP profiler Usage----------------------\n")
	fmt.Println("To print HTTP response --> './profiler --url www.xyz.com'")
	fmt.Println("To profile a URL --> './profiler --url www.xyz.com --profile 10'")
	fmt.Println("Currently on HTTPS webpoints are supported")
	fmt.Println("To print this message again --> './profiler --help'")
}

func invalidMessage() {
	fmt.Printf("\nInvalid input provided, Please read the help below for details\n")
}

func main() {
	argsWithoutProg := os.Args[1:]
	totalArgs := len(argsWithoutProg)

	// get help message if argument length is 1 and has flag --help
	if totalArgs == 1 && argsWithoutProg[0] == "--help" {
		helpMessage()
	} else if totalArgs == 2 || totalArgs == 4 {
		// call handleWorker() and set profiling to false if argument length is 2 and has flag --url
		if totalArgs == 2 && argsWithoutProg[0] == "--url" {
			link := argsWithoutProg[1]
			count := 1
			profile := false
			handleWorker(link, profile, count)
		} else if totalArgs == 4 && argsWithoutProg[2] == "--profile" && argsWithoutProg[0] == "--url" {
			// call handleWorker() and set profiling to true if argument length is 4 and has both the flags --url and --count
			count, err := strconv.Atoi(argsWithoutProg[3])
			if err != nil {
				os.Exit(1)
			}
			profile := true
			link := argsWithoutProg[1]
			handleWorker(link, profile, count)
		} else {
			invalidMessage()
			helpMessage()
		}
	} else {
		invalidMessage()
		helpMessage()
	}
}

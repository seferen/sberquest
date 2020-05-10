package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

func getResultOfRegexp(str string, exp string) [][]string {
	re := regexp.MustCompile(exp)
	match := re.FindAllStringSubmatch(str, -1)
	return match
}

func getResultOfForm(body string) string {

	body = strings.ReplaceAll(body, "\n", "")
	form := getResultOfRegexp(body, "<form .+form>")[0][0]

	resultMap := make(map[string]string)

	if strings.Contains(form, "type=\"text\"") {

		textResult := getResultOfRegexp(form, "<p><input type=\\\"text\\\" name=\\\"([\\w|\\d]+)\\\"><\\/p>")
		for i := range textResult {

			resultMap[textResult[i][1]] = "test"
		}
		// строка для получения списка .//input[@type='textResult']/@name
	}

	if strings.Contains(form, "select") {
		selectRes := getResultOfRegexp(form, "<p><select name=\\\"([\\w|\\d]+)\\\">(.*?)</select>")
		for i := range selectRes {

			optionOfSelect := getResultOfRegexp(selectRes[i][2], "<option value=\"([\\w|\\d]+)\">([\\w|\\d]+)<")

			sort.Slice(optionOfSelect[:], func(i, j int) bool {
				return len(optionOfSelect[i][1]) > len(optionOfSelect[j][1])
			})
			resultMap[selectRes[i][1]] = optionOfSelect[0][1]
		}

	}

	if strings.Contains(form, "type=\"radio\"") {
		radioRes := getResultOfRegexp(form, "<p><input type=\"radio\" name=\"([\\w|\\d]+)\".*?</p>")
		for i := range radioRes {

			radioVal := getResultOfRegexp(radioRes[i][0],
				"<input type=\"radio\" name=\"([\\w|\\d]+)\" value=\"([\\w|\\d]+)\"")

			sort.Slice(radioVal, func(i, j int) bool {
				return len(radioVal[i][2]) > len(radioVal[j][2])
			})
			resultMap[radioRes[i][1]] = radioVal[0][2]
		}

	}

	var s string = ""
	for key, value := range resultMap {
		s = s + key + "=" + value + "&"

	}

	return s[:len(s)-1]
}
func sendInitRequest(host string) string {
	resp, err := http.Get(host)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var cook = strings.Split(resp.Header.Get("Set-Cookie"), "; ")[0]

	log.Println("Init query was finished")

	return cook

}

func sendRequest(method string, host string, cook string, numberOfQuestion string, body string, client *http.Client) (string, string) {
	req, err := http.NewRequest(method,
		host+"question/"+numberOfQuestion,
		strings.NewReader(body))

	req.Header.Add("Cookie", cook)

	if strings.EqualFold(method, "POST") {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	resultReq, err := ioutil.ReadAll(resp.Body)

	if strings.Contains(string(resultReq), "Test successfully passed") {
		return "", "exit"
	}

	//log.Println("Result of query " + numberOfQuestion + "________________________\n" + string(resultReq))

	if err != nil {
		log.Fatal(err)
	}
	questNumber := getResultOfRegexp(string(resultReq), "<h1>Question (\\d)</h1>")[0][1]
	var bodyReq = getResultOfForm(string(resultReq))

	log.Println("Query: " + method + " Number of question: " + numberOfQuestion)

	return bodyReq, questNumber

}

func main() {
	println("Start Script")

	client := &http.Client{}

	var host string = "http://test.youplace.net/"

	//обращение к первой странице и ожидание куки

	cook := sendInitRequest(host)
	//второй запрос и проверка куки

	result, number := sendRequest("GET", host, cook, "1", "", client)

	for !strings.EqualFold(number, "exit") {
		result, number = sendRequest("POST", host, cook, number, result, client)
	}

	log.Println("Script was finished")

	println(result + " " + number)

}

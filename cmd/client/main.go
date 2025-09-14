package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	endpoint := "http://localhost:8080/"
	data := url.Values{}

	fmt.Println("Введите длинный URL")

	reader := bufio.NewReader(os.Stdin)
	long, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Ошибка чтения ввода: %v\n", err)
		return
	}
	data.Set("url", long)

	client := &http.Client{}

	request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		panic(err)
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Статус-код: %d\n", response.StatusCode)
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Ошибка чтения ответа: %v\n", err)
		return
	}

	fmt.Println(string(body))
}

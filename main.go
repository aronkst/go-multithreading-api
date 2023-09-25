package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Cep struct {
	ViaCep string
	ApiCep string
}

type ViaCep struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

type ApiCep struct {
	Uf          string `json:"uf"`
	Cidade      string `json:"cidade"`
	Bairro      string `json:"bairro"`
	Logradouro  string `json:"logradouro"`
	Cep         string `json:"cep"`
	Complemento string `json:"complemento"`
	Nome        string `json:"nome"`
	Status      string `json:"status"`
	Tipo        string `json:"tipo"`
	CodigoIbge  string `json:"codigo_ibge"`
}

func main() {
	fmt.Print("Digite o CEP: ")
	var input string
	fmt.Scanln(&input)

	cep, err := formatCep(input)
	if err != nil {
		panic(err)
	}

	viaCepURL := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep.ViaCep)
	apiCepURL := fmt.Sprintf("https://cdn.apicep.com/file/apicep/%s.json", cep.ApiCep)

	chViaCep := make(chan string)
	chApiCep := make(chan string)

	go getViaCep(viaCepURL, chViaCep)
	go getApiCep(apiCepURL, chApiCep)

	select {
	case result := <-chViaCep:
		fmt.Println(result)
		return
	case result := <-chApiCep:
		fmt.Println(result)
		return
	case <-time.After(1 * time.Second):
		fmt.Println("Erro de Timeout. As duas API demoraram para responder.")
	}
}

func formatCep(input string) (Cep, error) {
	if len(input) == 9 && strings.Contains(input, "-") {
		viaCep := strings.Replace(input, "-", "", 1)
		return Cep{ViaCep: viaCep, ApiCep: input}, nil
	} else if len(input) == 8 && !strings.Contains(input, "-") {
		apiCep := input[:5] + "-" + input[5:]
		return Cep{ViaCep: input, ApiCep: apiCep}, nil
	} else {
		return Cep{}, errors.New("CEP inválido")
	}
}

func getViaCep(url string, ch chan<- string) {
	startTime := time.Now()

	body, err := fetchAPI(url)
	if err != nil {
		ch <- err.Error()

		return
	}

	var viaCep ViaCep
	if err := json.Unmarshal(body, &viaCep); err != nil {
		ch <- fmt.Sprintf("Erro ao processar JSON da API %s", url)

		return
	}

	jsonFormated, err := json.MarshalIndent(viaCep, "", "    ")
	if err != nil {
		ch <- fmt.Sprintf("Erro ao formatar o JSON da API %s", url)

		return
	}

	processingTime := time.Since(startTime)
	ch <- fmt.Sprintf("Tempo de resposta da API %s: %s\n%s", url, processingTime, jsonFormated)
}

func getApiCep(url string, ch chan<- string) {
	startTime := time.Now()

	body, err := fetchAPI(url)
	if err != nil {
		ch <- err.Error()

		return
	}

	var apiCep ApiCep
	if err := json.Unmarshal(body, &apiCep); err != nil {
		ch <- fmt.Sprintf("Erro ao processar JSON da API %s", url)

		return
	}

	jsonFormated, err := json.MarshalIndent(apiCep, "", "    ")
	if err != nil {
		ch <- fmt.Sprintf("Erro ao formatar o JSON da API %s", url)

		return
	}

	processingTime := time.Since(startTime)
	ch <- fmt.Sprintf("Tempo de resposta da API %s: %s\n%s", url, processingTime, jsonFormated)
}

func fetchAPI(url string) ([]byte, error) {
	http := http.Client{}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Erro ao fazer a request da API %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Status Code != 200 da API %s", url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

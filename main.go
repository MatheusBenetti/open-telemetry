package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type ViaCEPBody struct {
	Cep string `json:"cep"`
}

type ViaCEP struct {
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
	Erro        bool   `json:"erro"`
}

type WeatherResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_c"`
	TempF float64 `json:"temp_f"`
	TempK float64 `json:"temp_k"`
}

type Current struct {
	WeatherResponse `json:"current"`
}

func main() {
	http.HandleFunc("/getTemperature", func(w http.ResponseWriter, r *http.Request) {
		var data ViaCEPBody
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid request body"))
			return
		}
		cep := data.Cep

		log.Println(cep)

		if len(cep) != 8 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("invalid zipcode"))
			return
		}

		viaCEP, err := fetchViaCep(cep)
		log.Println(viaCEP)

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("can not found zipcode"))
			return
		}

		if viaCEP.Erro {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("can not found zipcode"))
			return
		}

		removeSpaces := strings.ReplaceAll(viaCEP.Localidade, " ", "-")

		weather, err := fetchWeatherAPI(removeSpaces)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error getting weather data"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(weather.WeatherResponse)
	})

	log.Println("Server listening on port :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func fetchViaCep(cep string) (*ViaCEP, error) {
	req, err := http.Get("http://viacep.com.br/ws/" + cep + "/json/")

	if err != nil {
		return nil, fmt.Errorf("failed to make request to ViaCEP API: %v", err)
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var data ViaCEP
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return &data, nil
}

func fetchWeatherAPI(location string) (*Current, error) {
	req, err := http.Get("http://api.weatherapi.com/v1/current.json?q=" + location + "&key=50dbab8a6094453b8d4214401242301")

	if err != nil {
		return nil, fmt.Errorf("failed to make request to Weather API: %v", err)
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var data Current
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	var fahrenheit float64 = data.WeatherResponse.TempC
	var kelvin float64 = data.WeatherResponse.TempC

	formatString := strings.ReplaceAll(location, "-", " ")

	data.WeatherResponse.City = formatString
	data.WeatherResponse.TempF = celsiusToFahrenheit(fahrenheit)
	data.WeatherResponse.TempK = celsiusToKelvin(kelvin)
	return &data, nil
}

func celsiusToFahrenheit(celsius float64) float64 {
	return (celsius * 1.8) + 32
}

func celsiusToKelvin(celsius float64) float64 {
	return celsius + 273
}

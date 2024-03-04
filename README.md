# Open Telemetry

Digite o CEP desejado para receber a temperatura do momento em graus Celsius, Fahrenheit e Kelvin

## Conteúdo

- [Ambiente de Desenvolvimento](#developer)

## Ambiente de Desenvolvimento

Para rodar o projeto em ambiente de desenvolvimento, utilize para criar o container:
```
docker compose up --build
```
ou então utilize docker-compose up --build, dependendo da versão, após gerar a imagem, pode iniciar as próximas vezes com docker compose up.

Depois para acessar o bash, abra outro terminal e digite:
```
docker compose exec web bash
```
E então rode o comando:
```
go run main.go
```
Após isso é só fazer uma requisição no terminal com:
```
curl http://localhost:8080/getTemperature?cep=95670084
```
Ou via Postman/Insomnia


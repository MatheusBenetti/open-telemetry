# Open Telemetry

Digite o CEP desejado para receber a temperatura do momento em graus Celsius, Fahrenheit e Kelvin

## Ambiente de Desenvolvimento

Para rodar o projeto em ambiente de desenvolvimento, digite o seguinte código no terminal:
```
make prepare
```
Após a execução desse comando, digite para preparar os containers:
```
make run
```
Após isso realize requisições POST no Insomnia/Postman na URL http://localhost:8080/getCep com o seguinte body, podendo alterar o CEP para o desejado:
```
{
  "cep": "95670084"
}
```
## Zipkin

 - Para acessar o Zipkin, abra o seu navegador e digite a URL http://localhost:9411/;
 - Clique no botão "RUN QUERY" e será possível ver os resultados.


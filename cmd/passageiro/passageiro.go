package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	. "vaijunto/shared"
)

func main() {
	endereco := os.Getenv("endereco")
	if endereco == "" {
		fmt.Print("\nCade o endereco?!?!\n")
		return
	}

	conn, err := net.Dial("tcp", endereco)
	if err != nil {
		log.Fatalf("Erro ao conectar ao servidor: %v", err)
	}
	defer conn.Close()

	scanner := bufio.NewScanner(os.Stdin)
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	fmt.Println("=== CLIENTE PASSAGEIRO ===")

	var conectado bool = true
	for conectado {
		fmt.Println("\nEscolha uma opção: ")
		fmt.Println("1. Login")
		fmt.Println("2. Nova conta")
		fmt.Println("3. Encerrar programa")

		scanner.Scan()
		input := scanner.Text()

		switch input {
		// case 1 e 2, diferem em apenas duas linha, como posso melhorar isso?
		case "1": //login
			{
				fmt.Print("Digite o nome de Usuário:")
				scanner.Scan()
				nome := scanner.Text()

				fmt.Print("Digite a senha:")
				scanner.Scan()
				senha := scanner.Text()

				var req Request = Request{
					Acao:  "LOGIN_PASSAGEIRO",
					Nome:  nome,
					Senha: senha,
				}

				//escreve request na stream da conexão tcp
				_ = encoder.Encode(req)

				//le resposta da conexão
				var res Resposta
				_ = decoder.Decode(&res)

				//exibe resposta para o cliente
				fmt.Printf("\nServidor diz: [%s] %s\n", res.Status, res.Message)

				if res.Status == "SUCESSO" {
					LoggedIn(encoder, decoder, scanner)
				}
			}
		case "2": //cadastro
			{
				fmt.Println("Digite o nome de Usuário:")
				scanner.Scan()
				nome := scanner.Text()

				fmt.Println("Digite a senha:")
				scanner.Scan()
				senha := scanner.Text()

				var req Request = Request{
					Acao:  "CADASTRO_PASSAGEIRO",
					Nome:  nome,
					Senha: senha,
				}

				_ = encoder.Encode(req)

				var res Resposta
				_ = decoder.Decode(&res)

				fmt.Printf("\nServidor diz: [%s] %s\n", res.Status, res.Message)

				if res.Status == "SUCESSO" {
					fmt.Println("\nFazendo login Automático após cadastro...")
					LoggedIn(encoder, decoder, scanner)
				}
			}
		case "3":
			var req Request = Request{
				Acao: "DESCONECTAR",
			}

			_ = encoder.Encode(req)

			var res Resposta
			_ = decoder.Decode(&res)

			fmt.Printf("\nServidor diz: [%s] %s\n", res.Status, res.Message)

			fmt.Print("\nPrograma cliente encerrado.")

			conectado = false

		default:
			fmt.Print("\nOpção não reconhecido.")
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "Erro ao ler a entrada:", err)
		}

	}
}

/*
tratamento de opcoes para usario logado
*/
func LoggedIn(encoder *json.Encoder, decoder *json.Decoder, scanner *bufio.Scanner) {

	for {
		fmt.Println("\nEscolha uma opção: ")
		fmt.Println("1. Buscar Rota de viagem")
		fmt.Println("2. Remover uma reserva")
		fmt.Println("3. Listar reservas")
		fmt.Println("4. Ver notificações")
		fmt.Println("5. Log Out")

		scanner.Scan()
		input := scanner.Text()

		switch input {
		case "1": // Buscar e realizar reserva
			ImprimirCidades()

			rota, cancelado := ConstruirRota(scanner)
			if cancelado {
				fmt.Println("Operação cancelada.")
				continue
			}

			req := Request{
				Acao:    "BUSCAR_E_RESERVAR",
				Origem:  rota[0],
				Destino: rota[1],
			}

			if err := encoder.Encode(req); err != nil {
				log.Printf("Erro ao enviar requisição: %v. Operação encerrada", err)
				break
			}

			var res Resposta
			if err := decoder.Decode(&res); err != nil {
				log.Printf("Erro ao ler resposta: %v. Operação encerrada", err)
				break
			}

			fmt.Printf("\nServidor diz: \n[%s] %s\n", res.Status, res.Message)

			if res.Status != "SUCESSO" {
				fmt.Println("Infelizmente nenhuma rota foi encontrada. Operação encerrada")
				continue
			}

			//aqui está dentro do loop, caso haja trechos e vagas
			// loop de navegação pelas rotas encontradas
			for {

				//essa função enumera as viagens enviadas,
				// não é muito adequado, aqui que só tem uma
				ImprimirReservas(res.Viagens)

				fmt.Println("\nEscolha uma ação:")
				fmt.Println("1. Confirmar e reservar esta vaga")
				fmt.Println("2. Ver próxima opção de viagem")
				fmt.Println("3. Cancelar busca")
				fmt.Print(">> ")

				if !scanner.Scan() {
					break
				}
				input := strings.TrimSpace(scanner.Text())

				var acaoReq string
				switch input {
				case "1":
					acaoReq = "CONFIRMAR_ROTA"
				case "2":
					acaoReq = "PROXIMA_ROTA"
				case "3":
					acaoReq = "CANCELAR"
				default:
					fmt.Println("\nOpção inválida! Tente novamente.")
					continue
				}

				req = Request{Acao: acaoReq}
				if err := encoder.Encode(req); err != nil {
					log.Printf("Erro ao enviar requisição: %v", err)
					break
				}

				if err := decoder.Decode(&res); err != nil {
					log.Printf("Erro ao ler resposta: %v", err)
					break
				}

				fmt.Printf("\nServidor diz: \n[%s] %s\n", res.Status, res.Message)

				if acaoReq == "CONFIRMAR_ROTA" || acaoReq == "CANCELAR" || res.Status != "SUCESSO" {
					break
				}
			}

		case "2": //remover reserva
			{
				//listando para o passageiro
				var req = Request{
					Acao: "LISTAR_RESERVAS",
				}
				_ = encoder.Encode(req)
				var res Resposta
				_ = decoder.Decode(&res)
				fmt.Printf("\nServidor diz: [%s] %s\n", res.Status, res.Message)
				ImprimirReservas(res.Viagens)

				//
				fmt.Println("Digite 'c' para cancelar.")
				fmt.Println("Qual reserva deseja remover?:")
				num, cancelado := LerInt(scanner)
				if cancelado {
					fmt.Println("Operação cancelada")
					continue
				}
				indice := num - 1
				if indice < 0 || indice >= len(res.Viagens) {
					fmt.Println("Indice de viagem invalido.")
					fmt.Println("Operação encerrada.")
					continue
				}

				remover := res.Viagens[indice].ID

				req = Request{
					Acao: "REMOVER_RESERVA",
					ID:   remover,
				}

				_ = encoder.Encode(req)
				_ = decoder.Decode(&res)
				fmt.Printf("\nServidor diz: [%s] %s\n", res.Status, res.Message)
				fmt.Printf("Reserva de ID %s removida", remover)

			}
		case "3": //listar reservas
			{
				var req = Request{
					Acao: "LISTAR_RESERVAS",
				}

				_ = encoder.Encode(req)

				var res Resposta
				_ = decoder.Decode(&res)

				fmt.Printf("\nServidor diz: [%s] %s\n", res.Status, res.Message)

				ImprimirReservas(res.Viagens)
			}
		case "4":
			{
				var req = Request{
					Acao: "GET_NOTI",
				}

				_ = encoder.Encode(req)

				var res Resposta
				_ = decoder.Decode(&res)

				fmt.Printf("\nServidor diz: [%s] %s\n", res.Status, res.Message)

				ImprimirNotificacoes(res.Dados)

			}
		case "5": //logout
			{
				var req = Request{
					Acao: "LOG_OUT",
				}

				_ = encoder.Encode(req)

				var res Resposta
				_ = decoder.Decode(&res)

				fmt.Printf("\nServidor diz: [%s] %s\n", res.Status, res.Message)

				return
			}
		default:
		}
	}
}

/*
Constroi a rota desejada pelo usario
*/
func ConstruirRota(scanner *bufio.Scanner) ([]string, bool) {
	var paradas []string

	fmt.Println("\n--- Cadastro de Rota ---")
	fmt.Println("Digite 'c' para cancelar.")

	for {

		if len(paradas) == 0 {
			fmt.Printf("\nInforme o código da cidade de origem: ")
		} else {
			fmt.Printf("\nInforme o código da cidade de destino: ")
		}

		num, cancelado := LerInt(scanner)

		if cancelado {
			return nil, true
		}

		cidade := Cidade(num).GetCidade()
		if cidade == "INVALIDA" {
			fmt.Println("Erro: O código informado não corresponde a nenhuma cidade válida.")
			continue
		}

		if (len(paradas) > 0) && (cidade == paradas[len(paradas)-1]) {
			fmt.Println("Erro: A cidade de destino não pode ser igual a de origem.")
			continue
		}

		paradas = append(paradas, cidade)
		fmt.Printf("Adicionado: %s\n", cidade)

		if len(paradas) == 2 {
			break
		}
	}

	return paradas, false
}

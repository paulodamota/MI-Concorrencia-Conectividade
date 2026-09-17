package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	. "vaijunto/shared"
)

func main() {
	fmt.Println("=== INICIANDO TESTE DE CONCORRÊNCIA ===")

	const totalClientes = 10
	var wg sync.WaitGroup
	wg.Add(totalClientes)

	nomeUsuario := "teste_concorrente"
	senhaUsuario := "123456"

	cadastrarConta(nomeUsuario, senhaUsuario)

	// Dispara várias goroutines simultâneas
	for i := 1; i <= totalClientes; i++ {
		go func(id int) {
			defer wg.Done()
			
			conn, err := net.Dial("tcp", "172.16.230.17:6742")
			if err != nil {
				fmt.Printf("[Cliente %d] Erro ao conectar ao servidor: %v\n", id, err)
				return
			}
			defer conn.Close()

			encoder := json.NewEncoder(conn)
			decoder := json.NewDecoder(conn)

			// Envia requisição de login
			req := Request{
				Acao:  "LOGIN_PASSAGEIRO",
				Nome:  nomeUsuario,
				Senha: senhaUsuario,
			}

			if err := encoder.Encode(req); err != nil {
				fmt.Printf("[Cliente %d] Erro ao enviar requisição: %v\n", id, err)
				return
			}

			var res Resposta
			if err := decoder.Decode(&res); err != nil {
				fmt.Printf("[Cliente %d] Erro ao ler resposta: %v\n", id, err)
				return
			}

			fmt.Printf("[Cliente %d] Resposta recebida -> Status: %s | Mensagem: %s\n", id, res.Status, res.Message)
		}(i)
	}

	wg.Wait()
	fmt.Println("=== TESTE DE CONCORRÊNCIA FINALIZADO ===")
}

func cadastrarConta(nome, senha string) {
	conn, err := net.Dial("tcp", "172.16.230.17:6742")
	if err != nil {
		return
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	_ = encoder.Encode(Request{
		Acao:  "CADASTRO_PASSAGEIRO",
		Nome:  nome,
		Senha: senha,
	})
	var res Resposta
	_ = decoder.Decode(&res)
}
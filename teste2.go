package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	. "vaijunto/shared"
)

func main() {
	fmt.Println("=== INICIANDO TESTE DE OVERBOOKING (CONCORRÊNCIA DE RESERVAS) ===")

	motorista := "moto_teste"
	senhaMoto := "123"
	origem := "Centro"
	destino := "Praia"

	cadastrarMotoristaErota(motorista, senhaMoto, origem, destino)

	const totalPassageiros = 10
	var wg sync.WaitGroup
	wg.Add(totalPassageiros)

	var sucessos int
	var falhas int
	var mu sync.Mutex

	for i := 1; i <= totalPassageiros; i++ {
		go func(id int) {
			defer wg.Done()

			nomePassageiro := fmt.Sprintf("pass_%d", id)
			
			conn, err := net.Dial("tcp", "172.16.230.17:6742")
			if err != nil {
				fmt.Printf("[Passageiro %d] Erro ao conectar: %v\n", id, err)
				return
			}
			defer conn.Close()

			encoder := json.NewEncoder(conn)
			decoder := json.NewDecoder(conn)

			_ = encoder.Encode(Request{Acao: "CADASTRO_PASSAGEIRO", Nome: nomePassageiro,Senha: "123"})
			var res Resposta
			_ = decoder.Decode(&res)

			_ = encoder.Encode(Request{Acao: "BUSCAR_E_RESERVAR", Nome: nomePassageiro, Origem: origem, Destino: destino})
			_ = decoder.Decode(&res)

			if res.Status == "SUCESSO" {
				_ = encoder.Encode(Request{Acao: "CONFIRMAR_ROTA", Nome: nomePassageiro})
				_ = decoder.Decode(&res)

				mu.Lock()
				if res.Status == "SUCESSO" {
					sucessos++
					fmt.Printf("[Passageiro %d] -> CONSEGUIU reservar a vaga!\n", id)
				} else {
					falhas++
					fmt.Printf("[Passageiro %d] -> Tentou confirmar, mas falhou: %s\n", id, res.Message)
				}
				mu.Unlock()
			} else {
				mu.Lock()
				falhas++
				fmt.Printf("[Passageiro %d] -> Não encontrou vagas disponíveis.\n", id)
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("\n=== RESULTADO DO TESTE DE OVERBOOKING ===")
	fmt.Printf("Total de reservas bem-sucedidas: %d\n", sucessos)
	fmt.Printf("Total de reservas bloqueadas/falhas: %d\n", falhas)
	fmt.Println("==========================================")
}

func cadastrarMotoristaErota(motorista, senha, origem, destino string) {
	conn, err := net.Dial("tcp", "172.16.230.17:6742")
	if err != nil {
		return
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	_ = encoder.Encode(Request{Acao: "CADASTRO_MOTORISTA", Nome: motorista, Senha: senha})
	var res Resposta
	_ = decoder.Decode(&res)

	trechos := []TrechoData{
		{Origem: origem, Destino: destino, Vagas: 1, Preco: 10.0}, 
	}

	_ = encoder.Encode(Request{Acao: "CADASTRAR_ROTA", Nome: motorista, Trechos: trechos})
	_ = decoder.Decode(&res)
	fmt.Println("-> Motorista cadastrado e rota de 1 vaga criada com sucesso.")
}
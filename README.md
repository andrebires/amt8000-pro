# Console LAN para AMT 8000 Pro

Aplicacao web local para consultar e, futuramente, programar centrais de alarme
Intelbras AMT 8000 Pro pela rede local.

Este projeto nao e oficial da Intelbras. O objetivo e oferecer uma alternativa
multiplataforma ao AMT Remoto / Programador AMT 8000 para uso em Mac, Windows,
Linux, celulares e tablets conectados na mesma rede local.

## Estado atual

- Login pelo navegador informando IP da central, porta e senha de acesso remoto.
- Conexao local via Ethernet usando ISECNet v2 na porta TCP `9009`.
- Painel web com status basico da central.
- Parser inicial para modelo, versao de firmware, particoes, setores, sirene,
  tamper e bateria.
- Backlog de paridade com AMT Remoto e Programador AMT 8000 em
  `docs/backlog/amt-remoto-parity.md`.

Funcionalidades de escrita, como armar/desarmar, bypass, PGM e programacao da
central, ainda dependem de evidencia de protocolo e testes em painel real.

## Requisitos

- Go 1.26 ou superior
- Central AMT 8000 Pro acessivel a partir do computador que roda este servico
- Senha de acesso remoto/download da central

## Configuracao

O formulario de login no navegador pede:

- IP local da central
- Porta, normalmente `9009`
- Senha remota/download

As credenciais ficam em um cookie de sessao `HttpOnly` no navegador local e nao
sao gravadas no repositorio.

Para mudar o endereco HTTP do servidor, copie `.env.example` para `.env`:

```sh
AMT_HTTP_ADDR=:8080
```

Nao commite arquivos `.env`, senhas, capturas de pacote ou dados reais da sua
instalacao.

## Como rodar

```sh
go run ./cmd/amt8000-pro
```

Depois acesse:

```text
http://localhost:8080
```

## Testes

Testes automatizados:

```sh
go test ./...
```

Smoke test contra uma central real:

```sh
AMT_HOST=192.168.1.50 AMT_PASSWORD=123456 scripts/production-status-test.sh
```

O script gera um relatorio Markdown em `docs/test-runs/`.

## Politica de seguranca

- Recursos somente leitura podem ser testados diretamente contra a central.
- Recursos de controle, como armar/desarmar e PGM, exigem checklist manual de
  seguranca antes da implementacao.
- Escritas de configuracao exigem captura/evidencia de protocolo, backup previo
  e verificacao de leitura apos escrita.
- Atualizacao de firmware, reset de fabrica, desbloqueio de memoria e comandos
  destrutivos estao fora de escopo ate pedido explicito.

## Referencias

- Manual AMT Remoto:
  `https://backend.intelbras.com/sites/default/files/2024-09/Manual_AMT_Remoto_01-21_site%20-%20Arquivo%20Final.pdf`
- Manual Programador AMT 8000:
  `https://backend.intelbras.com/sites/default/files/2021-12/Manual_programador_AMT_8000_01-21_site.pdf`

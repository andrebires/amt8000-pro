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
- Aba de eventos com download manual da central, filtros locais, exportacao
  CSV/JSON e mapeamento dos codigos de evento ja observados.
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

O servidor nao mantem uma sessao persistente com a central. Cada atualizacao de
status, download de eventos ou exportacao abre uma nova conexao TCP, autentica,
executa os comandos necessarios e desconecta. Se a central ficar indisponivel,
a pagina continua logada e tenta reconectar no proximo comando; use `Log out`
para limpar o cookie e voltar ao login.

Comandos para a mesma central (`host:porta`) sao serializados no backend para
evitar sessoes concorrentes contra o painel. Testes com painel real indicam que
a central se comporta como dispositivo de uma sessao remota autenticada ativa;
detalhes estao em `docs/protocol/session-behavior.md`.

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

O script gera um relatorio Markdown em `docs/test-runs/` e uma fixture
sanitizada em `docs/fixtures/status/`. A fixture omite IP, senha e cookie, mas
mantem o frame/payload de resposta para testes e documentacao de protocolo.

Para gerar somente a fixture JSON do status:

```sh
AMT_HOST=192.168.1.50 AMT_PASSWORD=123456 go run ./cmd/amt8000-status-capture
```

Para descobrir comandos ainda nao mapeados usando AMT Remoto ou Programador AMT
8000, rode o proxy de captura e aponte o aplicativo oficial para o endereco do
proxy:

```sh
AMT_HOST=192.168.1.50 AMT_PROXY_ADDR=0.0.0.0:19009 go run ./cmd/amt8000-capture-proxy
```

Detalhes em `docs/protocol/capture-proxy.md`.

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

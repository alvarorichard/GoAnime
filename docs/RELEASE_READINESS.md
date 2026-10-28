# Release Readiness — GoAnime v2.0.0

> Documento interno de auditoria, bloqueadores e checklist de release.
>
> - **Última auditoria:** 2026-10-25
> - **Commit auditado:** `b56a7b8` (`dev`)
> - **Alvo:** `v2.0.0`
> - **Status:** **NÃO PRONTO** — 8 bloqueadores P0, nenhum dependente de
>   certificado pago
> - **Orçamento de assinatura:** **R$ 0,00** (caminho gratuito)
> - **Última release publicada:** `v1.8.7`

## Conteúdo

- [1. Veredito](#1-veredito)
- [2. Estado atual](#2-estado-atual)
- [Caminho de custo zero](#caminho-de-custo-zero)
- [3. P0 — bloqueadores](#3-p0--bloqueadores)
- [4. P1 — estabilidade e profissionalismo](#4-p1--estabilidade-e-profissionalismo)
- [5. Ordem de execução](#5-ordem-de-execução)
- [6. Definition of Done](#6-definition-of-done)
- [7. Matriz de teste mínima](#7-matriz-de-teste-mínima)
- [8. Comandos de auditoria](#8-comandos-de-auditoria)
- [9. Runbook de incidente](#9-runbook-de-incidente)
- [10. Referências](#10-referências)

## 1. Veredito

A base de engenharia está acima da média: CI multi-OS, `-race`, cobertura
mínima, `gosec`, `govulncheck`, CodeQL, checksums, changelog, AUR e testes
locais verdes. O que falta para uma release profissional não é feature; é
**confiança na distribuição, compatibilidade e um updater que não se
autossabote**.

Os oito itens da seção [P0](#3-p0--bloqueadores) precisam ser resolvidos antes
de taguear `v2.0.0`. Os itens da seção [P1](#4-p1--estabilidade-e-profissionalismo)
podem entrar em uma `v2.0.0-rc.1` ou logo depois, mas alguns deles — semver,
supply chain e testes E2E — são pré-requisitos para testar a própria release
com segurança.

**Nenhum P0 exige cartão de crédito.** A assinatura que o Windows SmartScreen
e o macOS Gatekeeper reconhecem nativamente é paga, mas ela não é a única
forma de dar autenticidade à release. O plano deste documento usa assinatura
criptográfica do manifesto de atualização, atestados de build, hashes
publicados e canais de distribuição que não dependem de Authenticode nem de
notarização. A seção [Caminho de custo zero](#caminho-de-custo-zero) detalha o
que é possível, o que não é, e como comunicar isso ao usuário sem enganá-lo.

### Resumo dos bloqueadores

| ID | Bloqueador | Status |
| --- | --- | --- |
| P0.1 | Processo de release e governança do repositório | Feito em 2026-10-28, exceto release imutável e pin de Actions por SHA; falta o PR `dev → main` |
| P0.2 | Confiança na distribuição, sem exigir assinatura paga | Aberto |
| P0.3 | Assinatura do manifesto de atualização | Pronto no código e no workflow (job `release` assina após aprovação do environment); falta configurar environment e secrets no GitHub |
| P0.4 | Updater do Windows quebrado no fallback | Resolvido no código; falta E2E em Windows real |
| P0.5 | Segurança local e robustez do updater | Aberto |
| P0.6 | Compatibilidade Linux (glibc) | Aberto |
| P0.7 | Instalador pode corromper o PATH | Resolvido no código; validado no Wine |
| P0.8 | Rollback e atualização atômica | Aberto |

## 2. Estado atual

### 2.1 Código e CI

| Item | Estado | Evidência |
| --- | --- | --- |
| `main` vs `dev` | `main` está **148 commits atrás** de `dev` | `git rev-list --left-right --count origin/main...origin/dev` → `0 148` |
| Tag `v2.0.0` | **não existe** | `git ls-remote --tags origin` vai até `v1.8.7` |
| PR aberto para a v2 | **nenhum** | `gh pr list --state open` → vazio |
| Testes locais | **passam** | `go test -short -count=1 ./...` |
| `gofmt` / `go vet` | **limpos** | `gofmt -l .` vazio; `go vet ./...` sem saída |
| Vulnerabilidades alcançáveis | **0** | `govulncheck ./...` |
| CI | **verde** no `dev` | últimos runs de CI/CodeQL concluídos com sucesso |
| Cobertura mínima | 66% no CI | `.github/workflows/ci.yml` |

### 2.2 Releases e assinatura

| Item | Estado | Evidência |
| --- | --- | --- |
| Release mais recente | `v1.8.7`, publicada em 2026-09-15 | `gh release list` |
| Binário Windows | **sem Authenticode** | Security Directory do PE = 0 em `goanime-windows-amd64.exe` |
| Instalador Windows | **sem Authenticode** | Security Directory do PE = 0 em `GoAnime-Installer-1.8.7.exe` |
| macOS | **sem Developer ID/notarização** | não há `codesign`/`notarytool` nos workflows |
| Checksums | SHA-256 e SHA-512 publicados | `checksums-sha256.txt`, `checksums-sha512.txt` |
| Assinatura dos checksums | **nenhuma** | não há `.minisig`, `.sig` ou bundle cosign |
| SBOM | **nenhum** | nenhum workflow gera SPDX/CycloneDX |
| Attestations | **nenhuma** | API de attestations retorna 404 |
| Release imutável | **não** | `immutable: false` na API |
| `mpv` Windows no build | **`latest` sem hash** | `.github/workflows/release.yml` busca `shinchiro/mpv-winbuild-cmake/releases/latest` e executa o binário no CI |

### 2.3 Configurações do repositório

| Item | Estado | Evidência |
| --- | --- | --- |
| Branch protection | **ativa** (ruleset `protect main`: PR + 5 checks, sem force-push nem deleção) | `gh api repos/alvarorichard/GoAnime/rulesets` |
| Ruleset `procted` | **removido**, substituído por `protect main`, `protect dev` e `protect release tags` | idem |
| Tags protegidas | **sim** (`v*`: só admin cria; mover, atualizar e apagar bloqueados) | ruleset `protect release tags` |
| Secret scanning | **ligado** | `security_and_analysis.secret_scanning.status = enabled` |
| Push protection | **ligado** | `secret_scanning_push_protection.status = enabled` |
| Dependabot security updates | **ligado** | `dependabot_security_updates.status = enabled` |
| Environments | **`release`**: reviewer obrigatório, só tags `v*`; guarda a chave minisign e os secrets da AUR | `gh api repos/alvarorichard/GoAnime/environments` |
| Actions permitidas | **todas**, sem exigência de SHA | `allowed_actions: all`, `sha_pinning_required: false` |
| Permissões padrão do workflow | `read` | bom |
| Chave SSH da AUR | secret de repositório | acessível a qualquer workflow; não está em environment protegido |

### 2.4 Compatibilidade observada

| Item | Estado | Evidência |
| --- | --- | --- |
| Linux x86_64 | exige **glibc 2.34** | maior símbolo `GLIBC_*` em `objdump -T goanime-linux-amd64` |
| Linux arm64 | exige **glibc 2.34** | mesmo comando no binário arm64 |
| Windows | PE x86-64, UCRT, Windows 10+ | `objdump -p` |
| macOS | universal x86_64 + arm64 | `file goanime-darwin-universal` |

O build Linux do release usa `ubuntu-latest` com `CGO_ENABLED=1`
(`.github/workflows/release.yml`), então a v2 tende a exigir glibc igual ou
mais nova que a v1.8.7. Debian 11, Ubuntu 20.04 e RHEL 8 não rodam o binário
atual.

## Caminho de custo zero

> **Resumo:** dá para lançar a v2.0.0 sem pagar assinatura. O Windows pode
> mostrar aviso do SmartScreen e o macOS pode bloquear download pelo
> navegador; a release será verificável por minisign/attestations e
> distribuída por Scoop, winget e Homebrew. Nenhum item obrigatório deste
> plano custa dinheiro.

O objetivo desta seção é separar três coisas que costumam ser confundidas:

1. **assinatura reconhecida pelo sistema operacional** — Authenticode no
   Windows, Developer ID + notarização no macOS. É paga ou depende de
   aprovação discricionária de um programa gratuito.
2. **autenticidade criptográfica da atualização** — minisign/cosign com chave
   pública embutida no binário. É gratuita, obrigatória e independe de
   cartão de crédito.
3. **confiança de distribuição** — canais de pacote, hashes, atestados de
   build, release imutável e transparência. É gratuita, mas não substitui a
   assinatura do sistema operacional.

Sem dinheiro, é possível entregar 2 e 3 integralmente e fazer o melhor
esforço em 1. O usuário ainda pode ver aviso do SmartScreen ou do Gatekeeper
em alguns cenários; o compromisso é que ele consiga **verificar** o que
baixou e entender por que o aviso aparece. Não se deve prometer uma
experiência “sem aviso” que o projeto não pode pagar.

### O que é gratuito e resolve de verdade

| Medida | Custo | O que resolve | Limite |
| --- | --- | --- | --- |
| minisign/cosign no `checksums-sha256.txt` | R$ 0 | Autenticidade do update: asset trocado com checksum trocado é rejeitado | Não remove aviso do SO |
| Chave pública embutida no updater | R$ 0 | Falha fechada mesmo se GitHub/release forem comprometidos | Exige disciplina de chave offline |
| GitHub artifact attestations | R$ 0 em repo público | Prova de que o artefato saiu do workflow daquele commit | Não é assinatura Authenticode |
| SBOM SPDX/CycloneDX | R$ 0 | Transparência de dependências e auditoria | Não melhora reputação sozinho |
| Release imutável + tags protegidas | R$ 0 | Impede substituição silenciosa de assets | Depende de configuração correta |
| Environment `release` com aprovação | R$ 0 | Impede publicação sem revisão | Exige disciplina do maintainer |
| Scoop, winget, Homebrew, AUR, `go install` | R$ 0 | Instalação e atualização por canais verificáveis | Não assinam o binário |
| Build Linux antigo/estático + `.deb`/`.rpm` | R$ 0 | Compatibilidade real | Não resolve assinatura |
| `SECURITY.md`, hashes e instruções claras | R$ 0 | Reduz medo e suporte | Não substitui certificado |

### Programas gratuitos de assinatura Windows

Os dois programas abaixo assinam binários Windows sem cobrança, mas **não são
um direito adquirido**: a aprovação é discricionária, exige projeto open
source qualificado e pode levar semanas. Aplique aos dois em paralelo e não
bloqueie a release esperando resposta.

| Programa | O que oferece | Condições práticas | Risco |
| --- | --- | --- | --- |
| [SignPath Foundation](https://signpath.org/) | Certificado Authenticode gratuito para OSS qualificado, via SignPath.io | Licença OSI, projeto já publicado e mantido, funcionalidade documentada, política de assinatura na home page, MFA, papéis de author/reviewer/approver, aprovação manual por release | O certificado é emitido em nome da SignPath Foundation, não do GoAnime; aprovação pode ser negada |
| [OSSign](https://ossign.org/) | Assinatura gratuita para OSS qualificado | Termos e elegibilidade mudam; confirmar condições atuais antes de aplicar | Programa menor; disponibilidade e suporte incertos |

O GoAnime é MIT e já tem releases públicas, então atende à parte objetiva. A
parte subjetiva é a reputação e a natureza do projeto: um aplicativo de
scraping/streaming pode passar por escrutínio e **não há garantia de
aprovação**. Isso não impede a release; apenas mantém o Windows sem
Authenticode.

O Microsoft Store também passou a ter registro gratuito para pessoa física e
empresa em 2026, e a Microsoft assina o MSIX publicado. Na prática, porém,
não é um bom caminho para o GoAnime agora: o app é TUI, faz scraping, baixa
mpv e tem atualização própria; isso pode conflitar com políticas da Store e
com o modelo de pacote. Trate como opção futura, não como plano de release.

### O que continua custando dinheiro

| Item | Custo aproximado | Observação |
| --- | --- | --- |
| Apple Developer Program + notarização | US$ 99/ano | Não há isenção para pessoa física/OSS comum; isenção só para nonprofit, educação ou governo |
| Azure Trusted Signing | ~US$ 10/mês | Alternativa paga à Authenticode; não usar neste plano |
| Certificado OV/EV tradicional | variável, com token/HSM | EV tende a construir reputação no SmartScreen mais rápido, mas é caro |
| Reputação de SmartScreen | US$ 0, mas incerta | Pode melhorar com volume e tempo; não é garantida para binário não assinado |

Não compre certificado antes de tentar SignPath/OSSign. Também não use
certificado autoassinado: ele não é confiável para o usuário, não remove o
aviso e ainda cria uma falsa sensação de segurança.

### Estratégia por plataforma

#### Windows

1. **Caminho principal:** ZIP portátil por usuário, sem admin, com
   `winget` e um bucket Scoop como canais de instalação/atualização.
2. **Instalador:** continua existindo, mas em modo per-user
   (`%LOCALAPPDATA%\Programs\GoAnime`, `PrivilegesRequired=lowest`) e
   explicitamente marcado como não assinado.
3. **SignPath/OSSign:** se aprovados, assinar `goanime.exe` antes do Inno
   Setup e o instalador depois; manter o ZIP portátil como fallback.
4. **SmartScreen:** a página de release e o README explicam o aviso e o
   procedimento “Mais informações → Executar assim mesmo” **somente depois**
   de conferir o hash e a assinatura minisign. Nunca instruir a desativar o
   SmartScreen ou o antivírus.
5. **Updater:** verifica minisign antes de ler o checksum; a assinatura do SO
   não é usada como critério de update.

#### macOS

1. **Caminho principal:** Homebrew tap com fórmula que compila do source (ou
   instala o binário universal) e `go install`. O Homebrew baixa por `curl`,
   que não aplica quarentena, então o fluxo tende a funcionar sem o aviso do
   Gatekeeper.
2. **Caminho secundário:** binário universal via `curl` + verificação de
   SHA-256/minisign. Também não recebe quarentena.
3. **Download pelo navegador:** o README explica que o macOS aplica
   quarentena e mostra `xattr -d com.apple.quarantine` **após** verificar o
   hash. É uma decisão consciente do usuário, não uma correção de assinatura.
4. **Nada de notarização:** não há caminho gratuito para Developer ID +
   notarização. Não anuncie “notarized” e não peça ao usuário para desativar
   o Gatekeeper globalmente.
5. **Futuro:** se o projeto gerar receita, reavaliar o Apple Developer
   Program. Até lá, Homebrew e source são a distribuição honesta.

#### Linux

1. Build em container com glibc antiga ou build estático/musl, conforme
   [P0.6](#p06--compatibilidade-linux).
2. Publicar `.deb`/`.rpm` e manter AUR.
3. Assinar o manifesto de checksums; o gerenciador de pacotes cuida da
   origem, e o minisign cobre o update direto.

### Ordem prática (sem gastar nada)

| Quando | Ação | Resultado |
| --- | --- | --- |
| Dia 1 | Aplicar em SignPath Foundation e OSSign; ativar MFA e definir papéis | Processo de assinatura iniciado sem bloquear o código |
| Dia 1 | Adicionar a “Code signing policy” no README | Requisito de elegibilidade atendido |
| Dia 1–2 | Gerar chave minisign offline, embutir a pública e implementar verificação no updater | Updates autenticados |
| Dia 2–3 | Ativar release imutável, tags protegidas e environment `release` | Assets não podem ser trocados |
| Dia 3–5 | Adicionar attestations e SBOM ao workflow | Proveniência verificável |
| Semana 1–2 | Publicar Scoop/winget/Homebrew e ajustar instalador per-user | Instalação sem admin e sem UAC |
| Semana 2–3 | Rodar a matriz E2E e publicar `v2.0.0-rc.1` | Update real validado |
| Quando aprovar | Assinar o próximo release no SignPath/OSSign | Aviso do Windows reduzido |

A aprovação pode demorar mais que a release. Nesse caso, publique sem
Authenticode, com a documentação de verificação, e assine o release seguinte.
Não atrase a v2 por causa de um certificado.

### Checklist para a aplicação no SignPath

O SignPath exige uma política de assinatura na home page do projeto. Um
modelo a adaptar no `README.md`:

```markdown
## Code signing policy

Free code signing provided by [SignPath.io](https://about.signpath.io),
certificate by [SignPath Foundation](https://signpath.org).

- Committers and reviewers: [members](https://github.com/alvarorichard/GoAnime/graphs/contributors)
- Approvers: [owners](https://github.com/alvarorichard)
- Privacy policy: <descrever as conexões reais do app, sem prometer o que ele
  não faz>
```

Antes de aplicar, confirme:

- [ ] Licença OSI aprovada (MIT) e sem componente proprietário.
- [ ] Projeto já publicado e mantido.
- [ ] Funcionalidade descrita na página de download/README.
- [ ] MFA ativo na conta GitHub e na SignPath.
- [ ] Papéis de author, reviewer e approver definidos.
- [ ] Política de privacidade compatível com o comportamento real do app.
- [ ] Aprovação manual por release aceita como parte do processo.

### Assinatura do manifesto (obrigatória e gratuita)

O `checksums-sha256.txt` é a raiz de confiança do updater. Assine-o com
minisign e publique `checksums-sha256.txt.minisig` no mesmo release.

```bash
# Gerar o par de chaves (fazer offline, com senha; guardar backup seguro)
minisign -G -p goanime.pub -s goanime.key

# Assinar o manifesto depois de gerar os checksums
minisign -Sm checksums-sha256.txt -s goanime.key

# Verificar (é o que o usuário e o CI devem rodar)
minisign -Vm checksums-sha256.txt -p goanime.pub
```

Regras:

- A chave privada **nunca** entra no repositório, em artifact ou em log.
- Se for inevitável usá-la no GitHub Actions, coloque-a em environment
  protegido com aprovação e rotacione; o ideal é assinar offline e anexar o
  `.minisig`.
- A chave pública fica embutida no binário (`internal/updater`) e no README.
- O updater verifica a assinatura **antes** de interpretar o checksum e
  falha fechado se ela estiver ausente ou inválida.
- Tenha um plano de rotação: publique o key ID e aceite duas chaves durante
  a transição.

### Atestados, SBOM e imutabilidade

- `actions/attest-build-provenance` para cada asset publicado; o usuário
  verifica com `gh attestation verify <asset> --repo alvarorichard/GoAnime`.
- Syft (ou equivalente) gerando SPDX e CycloneDX no workflow de release.
- Release imutável e tag `v*` protegida para que o asset não seja substituído
  depois da publicação.
- Job de verificação pós-release que baixa os assets, confere SHA-256,
  minisign e attestation.

### O que não fazer

- Não usar certificado autoassinado como se fosse solução.
- Não confundir assinatura ad-hoc do macOS (`codesign -s -`) com notarização:
  ela não passa no Gatekeeper.
- Não mandar o usuário desativar SmartScreen, antivírus ou Gatekeeper.
- Não chamar checksum de “assinatura” nem dizer que o binário é “signed”
  quando não é.
- Não colocar a chave privada em secret de repositório sem necessidade.
- Não usar `curl | sudo bash` como instalação principal.
- Não esconder que o binário é não assinado; a transparência é parte da
  confiança.
- Não pagar por certificado antes de esgotar SignPath/OSSign e os canais de
  pacote.

### Custo total do plano

| Item | Custo |
| --- | --- |
| minisign, attestations, SBOM, release imutável | R$ 0 |
| SignPath Foundation / OSSign (se aprovados) | R$ 0 |
| Scoop, winget, Homebrew, AUR, `go install` | R$ 0 |
| Instalador per-user e build Linux antigo | R$ 0 |
| Apple Developer Program | fora do escopo; US$ 99/ano |
| Authenticode pago | fora do escopo; só se os programas gratuitos negarem |

**Total para a v2.0.0: R$ 0,00.** O que não é gratuito é a assinatura
reconhecida nativamente pelo Windows e pelo macOS; o plano entrega
autenticidade do update, proveniência e canais de distribuição sem ela.

## 3. P0 — bloqueadores

### P0.1 — Processo de release e governança do repositório

#### P0.1 — Problema

- `main` está 148 commits atrás de `dev`; a v2.0.0 só existe em `dev`.
- Não há tag, PR ou release candidate.
- Qualquer pessoa com write pode empurrar uma tag e publicar.
- A chave SSH da AUR está em secret de repositório, acessível a qualquer
  workflow; não há environment com aprovação.
- Branch protection, tag protection, secret scanning e Dependabot security
  updates estão desligados.

#### P0.1 — Impacto

Um comprometimento de conta ou um push acidental publica uma release (e um
pacote AUR) sem revisão. É o risco mais grave do processo inteiro.

#### P0.1 — Correção

1. Abrir PR `dev → main`, com CI verde e revisão.
2. Proteger `main` e `dev`: exigir PR, checks obrigatórios, sem force-push e
   sem deleção.
3. Proteger tags `v*` e ativar releases imutáveis.
4. Criar environment `release` com reviewer obrigatório e regra de tag `v*`.
5. Mover `AUR_SSH_PRIVATE_KEY`, `AUR_USERNAME` e `AUR_EMAIL` para esse
   environment; rotacionar a chave.
6. Ativar secret scanning, push protection e Dependabot security updates.
7. Restringir Actions a ações verificadas e exigir pin por SHA.

#### P0.1 — Critério de aceite

- `main` só recebe merge via PR com checks obrigatórios.
- Tag `v*` só pode ser criada por maintainer.
- O job de release e o de AUR exigem aprovação do environment.
- Tag `v2.0.0` aponta para um commit da `main`.

### P0.2 — Confiança na distribuição (sem assinatura paga)

#### P0.2 — Problema

- O `.exe` e o instalador do Windows não têm Authenticode.
- O binário macOS não tem Developer ID nem notarização.
- Não existe hoje uma cadeia de verificação pública que o usuário consiga
  seguir para confirmar que baixou o binário certo.
- A página de release e o README não explicam o aviso do SmartScreen nem o
  bloqueio do Gatekeeper.

#### P0.2 — Impacto

- Windows SmartScreen e antivírus assustam ou bloqueiam o usuário.
- macOS Gatekeeper recusa downloads feitos pelo navegador.
- Sem uma explicação honesta, o usuário aprende a ignorar avisos de
  segurança ou desiste do app.
- Sem assinatura do manifesto, o checksum do mesmo release não prova
  autenticidade.

#### P0.2 — Correção

O caminho obrigatório é gratuito e não depende de Authenticode:

1. **Assinar o manifesto** `checksums-sha256.txt` com minisign e embutir a
   chave pública no updater, conforme [P0.3](#p03--assinatura-do-manifesto-de-atualização).
2. **Publicar instruções de verificação** no README e na release: hash,
   minisign, e o significado do aviso de cada plataforma.
3. **Windows principal:** ZIP portátil per-user, Scoop e winget; instalador
   per-user e explicitamente marcado como não assinado.
4. **macOS principal:** Homebrew tap e `go install`; binário via `curl` +
   hash como alternativa. Não prometer notarização.
5. **Tentar assinatura gratuita em paralelo:** aplicar em
   [SignPath Foundation](https://signpath.org/) e
   [OSSign](https://ossign.org/). Se aprovado, assinar `goanime.exe` antes do
   Inno Setup e o instalador depois, com timestamp, e verificar no CI com
   `Get-AuthenticodeSignature`.
6. **Nunca** usar certificado autoassinado, desativar SmartScreen/Gatekeeper
   globalmente ou anunciar “signed” quando só existe checksum.

Se o SignPath/OSSign aprovar, o certificado é emitido em nome da SignPath
Foundation (no caso do SignPath), não do GoAnime, e cada release exige
aprovação manual. A aprovação é discricionária e pode ser negada; a release
não deve ficar bloqueada esperando resposta.

#### P0.2 — Critério de aceite

- Windows: o ZIP portátil instala e atualiza sem admin; Scoop e winget
  instalam a mesma versão; o instalador per-user não corrompe o PATH; README
  e release notes explicam que o binário não é assinado e como verificar.
- Se SignPath/OSSign aprovar: `Get-AuthenticodeSignature` retorna `Valid`
  para o `.exe` e o instalador, com timestamp.
- macOS: Homebrew e `go install` documentados e testados; o caminho `curl`
  com verificação de hash funciona; não há alegação de notarização; o
  procedimento de quarentena é explícito e não pede para desativar o
  Gatekeeper globalmente.
- `checksums-sha256.txt.minisig` verifica em uma máquina limpa; o updater
  rejeita assinatura ausente ou inválida.
- Nenhum artefato autoassinado é publicado.

#### P0.2 — Custo/observação

| Caminho | Custo | Resultado |
| --- | --- | --- |
| minisign + README + canais de pacote | R$ 0 | Update autenticado e distribuição verificável; aviso do SO pode aparecer |
| SignPath Foundation / OSSign | R$ 0, se aprovado | Authenticode real no Windows; sem garantia e sem macOS |
| Apple Developer Program + notarização | US$ 99/ano | Remove o aviso do Gatekeeper; fora do orçamento |
| Azure Trusted Signing / OV / EV | pago | Authenticode alternativo; fora do orçamento |

O que não dá para comprar com R$ 0 é a reputação imediata no SmartScreen e a
notarização da Apple. O plano não finge que isso está resolvido; ele garante
que o usuário consiga verificar a origem e o conteúdo do download.

### P0.3 — Assinatura do manifesto de atualização

#### P0.3 — Problema

O updater baixa `checksums-sha256.txt` do **mesmo release** e compara o hash
do asset. Se a conta GitHub ou a release for comprometida, o atacante troca o
binário e o checksum juntos. O próprio `docs/AUTO_UPDATE.md` reconhece que
isso não protege contra release comprometida.

#### P0.3 — Impacto

Um comprometimento de conta vira execução remota de código em todos os
usuários que rodarem `--update`. Checksum sozinho garante integridade do
download, não autenticidade da origem.

#### P0.3 — Correção

1. Assinar `checksums-sha256.txt` com **minisign** ou **cosign** e publicar
   `checksums-sha256.txt.minisig` (ou bundle) no release.
2. Embutir a chave pública no binário GoAnime.
3. No updater, verificar a assinatura **antes** de ler os checksums e antes de
   extrair/instalar qualquer coisa. Falhar fechado.
4. Manter a chave de assinatura offline ou em environment protegido. Se a
   chave estiver no GitHub Actions, um comprometimento do repositório também
   assina.
5. Ativar release imutável e tag protection para reduzir substituição de
   assets.

Essa é a única assinatura que custa **R$ 0** e depende só do projeto. Ela é
obrigatória mesmo que o Windows e o macOS continuem sem assinatura nativa; é
o que impede um release comprometido de virar execução remota de código.
Atualize também o `docs/AUTO_UPDATE.md`, que hoje descreve apenas o checksum.

#### P0.3 — Critério de aceite

- Um asset alterado, com checksum atualizado pelo atacante, é rejeitado.
- Um manifesto com assinatura inválida ou ausente é rejeitado.
- O teste de updater cobre assinatura válida, inválida e ausente.
- `gh attestation verify` ou `minisign -Vm` funciona nos artefatos publicados.

### P0.4 — Updater do Windows quebrado no fallback

#### P0.4 — Problema

Em `internal/updater/updater.go`:

- `PerformUpdate` agenda `cleanupUpdateFile()` (linha ~155) e
  `os.Remove(tempFile)` (linha ~127).
- Se o executável em uso não puder ser renomeado, `replaceExecutable` cria um
  `.bat` (`createWindowsUpdateScript`, linha ~904).
- O `.bat` espera o processo atual terminar para copiar o novo `.exe`.
- Quando `PerformUpdate` retorna, os `defer` já apagaram o `.exe` extraído e o
  zip. O script tenta copiar um arquivo que não existe mais.
- O backup também é removido no retorno, então não há rollback.

Além disso, o instalador usa `Program Files` com `PrivilegesRequired=admin`
(`build/install.iss:29`). O updater não eleva; sem admin, nem o backup nem o
script conseguem escrever.

#### P0.4 — Impacto

`goanime --update` no Windows não é confiável: pode falhar por permissão ou
“concluir” sem instalar nada. Usuários do instalador ficam presos até baixar
o instalador novo manualmente.

#### P0.4 — Correção

Implementado em `internal/updater/updater.go`:

- Instalação per-user em `%LOCALAPPDATA%\Programs\GoAnime`, com
  `PrivilegesRequired=lowest`. O updater escreve sem UAC.
- O erro de atualização adiada usa o sentinela `errUpdateDeferred`, não
  comparação de string. `PerformUpdate` não tenta rollback nesse caso.
- O novo `.exe` é copiado para um caminho estável ao lado do executável
  (`goanime.exe.new`) antes de o helper ser criado. Os `defer` podem apagar o
  zip e o diretório de extração sem quebrar o helper.
- O helper `.bat` espera o PID, verifica o hash do `.new` com `certutil`,
  garante o backup, copia com 15 tentativas, verifica o hash do destino e
  restaura o backup se a verificação falhar.
- O backup (`goanime.exe.backup`) não é apagado no retorno de `PerformUpdate`.
  O helper grava `goanime.exe.update-pending` somente depois de uma cópia
  verificada. No próximo start, `updater.CleanupAfterUpdate()` confere o hash do
  executável em execução e só então remove backup, `.new`, `.bat` e marcador.
  Se o hash não bater, o backup é mantido e o marcador é descartado.
- O helper apaga o próprio `.bat` em sucesso e em falha.

#### P0.4 — Critério de aceite

- [x] Atualização adiada não é tratada como erro nem sofre rollback.
- [x] Backup sobrevive ao retorno de `PerformUpdate` e só é removido depois
  que o binário novo inicia com o hash esperado.
- [x] Falha de verificação no helper restaura o binário anterior (testado no
  Wine com `certutil` simulado).
- [ ] Atualizar com o app aberto, reiniciar e ver a versão nova em Windows real.
- [ ] Simular falha de cópia em Windows real e confirmar que a versão anterior
  volta a abrir.
- [x] Instalação 1.x em `Program Files` é detectada (HKLM64 no Wine); o
  instalador pede a desinstalação antes de criar a instalação per-user, para
  não manter duas entradas no PATH.

### P0.5 — Segurança local e robustez do updater

#### P0.5 — Problema

- `safeTempFile` (`internal/updater/updater.go:561`) usa nome fixo em
  `/tmp/goanime-update-<asset>`.
- `os.Create` (`:612`) segue symlink. Com `sudo goanime --update`, um usuário
  local pode pré-criar um symlink e sobrescrever um arquivo arbitrário.
- `downloadAssetWithTestFlag` usa `http.Get` (`:591`) sem timeout, sem limite
  de tamanho e sem revalidar redirects.
- O zip extraído usa `io.LimitReader` de 512 MiB, mas não rejeita uma entrada
  maior que o limite: ela é truncada silenciosamente e instalada.
- A cópia para o executável não é atômica; falha no meio pode deixar o
  binário parcial.

#### P0.5 — Correção

- Baixar para `os.MkdirTemp` com `0700`, usando `os.CreateTemp` com `0600`;
  nunca usar nome fixo em diretório compartilhado.
- Abrir com `O_NOFOLLOW` (ou verificar `Lstat` antes) e remover o arquivo
  temporário em caso de erro.
- `http.Client` com `Timeout`, `CheckRedirect` validando cada host e
  `LimitReader` com limite explícito; rejeitar `Content-Length` acima do
  limite.
- Rejeitar zip cujo entry exceda o limite, em vez de truncar.
- Gravar o novo binário em arquivo temporário no mesmo diretório, `fsync`,
  `chmod` e `rename` atômico sobre o destino (Unix). No Windows, usar helper
  com `MoveFileEx`/instalador.
- Adicionar lock para impedir duas execuções concorrentes de `--update`.

#### P0.5 — Critério de aceite

- Teste com symlink malicioso não sobrescreve o alvo.
- Teste com servidor que não responde falha em tempo limitado.
- Teste com asset acima do limite é rejeitado.
- Teste com redirect para host não permitido é rejeitado.
- Teste de interrupção no meio da cópia deixa o binário anterior funcional.

### P0.6 — Compatibilidade Linux

#### P0.6 — Problema

O binário v1.8.7 exige `GLIBC_2.34`. O release é construído em
`ubuntu-latest` com CGO (`release.yml:30` e `:192`). Debian 11, Ubuntu 20.04 e
RHEL 8 não conseguem executar.

#### P0.6 — Impacto

“Funciona para todos” não é verdade: uma parte relevante dos usuários Linux
recebe `GLIBC_2.34 not found` e não consegue nem abrir o app.

#### P0.6 — Correção

1. Construir o Linux em container com a glibc mais antiga suportada (por
   exemplo `debian:11` ou `debian:12`), com Go e gcc instalados.
2. Testar o binário na mesma imagem e em Ubuntu 22.04/24.04, Debian 11/12 e
   Fedora.
3. Publicar `.deb` e `.rpm` (nfpm ou equivalente) com dependências explícitas
   (`mpv`, `ffmpeg` opcional).
4. Documentar a matriz mínima: kernel, glibc, arquitetura e mpv.
5. Opcional: build estático/`musl` para Alpine e ambientes sem glibc.

#### P0.6 — Critério de aceite

- `objdump -T` mostra a maior versão de glibc documentada.
- O binário roda em todas as distros da matriz mínima.
- A página de release diz claramente “Linux mínimo: glibc X”.

### P0.7 — Instalador pode corromper o PATH

#### P0.7 — Problema

`build/install.iss:58` executa:

```bat
setx PATH "%PATH%;{app};{app}\bin"
```

Isso expande o PATH inteiro no momento da instalação. O `setx` tem limite
histórico de ~1024 caracteres e pode truncar ou corromper o PATH do usuário.

#### P0.7 — Impacto

Usuário perde entradas do PATH, comandos deixam de funcionar e o suporte fica
difícil de diagnosticar.

#### P0.7 — Correção

Implementado em `build/install.iss`:

- O Inno Setup **não tem** seção `[Environment]`; ela é do NSIS. A escrita é
  feita em `[Code]`, com `RegQueryStringValue`/`RegWriteExpandStringValue` em
  `HKCU\Environment`, preservando cada entrada existente.
- `ChangesEnvironment=yes` notifica o Explorer para recarregar o ambiente.
- `PrivilegesRequired=lowest` e
  `DefaultDirName={localappdata}\Programs\GoAnime`.
- `setx` foi removido: ele expande e trunca o PATH e pode gravar no ambiente
  errado.
- A desinstalação remove apenas `{app}` e `{app}\bin`, sem tocar nas demais
  entradas. A remoção usa `StringSplit`/`SameText`/`StringJoin`, não um laço
  `Pos`/`Delete`: o laço antigo errava o índice por causa do `;` inicial e
  deixava um `;` sobrando no fim do PATH. Entradas vazias no meio são
  preservadas.
- O instalador detecta uma instalação 1.x per-machine pelo `UninstallString`
  em HKLM. O instalador 1.x usava `ArchitecturesInstallIn64BitMode`, então a
  chave fica na view 64-bit; por isso a checagem inclui `HKLM64` (com guarda
  `IsWin64`), além de `HKLM`/`HKLM32`. A chave inclui as chaves literais em
  volta do GUID (`{GUID}_is1`).
- O aviso de 1.x usa `SuppressibleMsgBox`, não `MsgBox`. Com `MsgBox`, o
  `/SUPPRESSMSGBOXES` não fecha a caixa e uma instalação silenciosa (ou o CI)
  fica esperando um humano para sempre.

#### P0.7 — Critério de aceite

- [x] PATH antes e depois da instalação difere apenas pelas entradas do
  GoAnime.
- [x] Reinstalar não duplica entradas.
- [x] Desinstalar remove apenas as entradas adicionadas e restaura o PATH
  original, sem `;` sobrando.
- [x] Instalação silenciosa em perfil Wine limpo cria
  `%LOCALAPPDATA%\Programs\GoAnime` e grava
  `...\GoAnime;...\GoAnime\bin` em `HKCU\Environment\Path`.
- [x] PATH acima de 1024 caracteres (1252), com `%USERPROFILE%`/`%JAVA_HOME%`,
  PATH vazio, PATH ausente, entrada vazia no meio e `;` final validados no
  Wine; o PATH útil é restaurado exatamente.
- [x] Instalação 1.x per-machine simulada em HKLM64 bloqueia o instalador
  (rc 1), sem criar pasta nem alterar o PATH.
- [ ] Repetir a matriz de PATH e a detecção 1.x em Windows 10/11 real. O job
  `windows-installer-smoke` em `.github/workflows/ci.yml` foi adicionado para
  isso; falta o primeiro run verde no `windows-latest`.

### P0.8 — Rollback e atualização atômica

#### P0.8 — Problema

- O backup é apagado logo após o sucesso.
- Não há health check depois do update.
- A cópia do novo binário não é atômica.
- Se o novo binário não abrir, o usuário fica sem app e sem instrução de
  rollback.

#### P0.8 — Correção

- Manter a versão anterior como `goanime.previous` (ou em diretório de dados)
  até o primeiro start bem-sucedido.
- Adicionar `goanime --rollback`.
- Gravar o novo binário em temporário + `fsync` + `rename` atômico.
- Criar marcador `update-in-progress`; no start, se o marcador existir e o app
  não conseguir iniciar, restaurar a versão anterior.
- No Windows, o helper deve fazer backup, copiar, verificar hash e restaurar
  em caso de erro.
- Apagar o backup só depois do primeiro start OK.

#### P0.8 — Critério de aceite

- Matar o updater no meio da cópia mantém o app antigo funcional.
- Instalar uma versão que falha no start e confirmar rollback automático ou
  `--rollback`.
- O backup só é removido depois de um start bem-sucedido.

## 4. P1 — estabilidade e profissionalismo

### P1.1 — Semver

`isVersionNewer` (`internal/updater/updater.go:394`) compara partes com
`strconv.Atoi`. Isso falha para pré-release e build metadata:

```text
latest=2.0.0 current=2.0.0-rc.1 -> newer=false err=invalid version format in current: 2.0.0-rc.1
latest=2.0.0 current=2.0.0+build.1 -> newer=false err=invalid version format in current: 2.0.0+build.1
```

**Consequência:** se você publicar `v2.0.0-rc.1`, quem estiver no RC não
consegue atualizar para o stable. Corrija antes de qualquer RC.

**Correção:** usar `golang.org/x/mod/semver` (ou equivalente), com testes para
`v`, pré-release, build metadata, `2.0` vs `2.0.0` e canais stable/beta.

### P1.2 — Supply chain do build

- O mpv do Windows é baixado de `releases/latest` e executado no CI sem hash
  (`.github/workflows/release.yml:331` em diante). Pin de versão + SHA-256, ou
  verificação de attestation do upstream.
- As actions são referenciadas por tag (`@v7`, `@v3`), não por SHA, e
  `sha_pinning_required` está `false`.
- `docker://archlinux:latest` não é pinado por digest.
- Não há SBOM, attestation de build nem verificação pós-release.
- Releases são mutáveis.

**Correção:** pin de mpv e actions; `actions/attest-build-provenance`; Syft
para SPDX/CycloneDX; releases imutáveis; job de verificação que baixa os
assets publicados e confere assinatura/checksum.

### P1.3 — Testes E2E da release

O job de release roda apenas `go test -short` no Linux e um `--version` do
binário universal macOS. Ele não testa os outros artefatos, a instalação, o
PATH nem o update.

O CI agora tem o job `windows-installer-smoke`, que baixa o artefato
`goanime-windows-installer`, confere o SHA-256 publicado, exige
`Get-AuthenticodeSignature` = `NotSigned`, instala em silêncio, roda
`goanime --version`, reinstala sem duplicar PATH, desinstala restaurando o
PATH, testa PATH longo com `%USERPROFILE%`/`%JAVA_HOME%`, PATH ausente e
entradas vazias, e simula uma instalação 1.x em HKLM64 para exigir o bloqueio.
Ele ainda não cobre os demais itens abaixo.

Adicionar uma matriz de smoke/E2E:

- [ ] Instalação silenciosa e desinstalação do instalador Windows no CI
  (`windows-installer-smoke` implementado; aguarda o primeiro run verde).
- [ ] Verificação do PATH antes/depois no Windows real (job acima; aguarda o
  primeiro run verde).
- [ ] Verificação de checksum e status Authenticode do instalador no CI
  (job acima; aguarda o primeiro run verde).
- [ ] `goanime --version` e `--help` em cada binário publicado.
- [ ] Update real de `v1.8.7` para `v2.0.0` em Windows, macOS e Linux.
- [ ] Rollback forçado.
- [ ] Verificação de assinatura e checksum dos assets baixados do release.

### P1.4 — Gerenciadores de pacote

Hoje, de pacote binário gerenciado, só a AUR é suportada. O restante é
`curl + sudo` (ou `go install`), e o updater não detecta instalações
gerenciadas.

No caminho de custo zero, esses canais deixam de ser “nice to have”: eles são
a principal forma de instalar sem UAC e sem depender de Authenticode.
Priorize **Scoop + winget** no Windows e **Homebrew** no macOS.

- Detectar AUR/Homebrew/`go install`/Nix e **recusar** `--update`.
- Adicionar Homebrew tap, winget, Scoop e `.deb`/`.rpm`.
- Publicar chaves de repositório e metadados assinados se houver repo
  apt/dnf.
- Documentar “instalado com X → atualize com X”.
- Cada manifesto deve apontar para o asset com SHA-256 e para a release
  imutável; quando o canal compilar do source, documentar a versão do Go e a
  dependência de CGO.
- O instalador Windows deixa de ser o caminho principal; ele continua para
  quem quer mpv incluído, mas o ZIP portátil e o Scoop são o padrão.

### P1.5 — Diagnóstico e suporte

- `goanime --doctor`: versão, OS/arch, glibc/macOS/Windows, caminho e versão
  do mpv, ffmpeg, yt-dlp, paths de config/log, conectividade e últimas linhas
  de log.
- Rotação/retenção de logs; hoje cada execução cria um arquivo novo.
- Crash report opt-in, sem dados sensíveis.
- `SECURITY.md` com versões suportadas, canal de divulgação e prazo de
  resposta.
- Template de issue pedindo o output do `--doctor`.

### P1.6 — Licenças de terceiros

O projeto redistribui:

- **mpv** (GPL/LGPL) no instalador Windows;
- **Windows Terminal** (MIT) embutido no binário;
- **Noto Sans Mono CJK** (SIL OFL) no pacote de fontes.

Não há `THIRD_PARTY_NOTICES` nem oferta de source do mpv. Inclua os textos de
licença, avisos de copyright e a oferta de source no instalador e nos
arquivos de release. Isso é exigência legal e também requisito de entrada em
Homebrew, winget e ambientes corporativos.

### P1.7 — Estabilidade percebida das fontes

As issues #206 e #176 mostram fontes quebrando e usuários sem playback. Um
binário impecável não parece “estável” se a fonte não abre.

- Manter o source-health e publicar “known issues” no release.
- O kill-switch local (`GOANIME_DISABLED_SOURCES`) já existe; falta um caminho
  remoto/assinado para desabilitar uma fonte quebrada sem nova release.
- Não chamar de stable uma versão com regressão conhecida de playback.

## 5. Ordem de execução

| Fase | Trabalho | Saída |
| --- | --- | --- |
| 0. Freeze | Congelar `dev`, rodar testes, registrar commit | Baseline auditada |
| 1. Governança | PR `dev → main`, proteções, environment `release`, secrets | Repo pronto para taguear |
| 2. Updater | P0.4, P0.5, P0.8 e P1.1 | Update confiável e com rollback |
| 3. Confiança | P0.2, P0.3 e P1.2 | Update autenticado, proveniência e, se aprovado, Authenticode |
| 4. Compatibilidade | P0.6, P0.7 e P1.4 | Linux antigo, PATH e pacotes |
| 5. RC | P1.3, matriz E2E, `v2.0.0-rc.1` | Update real validado |
| 6. Stable | `v2.0.0`, monitoramento e post-mortem | Release pública |

Estimativa grosseira: **1 a 2 semanas** de trabalho técnico. A aprovação do
SignPath/OSSign pode levar mais e **não bloqueia a release**; publique com o
caminho de custo zero e assine o release seguinte se a aprovação chegar. O
caminho crítico é código, compatibilidade e testes — não certificado.

## 6. Definition of Done

### Governança

- [ ] `main` protegida com PR e checks obrigatórios.
- [ ] `dev` protegida.
- [ ] Tags `v*` protegidas.
- [ ] Environment `release` com reviewer obrigatório.
- [ ] Secrets da AUR somente no environment; chave rotacionada.
- [ ] Secret scanning, push protection e Dependabot security updates ativos.

### Build, assinatura e proveniência

- [ ] SignPath Foundation e OSSign aplicados; decisão registrada.
- [ ] Se aprovado: `.exe` e instalador Windows com Authenticode e timestamp;
      `Get-AuthenticodeSignature` retorna `Valid`.
- [ ] Se não aprovado: ZIP portátil, Scoop e winget são o caminho principal;
      README e release notes explicam que o binário não é assinado e como
      verificar hash/minisign.
- [ ] macOS sem alegação de notarização; Homebrew e `go install` testados;
      instruções de quarentena explícitas.
- [ ] `checksums-sha256.txt` assinado com minisign/cosign.
- [ ] Chave pública embutida no updater; falha fechada.
- [ ] Release imutável.
- [ ] SBOM e attestations publicados.
- [ ] Nenhum artefato autoassinado ou instrução para desativar
      SmartScreen/Gatekeeper.

### Updater

- [ ] Semver correto para RC, stable e build metadata.
- [ ] Temp file privado e sem symlink.
- [ ] Timeout, limite de tamanho e redirects validados.
- [ ] Backup mantido até o primeiro start OK.
- [ ] `--rollback` funcional.
- [ ] Windows: portátil, Scoop e winget atualizam sem admin; o instalador
      per-user também atualiza sem intervenção manual.
- [ ] Teste E2E de update em Windows 10/11, macOS Intel/ARM e Linux.

### Compatibilidade e pacotes

- [ ] Linux roda na glibc mínima documentada.
- [ ] `.deb`/`.rpm` publicados ou matriz mínima explicitamente documentada.
- [ ] Instalação silenciosa no Windows não altera o PATH além do esperado.
- [ ] Desinstalação remove apenas o que instalou.
- [ ] AUR, Homebrew, winget, Scoop e `go install` documentados.
- [ ] `--update` recusa instalações gerenciadas.

### Documentação e suporte

- [ ] `SECURITY.md`.
- [ ] Matriz de plataformas suportadas.
- [ ] `THIRD_PARTY_NOTICES` e licenças no release.
- [ ] `--doctor`.
- [ ] Release notes com “known issues” e instruções de rollback.
- [ ] Template de issue atualizado.

## 7. Matriz de teste mínima

| Plataforma | Versões | Instalação | Update | Rollback |
| --- | --- | --- | --- | --- |
| Windows | 10 22H2, 11 23H2/24H2 | portátil, Scoop e winget; instalador opcional | sim | sim |
| macOS | 13, 14, 15 | Homebrew, `go install` e binário universal | sim | sim |
| Ubuntu | 22.04, 24.04 | binário e `.deb` | sim | sim |
| Debian | 11, 12 | binário e `.deb` | sim | sim |
| Fedora | atual e anterior | binário e `.rpm` | sim | sim |
| Arch | atual | AUR | via AUR | via AUR |

## 8. Comandos de auditoria

```bash
# Estado do código
git rev-list --left-right --count origin/main...origin/dev
git ls-remote --tags origin
gh pr list --state open

# Qualidade local
gofmt -l .
go vet ./...
go test -short -race -count=1 ./...
govulncheck ./...
gosec -exclude-generated ./...

# Release atual
gh release list
gh release view v1.8.7 --json assets
gh api repos/alvarorichard/GoAnime/releases/latest --jq '{tag_name,immutable,prerelease}'

# Proteções e segurança do repo
gh api repos/alvarorichard/GoAnime/branches/main/protection
gh api repos/alvarorichard/GoAnime/rulesets
gh api repos/alvarorichard/GoAnime/environments
gh api repos/alvarorichard/GoAnime/actions/permissions/workflow

# Compatibilidade Linux
objdump -T goanime-linux-amd64 | grep -o 'GLIBC_[0-9.]*' | sort -Vu | tail

# Confiança da release (custo zero)
minisign -Vm checksums-sha256.txt -p goanime.pub
gh attestation verify goanime-linux-amd64.tar.gz --repo alvarorichard/GoAnime

# Assinatura Windows, se SignPath/OSSign aprovar (PE)
objdump -x GoAnime-Installer-2.0.0.exe | grep -i "Security Directory"
powershell -c "Get-AuthenticodeSignature .\GoAnime-Installer-2.0.0.exe | Format-List"
```

## 9. Runbook de incidente

Se uma release ruim for publicada:

1. **Parar a distribuição:** marcar a release como pré-release ou definir a
   release anterior como “latest” no GitHub.
2. **Ativar kill-switch** (se existir) para impedir novos `--update` para a
   versão ruim.
3. **Publicar hotfix** imediatamente; não tentar “consertar” asset de release
   imutável.
4. **Comunicar:** issue fixada, release notes e Discord/Reddit, com instruções
   de rollback ou reinstalação.
5. **Preservar evidências:** não apagar assets, logs ou attestations.
6. **Post-mortem:** causa raiz, detecção, tempo de resposta e ação preventiva.

## 10. Referências

- [Atualização automática](AUTO_UPDATE.md)
- [Guia de desenvolvimento](Development.md)
- [Opções de build](BUILD_OPTIONS.md)
- [Updater](../internal/updater/updater.go)
- [Workflow de release](../.github/workflows/release.yml)
- [Workflow de CI](../.github/workflows/ci.yml)
- [Instalador Windows](../build/install.iss)
- [SignPath Foundation](https://signpath.org/) — assinatura gratuita para OSS
- [OSSign](https://ossign.org/) — assinatura gratuita para OSS
- [Opções de assinatura da Microsoft](https://learn.microsoft.com/en-us/windows/apps/package-and-deploy/code-signing-options)
- [GitHub artifact attestations](https://docs.github.com/en/actions/security-for-github-actions/using-artifact-attestations)
- [minisign](https://jedisct1.github.io/minisign/)
- [Scoop](https://scoop.sh/) e [winget](https://learn.microsoft.com/en-us/windows/package-manager/)
- [Homebrew](https://brew.sh/)
- `SECURITY.md` *(criar na raiz do repositório)*

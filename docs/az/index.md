> **Languages:** [English](../en/README.md) · [Русский](../ru/README.md) · [Türkçe](../tr/README.md) · [Azərbaycanca](../az/README.md)

# Telepath — Sənədləşmə

Telepath — tapşırıqları Telegram-dan öz maşınınızda işləyən kodlaşdırma agentinə göndərməyə imkan verən, öz serverinizdə yerləşdirilən körpüdür. Agentin irəliləyişini real vaxtda geri ötürür: nə ilə məşğul olduğunu, hansı faylları redaktə etdiyini, hansı əmrləri icra etdiyini, təsdiq sorğularını və tam git diff-i ilə yekun xülasəni. O, **agent-agnostikdir** (Claude Code, Gemini CLI, OpenAI Codex CLI, Aider vahid interfeys arxasında) və hər tapşırığa öz **git branch**-ini verir, üstəlik istəyə görə Telegram-dan canlı **təsdiqlər** mümkündür.

## Mündəricat

| Səhifə | Nələr var |
|--------|-----------|
| [Başlanğıc](getting-started.md) | Tələblər, hər OS üçün quraşdırma, bot yaratmaq, ilk işə salma, ilk tapşırıq |
| [Konfiqurasiya](configuration.md) | Hər mühit dəyişəni və CLI bayrağı, standart dəyərləri və prioritetləri ilə |
| [Telegram əmrləri](telegram-commands.md) | Hər əmr və inline düymə, nümunələrlə |
| [Agentlər](agents.md) | Çoxagentli model, imkanlar cədvəli və dəqiq CLI çağırışları |
| [İcazələr](permissions.md) | `auto` / `allowlist` / `interactive` və canlı təsdiqin necə işlədiyi |
| [Branch iş axını](branch-workflow.md) | Hər tapşırıq üzrə branch-lar, commit-lər, qoruyucular və təmizlənmə |
| [Arxitektura](architecture.md) | Pipeline, hadisə modeli, paket xəritəsi, MCP server |
| [Docker](docker.md) | Müstəqil və yüngül image-lər, compose, mount-lar |
| [Agent əlavə etmək](adding-an-agent.md) | `Agent` interfeysini implement edib yeni adapter qeydiyyatdan keçirmək |
| [Təhlükəsizlik](security.md) | Allowlist, agent autentifikasiyası, `auto` ehtiyatlılığı, git zəmanəti |
| [Problemlərin həlli](troubleshooting.md) | Ümumi xətalar və onların həlli |
| [Tez-tez verilən suallar](faq.md) | Ümumi suallara qısa cavablar |

## Yadda saxlanılmalı bir şey

**Telepath heç vaxt push etmir, heç vaxt merge etmir və heç vaxt `main`/`master`-ə və ya remote-a toxunmur.** Bir tapşırığın edə biləcəyi ən pis şey, sizin merge edib özünüz push edəcəyiniz, yoxlanıla bilən yerli branch qoymaqdır.

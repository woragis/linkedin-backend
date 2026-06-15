# Experimento A/B — feed cronológico vs ranqueado

## Hipótese

O feed pré-computado (`ranked`) aumenta a taxa de engajamento (visualizações + reações) em relação ao feed cronológico (`chronological`).

## Desenho

- **Experimento:** `feed_ranking_v1` (tabela `ab_experiments`)
- **Variantes:** `chronological` (controle) e `ranked` (tratamento)
- **Atribuição:** determinística por hash do `user_id` (persistida em `ab_assignments`)
- **Métrica primária:** `engagement_rate` = eventos `post_viewed` + `post_liked` por usuário (janela 7 dias)

## Análise

O worker `analytics_rollup` calcula diariamente:

- taxa de engajamento por variante
- intervalo de confiança 95% (aproximação normal)

Resultados em `ab_experiment_results`.

## Interpretação

Compare `metric_value` e intervalos `[ci_lower, ci_upper]` entre variantes. Sobreposição forte dos ICs indica ausência de evidência estatística de diferença ao nível 5%.

## Referências

- Modelo de ranking: `user_feed_scores` (recência + engajamento + PageRank do autor)
- Infraestrutura: `docs/WORKERS.md`, migration `000004_experiments.sql`
